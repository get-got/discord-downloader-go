package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Davincible/goinsta/v3"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	twitterscraper "github.com/imperatrona/twitter-scraper"
	"github.com/rivo/duplo"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

var (
	// General
	err                error
	loop               chan os.Signal
	mainWg             sync.WaitGroup
	startTime          time.Time
	ddgUpdateAvailable bool = false

	// Downloads
	timeLastUpdated      time.Time
	timeLastDownload     time.Time
	timeLastMessage      time.Time
	cachedDownloadID     int
	configReloadLastTime time.Time

	// Discord
	bot         *discordgo.Session
	botUser     *discordgo.User
	botCommands *exrouter.Route
	selfbot     bool = false
	botReady    bool = false

	// Storage
	myDB         *db.DB
	duploCatalog *duplo.Store

	// APIs
	twitterConnected   bool = false
	twitterScraper     *twitterscraper.Scraper
	instagramConnected bool = false
	instagramClient    *goinsta.Instagram
)

func versions(multiline bool) string {
	if multiline {
		return fmt.Sprintf("%s/%s / %s\ndiscordgo v%s (API v%s)",
			runtime.GOOS, runtime.GOARCH, runtime.Version(), discordgo.VERSION, discordgo.APIVersion)
	} else {
		return fmt.Sprintf("%s/%s / %s / discordgo v%s (API v%s)",
			runtime.GOOS, runtime.GOARCH, runtime.Version(), discordgo.VERSION, discordgo.APIVersion)
	}
}

func botLoad() {
	mainWg.Add(1)
	botLoadAPIs()

	mainWg.Add(1)
	botLoadDiscord()
}

func init() {

	//#region Initialize Logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(color.HiGreenString(wrapHyphensW(fmt.Sprintf("Welcome to %s v%s", projectName, projectVersion))))
	//#endregion

	//#region Initialize Variables
	loop = make(chan os.Signal, 1)

	startTime = time.Now()

	historyJobs = orderedmap.New[string, historyJob]()

	if len(os.Args) > 1 {
		configFileBase = os.Args[1]
	}
	//#endregion

	//#region Github Update Check
	if config.GithubUpdateChecking {
		ddgUpdateAvailable = !isLatestGithubRelease()
	}
	//#endregion
}

func main() {

	//#region <<< CRITICAL INIT >>>

	loadConfig()
	openDatabase()

	//#endregion

	// Output Flag Warnings
	if config.Verbose {
		log.Println(lg("VERBOSE", "", color.HiBlueString, "VERBOSE OUTPUT ENABLED ... just some extra info..."))
	}
	if config.Debug {
		log.Println(lg("DEBUG", "", color.HiYellowString, "DEBUGGING OUTPUT ENABLED ... some troubleshooting data..."))
	}
	if config.DebugExtra {
		log.Println(lg("DEBUG2", "", color.YellowString, "EXTRA DEBUGGING OUTPUT ENABLED ... some in-depth troubleshooting data..."))
	}

	mainWg.Wait() // wait because credentials from config

	//#region <<< CONNECTIONS >>>

	mainWg.Add(2)

	go botLoadAPIs()

	go botLoadDiscord()

	//#endregion

	//#region Initialize Regex
	mainWg.Add(1)
	go func() {
		if err = compileRegex(); err != nil {
			log.Println(lg("Regex", "", color.HiRedString, "Error initializing:\t%s", err))
			return
		}
		mainWg.Done()
	}()
	//#endregion

	//#region [Loops] History Job Processing
	go func() {
		for {
			// Empty Local Cache
			nhistoryJobCnt,
				nhistoryJobCntWaiting,
				nhistoryJobCntRunning,
				nhistoryJobCntAborted,
				nhistoryJobCntErrored,
				nhistoryJobCntCompleted := historyJobs.Len(), 0, 0, 0, 0, 0

			//MARKER: history jobs launch
			// do we even bother?
			if nhistoryJobCnt > 0 {
				// New Cache
				for pair := historyJobs.Oldest(); pair != nil; pair = pair.Next() {
					job := pair.Value
					if job.Status == historyStatusWaiting {
						nhistoryJobCntWaiting++
					} else if job.Status == historyStatusRunning {
						nhistoryJobCntRunning++
					} else if job.Status == historyStatusAbortRequested || job.Status == historyStatusAbortCompleted {
						nhistoryJobCntAborted++
					} else if job.Status == historyStatusErrorReadMessageHistoryPerms || job.Status == historyStatusErrorRequesting {
						nhistoryJobCntErrored++
					} else if job.Status >= historyStatusCompletedNoMoreMessages {
						nhistoryJobCntCompleted++
					}
				}

				// Should Start New Job(s)?
				if nhistoryJobCntRunning < config.HistoryMaxJobs || config.HistoryMaxJobs < 1 {
					openSlots := config.HistoryMaxJobs - nhistoryJobCntRunning
					newJobs := make([]historyJob, openSlots)
					filledSlots := 0
					// Find Jobs
					for pair := historyJobs.Oldest(); pair != nil; pair = pair.Next() {
						if filledSlots == openSlots {
							break
						}
						if pair.Value.Status == historyStatusWaiting {
							newJobs = append(newJobs, pair.Value)
							filledSlots++
						}
					}
					// Start Jobs
					if len(newJobs) > 0 {
						for _, job := range newJobs {
							if job != (historyJob{}) {
								go handleHistory(job.TargetCommandingMessage, job.TargetChannelID, job.TargetBefore, job.TargetSince)
							}
						}
					}
				}
			}

			// Update Cache
			historyJobCnt = nhistoryJobCnt
			historyJobCntWaiting = nhistoryJobCntWaiting
			historyJobCntRunning = nhistoryJobCntRunning
			historyJobCntAborted = nhistoryJobCntAborted
			historyJobCntErrored = nhistoryJobCntErrored
			historyJobCntCompleted = nhistoryJobCntCompleted

			// Wait before checking again
			time.Sleep(time.Duration(config.HistoryManagerRate) * time.Second)
		}
	}()
	//#endregion

	mainWg.Wait() // Once complete, bot is functional

	//#region MISC STARTUP OUTPUT - Github Update Notification, Version, Discord Invite

	log.Println(lg("Version", "", color.MagentaString, versions(false)))

	if config.GithubUpdateChecking {
		if ddgUpdateAvailable {
			log.Println(lg("Version", "UPDATE", color.HiGreenString, "***\tUPDATE AVAILABLE\t***"))
			log.Println(lg("Version", "UPDATE", color.HiGreenString, "DOWNLOAD:\n\n"+projectRepoURL+"/releases/latest\n\n"))
			log.Println(lg("Version", "UPDATE", color.HiGreenString,
				fmt.Sprintf("You are on v%s, latest is %s", projectVersion, latestGithubRelease),
			))
			log.Println(lg("Version", "UPDATE", color.GreenString, "*** See changelogs for information ***"))
			log.Println(lg("Version", "UPDATE", color.GreenString, "Check ALL changelogs since your last update!"))
			log.Println(lg("Version", "UPDATE", color.HiGreenString, "SOME SETTINGS-BREAKING CHANGES MAY HAVE OCCURED!!"))
			time.Sleep(5 * time.Second)
		} else {
			if "v"+projectVersion == latestGithubRelease {
				log.Println(lg("Version", "UPDATE", color.GreenString, "You are on the latest version, v%s", projectVersion))
			} else {
				log.Println(lg("Version", "UPDATE", color.GreenString, "No updates available, you are on v%s, latest is %s", projectVersion, latestGithubRelease))
			}
		}
	}

	log.Println(lg("Info", "", color.HiCyanString, "** NEED HELP? discord-downloader-go Discord Server: https://discord.gg/6Z6FJZVaDV **"))

	//#endregion

	//#region <<< MAIN STARTUP COMPLETE - BOT IS FUNCTIONAL >>>

	if config.Verbose {
		log.Println(lg("Verbose", "Startup", color.HiBlueString, "Startup finished, took %s...", uptime()))
	}
	log.Println(lg("Main", "", color.HiGreenString,
		wrapHyphensW(fmt.Sprintf("%s v%s is online with access to %d server%s",
			projectLabel, projectVersion, len(bot.State.Guilds), pluralS(len(bot.State.Guilds))))))
	log.Println(lg("Main", "", color.RedString, "CTRL+C to exit..."))

	// Log Status
	go sendStatusMessage(sendStatusStartup)

	//#endregion

	//#region Autorun History
	type arh struct{ channel, before, since string }
	var autoHistoryChannels []arh
	// Compile list of channels to autorun history
	for _, channel := range getAllRegisteredChannels() {
		sourceConfig := getSource(&discordgo.Message{ChannelID: channel.ChannelID})
		if sourceConfig.AutoHistory != nil {
			if *sourceConfig.AutoHistory {
				var autoHistoryChannel arh
				autoHistoryChannel.channel = channel.ChannelID
				autoHistoryChannel.before = *sourceConfig.AutoHistoryBefore
				autoHistoryChannel.since = *sourceConfig.AutoHistorySince
				autoHistoryChannels = append(autoHistoryChannels, autoHistoryChannel)
			}
			continue
		}
	}
	// Process auto history
	for _, ah := range autoHistoryChannels {
		//MARKER: history jobs queued from auto
		if job, exists := historyJobs.Get(ah.channel); !exists ||
			(job.Status != historyStatusRunning && job.Status != historyStatusAbortRequested) {
			job.Status = historyStatusWaiting
			job.OriginChannel = "AUTORUN"
			job.OriginUser = "AUTORUN"
			job.TargetCommandingMessage = nil
			job.TargetChannelID = ah.channel
			job.TargetBefore = ah.before
			job.TargetSince = ah.since
			job.Updated = time.Now()
			job.Added = time.Now()
			historyJobs.Set(ah.channel, job)
			//TODO: signals for this and typical history cmd??
		}
	}
	if len(autoHistoryChannels) > 0 {
		log.Println(lg("History", "Autorun", color.HiYellowString,
			"History Autoruns completed (for %d channel%s)",
			len(autoHistoryChannels), pluralS(len(autoHistoryChannels))))
		log.Println(lg("History", "Autorun", color.CyanString,
			"Waiting for something else to do..."))
	}
	//#endregion

	//#region [Loops] Tickers
	tickerCheckup := time.NewTicker(time.Duration(config.CheckupRate) * time.Minute)
	tickerPresence := time.NewTicker(time.Duration(config.PresenceRefreshRate) * time.Minute)
	tickerConnection := time.NewTicker(time.Duration(config.ConnectionCheckRate) * time.Minute)
	go func() {
		for {
			select {

			case <-tickerCheckup.C:
				if config.Debug {
					//MARKER: history jobs polled for waiting count in checkup
					historyJobsWaiting := 0
					if historyJobs.Len() > 0 {
						for jobPair := historyJobs.Oldest(); jobPair != nil; jobPair.Next() {
							job := jobPair.Value
							if job.Status == historyStatusWaiting {
								historyJobsWaiting++
							}
						}
					}
					str := fmt.Sprintf("... %dms latency,\t\tlast discord heartbeat %s ago,\t\t%s uptime",
						bot.HeartbeatLatency().Milliseconds(),
						timeSinceShort(bot.LastHeartbeatSent),
						timeSinceShort(startTime))
					if !timeLastMessage.IsZero() {
						str += fmt.Sprintf(",\tlast message %s ago",
							timeSinceShort(timeLastMessage))
					}
					if !timeLastDownload.IsZero() {
						str += fmt.Sprintf(",\tlast download %s ago",
							timeSinceShort(timeLastDownload))
					}
					if historyJobsWaiting > 0 {
						str += fmt.Sprintf(",\t%d history jobs waiting", historyJobsWaiting)
					}
					log.Println(lg("Checkup", "", color.YellowString, str))
				}

			case <-tickerPresence.C:
				// If bot experiences connection interruption the status will go blank until updated by message, this fixes that
				go updateDiscordPresence()

			case <-tickerConnection.C:
				if config.ConnectionCheck {
					doReconnect := func() {
						log.Println(lg("Discord", "", color.YellowString, "Closing Discord connections..."))
						bot.Client.CloseIdleConnections()
						bot.CloseWithCode(1001)
						bot = nil
						log.Println(lg("Discord", "", color.RedString, "Discord connections closed!"))
						time.Sleep(15 * time.Second)
						if config.ExitOnBadConnection {
							properExit()
						} else {
							log.Println(lg("Discord", "", color.GreenString, "Logging in..."))
							botLoad()
							log.Println(lg("Discord", "", color.HiGreenString,
								"Reconnected! The bot *should* resume working..."))
							// Log Status
							sendStatusMessage(sendStatusReconnect)
						}
					}
					gate, err := bot.Gateway()
					if err != nil || gate == "" {
						log.Println(lg("Discord", "", color.HiYellowString,
							"Bot encountered a gateway error: GATEWAY: %s,\tERR: %s", gate, err))
						doReconnect()
					} else if time.Since(bot.LastHeartbeatAck).Seconds() > 4*60 {
						log.Println(lg("Discord", "", color.HiYellowString,
							"Bot has not received a heartbeat from Discord in 4 minutes..."))
						doReconnect()
					}
				}
			}
		}
	}()
	//#endregion

	//#region [Loop] Settings Watcher
	if config.WatchSettings {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Println(lg("Settings", "Watcher", color.HiRedString, "Error creating NewWatcher:\t%s", err))
		}
		defer watcher.Close()
		if err = watcher.Add(configFile); err != nil {
			log.Println(lg("Settings", "Watcher", color.HiRedString, "Error adding watcher for settings:\t%s", err))
		}
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						// It double-fires the event without time check, might depend on OS but this works anyways
						if time.Since(configReloadLastTime).Milliseconds() > 1 {
							time.Sleep(1 * time.Second)
							log.Println(lg("Settings", "Watcher", color.YellowString,
								"Detected changes in \"%s\", reloading...", configFile))
							mainWg.Add(1)
							go loadConfig()
							allString := ""
							if config.All != nil {
								allString = ", ALL ENABLED"
							}
							log.Println(lg("Settings", "Watcher", color.HiYellowString,
								"Reloaded - bound to %d channel%s, %d categories, %d server%s, %d user%s%s",
								getBoundChannelsCount(), pluralS(getBoundChannelsCount()),
								getBoundCategoriesCount(),
								getBoundServersCount(), pluralS(getBoundServersCount()),
								getBoundUsersCount(), pluralS(getBoundUsersCount()), allString,
							))

							go updateDiscordPresence()
							go sendStatusMessage(sendStatusSettings)
							configReloadLastTime = time.Now()
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println(color.HiRedString("[Watchers] Error:\t%s", err))
				}
			}
		}()
	}
	//#endregion

	//#region Database Backup

	if config.BackupDatabaseOnStart {
		if err = backupDatabase(); err != nil {
			log.Println(lg("Database", "Backup", color.HiRedString, "Error backing up database:\t%s", err))
		}
	}

	//#endregion

	//#region Cache Constants

	go func() {
		constants := make(map[string]string)
		//--- Compile constants
		for _, server := range bot.State.Guilds {
			serverKey := fmt.Sprintf("SERVER_%s", stripSymbols(server.Name))
			serverKey = strings.ReplaceAll(serverKey, " ", "_")
			for strings.Contains(serverKey, "__") {
				serverKey = strings.ReplaceAll(serverKey, "__", "_")
			}
			serverKey = strings.ToUpper(serverKey)
			if constants[serverKey] == "" {
				constants[serverKey] = server.ID
			}
			for _, channel := range server.Channels {
				if channel.Type != discordgo.ChannelTypeGuildCategory {
					categoryName := ""
					if channel.ParentID != "" {
						channelParent, err := bot.State.Channel(channel.ParentID)
						if err != nil {
							channelParent, err = bot.Channel(channel.ParentID)
						}
						if err == nil {
							categoryName = channelParent.Name
						}
					}
					channelKey := fmt.Sprintf("CHANNEL_%s_%s_%s",
						stripSymbols(server.Name), stripSymbols(categoryName), stripSymbols(channel.Name))
					channelKey = strings.ReplaceAll(channelKey, " ", "_")
					for strings.Contains(channelKey, "__") {
						channelKey = strings.ReplaceAll(channelKey, "__", "_")
					}
					channelKey = strings.ToUpper(channelKey)
					if constants[channelKey] == "" {
						constants[channelKey] = channel.ID
					}
				}
			}
		}
		//--- Save constants
		if _, err := os.Stat(pathConstants); err == nil {
			err = os.Remove(pathConstants)
			if err != nil {
				log.Println(lg("Constants", "", color.HiRedString, "Encountered error deleting cache file:\t%s", err))
			}
		}
		var constantsStruct map[string]string
		constantsStruct = constants
		newJson, err := json.MarshalIndent(constantsStruct, "", "\t")
		if err != nil {
			log.Println(lg("Constants", "", color.HiRedString, "Failed to format constants...\t%s", err))
		} else {
			err := os.WriteFile(pathConstants, newJson, 0644)
			if err != nil {
				log.Println(lg("Constants", "", color.HiRedString, "Failed to save new constants file...\t%s", err))
			}
		}
	}()

	//#endregion

	//#region Download Emojis & Stickers (after 5s delay)

	go func() {
		time.Sleep(5 * time.Second)

		downloadDiscordEmojis()

		downloadDiscordStickers()

	}()

	//#endregion

	//#region <<< BACKGROUND STARTUP COMPLETE >>>

	if config.Verbose {
		log.Println(lg("Verbose", "Startup", color.HiBlueString, "Background task startup finished, took %s...", uptime()))
	}

	//#endregion

	// <<<<<< RUNNING >>>>>>

	//#region ----------- TEST ENV / main

	//#endregion ------------------------

	//#region Exit...
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	sendStatusMessage(sendStatusExit) // not goroutine because we want to wait to send this before logout

	// Log out of twitter if authenticated.
	if twitterScraper != nil {
		if twitterScraper.IsLoggedIn() {
			twitterScraper.Logout()
		}
	}

	log.Println(lg("Discord", "", color.GreenString, "Logging out of discord..."))
	bot.Close()

	log.Println(lg("Database", "", color.YellowString, "Closing database..."))
	myDB.Close()

	log.Println(lg("Main", "", color.HiRedString, "Exiting... "))
	//#endregion

}
