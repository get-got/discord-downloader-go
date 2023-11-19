package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	twitterscraper "github.com/n0madic/twitter-scraper"
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

	//#region Critical to functionality
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

	//#region Connections
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

	//#region Misc Startup Output - Github Update Notification, Version, Discord Invite

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

	//#region MAIN STARTUP COMPLETE, BOT IS FUNCTIONAL

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
		sourceConfig := getSource(&discordgo.Message{ChannelID: channel}, nil)
		if sourceConfig.AutoHistory != nil {
			if *sourceConfig.AutoHistory {
				var autoHistoryChannel arh
				autoHistoryChannel.channel = channel
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
	//#endregion

	//#region BG STARTUP COMPLETE

	if config.Verbose {
		log.Println(lg("Verbose", "Startup", color.HiBlueString, "Background task startup finished, took %s...", uptime()))
	}

	//#endregion

	//#region Download Emojis & Stickers

	go func() {
		time.Sleep(5 * time.Second)

		dataKeysEmoji := func(emoji discordgo.Emoji, serverID string) string {
			ret := config.EmojisFilenameFormat
			keys := [][]string{
				{"{{ID}}", emoji.ID},
				{"{{name}}", emoji.Name},
			}
			for _, key := range keys {
				if strings.Contains(ret, key[0]) {
					ret = strings.ReplaceAll(ret, key[0], key[1])
				}
			}
			return ret
		}

		dataKeysSticker := func(sticker discordgo.Sticker) string {
			ret := config.StickersFilenameFormat
			keys := [][]string{
				{"{{ID}}", sticker.ID},
				{"{{name}}", sticker.Name},
			}
			for _, key := range keys {
				if strings.Contains(ret, key[0]) {
					ret = strings.ReplaceAll(ret, key[0], key[1])
				}
			}
			return ret
		}

		if config.EmojisServers != nil {
			// Handle destination
			destination := "emojis"
			if config.EmojisDestination != nil {
				destination = *config.EmojisDestination
			}
			if err = os.MkdirAll(destination, 0755); err != nil {
				log.Println(lg("Discord", "Emojis", color.HiRedString, "Error while creating destination folder \"%s\": %s", destination, err))
			}
			// Start
			log.Println(lg("Discord", "Emojis", color.MagentaString, "Starting emoji downloads..."))
			for _, serverID := range *config.EmojisServers {
				emojis, err := bot.GuildEmojis(serverID)
				if err != nil {
					log.Println(lg("Discord", "Emojis", color.HiRedString, "Error fetching emojis from %s... %s", serverID, err))
				} else {
					guildName := "UNKNOWN"
					guild, err := bot.Guild(serverID)
					if err == nil {
						guildName = guild.Name
					}
					subfolder := destination + string(os.PathSeparator) + clearPath(guildName)
					if err = os.MkdirAll(subfolder, 0755); err != nil {
						log.Println(lg("Discord", "Emojis", color.HiRedString, "Error while creating subfolder \"%s\": %s", subfolder, err))
					}

					countDownloaded := 0
					countSkipped := 0
					countFailed := 0
					for _, emoji := range emojis {
						url := "https://cdn.discordapp.com/emojis/" + emoji.ID

						status, _ := downloadRequestStruct{
							InputURL:   url,
							Filename:   dataKeysEmoji(*emoji, serverID),
							Path:       subfolder,
							Message:    nil,
							FileTime:   time.Now(),
							HistoryCmd: false,
							EmojiCmd:   true,
							StartTime:  time.Now(),
						}.handleDownload()

						if status.Status == downloadSuccess {
							countDownloaded++
						} else if status.Status == downloadSkippedDuplicate {
							countSkipped++
						} else {
							countFailed++
							log.Println(lg("Discord", "Emojis", color.HiRedString,
								"Failed to download emoji \"%s\": \t[%d - %s] %v",
								url, status.Status, getDownloadStatusString(status.Status), status.Error))
						}
					}

					// Log
					destinationOut := destination
					abs, err := filepath.Abs(destination)
					if err == nil {
						destinationOut = abs
					}
					log.Println(lg("Discord", "Emojis", color.HiMagentaString,
						fmt.Sprintf("%d emojis downloaded, %d skipped, %d failed - Destination: %s",
							countDownloaded, countSkipped, countFailed, destinationOut,
						)))
				}
			}
		}

		if config.StickersServers != nil {
			// Handle destination
			destination := "stickers"
			if config.StickersDestination != nil {
				destination = *config.StickersDestination
			}
			if err = os.MkdirAll(destination, 0755); err != nil {
				log.Println(lg("Discord", "Stickers", color.HiRedString, "Error while creating destination folder \"%s\": %s", destination, err))
			}
			log.Println(lg("Discord", "Stickers", color.MagentaString, "Starting sticker downloads..."))
			for _, serverID := range *config.StickersServers {
				guildName := "UNKNOWN"
				guild, err := bot.Guild(serverID)
				if err != nil {
					log.Println(lg("Discord", "Stickers", color.HiRedString, "Error fetching server %s... %s", serverID, err))
				} else {
					guildName = guild.Name
					subfolder := destination + string(os.PathSeparator) + clearPath(guildName)
					if err = os.MkdirAll(subfolder, 0755); err != nil {
						log.Println(lg("Discord", "Emojis", color.HiRedString, "Error while creating subfolder \"%s\": %s", subfolder, err))
					}

					countDownloaded := 0
					countSkipped := 0
					countFailed := 0
					for _, sticker := range guild.Stickers {
						url := "https://media.discordapp.net/stickers/" + sticker.ID

						status, _ := downloadRequestStruct{
							InputURL:   url,
							Filename:   dataKeysSticker(*sticker),
							Path:       subfolder,
							Message:    nil,
							FileTime:   time.Now(),
							HistoryCmd: false,
							EmojiCmd:   true,
							StartTime:  time.Now(),
						}.handleDownload()

						if status.Status == downloadSuccess {
							countDownloaded++
						} else if status.Status == downloadSkippedDuplicate {
							countSkipped++
						} else {
							countFailed++
							log.Println(lg("Discord", "Stickers", color.HiRedString,
								"Failed to download sticker \"%s\": \t[%d - %s] %v",
								url, status.Status, getDownloadStatusString(status.Status), status.Error))
						}
					}

					// Log
					destinationOut := destination
					abs, err := filepath.Abs(destination)
					if err == nil {
						destinationOut = abs
					}
					log.Println(lg("Discord", "Stickers", color.HiMagentaString,
						fmt.Sprintf("%d stickers downloaded, %d skipped, %d failed - Destination: %s",
							countDownloaded, countSkipped, countFailed, destinationOut,
						)))
				}
			}
		}

	}()

	//#endregion

	// ~~~ RUNNING ~~~

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

func openDatabase() {
	var openT time.Time
	var createT time.Time
	// Database
	log.Println(lg("Database", "", color.YellowString, "Opening database...\t(this can take a bit...)"))
	openT = time.Now()
	myDB, err = db.OpenDB(pathDatabaseBase)
	if err != nil {
		log.Println(lg("Database", "", color.HiRedString, "Unable to open database: %s", err))
		return
	}
	if myDB.Use("Downloads") == nil {
		log.Println(lg("Database", "Setup", color.YellowString, "Creating database, please wait..."))
		createT = time.Now()
		if err := myDB.Create("Downloads"); err != nil {
			log.Println(lg("Database", "Setup", color.HiRedString, "Error while trying to create database: %s", err))
			return
		}
		log.Println(lg("Database", "Setup", color.HiYellowString, "Created new database...\t(took %s)", timeSinceShort(createT)))
		//
		log.Println(lg("Database", "Setup", color.YellowString, "Structuring database, please wait..."))
		createT = time.Now()
		indexColumn := func(col string) {
			if err := myDB.Use("Downloads").Index([]string{col}); err != nil {
				log.Println(lg("Database", "Setup", color.HiRedString, "Unable to create index for %s: %s", col, err))
				return
			}
		}
		indexColumn("URL")
		indexColumn("ChannelID")
		indexColumn("UserID")
		log.Println(lg("Database", "Setup", color.HiYellowString, "Created database structure...\t(took %s)", timeSinceShort(createT)))
	}
	// Cache download tally
	cachedDownloadID = dbDownloadCount()
	log.Println(lg("Database", "", color.HiYellowString, "Database opened, contains %d entries...\t(took %s)", cachedDownloadID, timeSinceShort(openT)))

	// Duplo
	if config.Duplo || sourceHasDuplo {
		log.Println(lg("Duplo", "", color.HiRedString, "!!! Duplo is barely supported and may cause issues, use at your own risk..."))
		duploCatalog = duplo.New()
		if _, err := os.Stat(pathCacheDuplo); err == nil {
			log.Println(lg("Duplo", "", color.YellowString, "Opening duplo image catalog..."))
			openT = time.Now()
			storeFile, err := os.ReadFile(pathCacheDuplo)
			if err != nil {
				log.Println(lg("Duplo", "", color.HiRedString, "Error opening duplo catalog:\t%s", err))
			} else {
				err = duploCatalog.GobDecode(storeFile)
				if err != nil {
					log.Println(lg("Duplo", "", color.HiRedString, "Error decoding duplo catalog:\t%s", err))
				}
				if duploCatalog != nil {
					log.Println(lg("Duplo", "", color.HiYellowString, "Duplo catalog opened (%d)\t(took %s)", duploCatalog.Size(), timeSinceShort(openT)))
				}
			}
		}
	}
}

func botLoad() {
	mainWg.Add(1)
	botLoadAPIs()

	mainWg.Add(1)
	botLoadDiscord()
}

func botLoadAPIs() {
	// Twitter API
	if *config.Credentials.TwitterAuthEnabled {
		go func() {
			twitterScraper = twitterscraper.New()
			if config.Credentials.TwitterUsername != "" &&
				config.Credentials.TwitterPassword != "" {
				log.Println(lg("API", "Twitter", color.MagentaString, "Connecting..."))

				// Proxy
				twitterProxy := func(logIt bool) {
					if config.Credentials.TwitterProxy != "" {
						err := twitterScraper.SetProxy(config.Credentials.TwitterProxy)
						if logIt {
							if err != nil {
								log.Println(lg("API", "Twitter", color.HiRedString, "Error setting proxy: %s", err.Error()))
							} else {
								log.Println(lg("API", "Twitter", color.HiMagentaString, "Proxy set to "+config.Credentials.TwitterProxy))
							}
						}
					}
				}
				twitterProxy(true)

				twitterImport := func() error {
					f, err := os.Open(pathCacheTwitter)
					if err != nil {
						return err
					}
					var cookies []*http.Cookie
					err = json.NewDecoder(f).Decode(&cookies)
					if err != nil {
						return err
					}
					twitterScraper.SetCookies(cookies)
					twitterScraper.IsLoggedIn()
					_, err = twitterScraper.GetProfile("x")
					if err != nil {
						return err
					}
					return nil
				}

				twitterExport := func() error {
					cookies := twitterScraper.GetCookies()
					js, err := json.Marshal(cookies)
					if err != nil {
						return err
					}
					f, err := os.Create(pathCacheTwitter)
					if err != nil {
						return err
					}
					f.Write(js)
					return nil
				}

				// Login Loop
				twitterLoginCount := 0
			do_twitter_login:
				twitterLoginCount++
				if twitterLoginCount > 1 {
					time.Sleep(3 * time.Second)
				}

				if twitterImport() != nil {
					twitterScraper.ClearCookies()
					if err := twitterScraper.Login(config.Credentials.TwitterUsername, config.Credentials.TwitterPassword); err != nil {
						log.Println(lg("API", "Twitter", color.HiRedString, "Login Error: %s", err.Error()))
						if twitterLoginCount <= 3 {
							goto do_twitter_login
						} else {
							log.Println(lg("API", "Twitter", color.HiRedString,
								"Failed to login to Twitter (X), the bot will not fetch this media..."))
						}
					} else {
						twitterConnected = true
						defer twitterExport()
						if twitterScraper.IsLoggedIn() {
							log.Println(lg("API", "Twitter", color.HiMagentaString, fmt.Sprintf("Connected to @%s via new login", config.Credentials.TwitterUsername)))
						} else {
							log.Println(lg("API", "Twitter", color.HiRedString,
								"Scraper login seemed successful but bot is not logged in, Twitter (X) parsing may not work..."))
						}
					}
				} else {
					log.Println(lg("API", "Twitter", color.HiMagentaString,
						"Connected to @%s via cache", config.Credentials.TwitterUsername))
					twitterConnected = true
				}

				if twitterConnected {
					twitterProxy(false)
				}
			} else {
				log.Println(lg("API", "Twitter", color.MagentaString,
					"Twitter (X) login missing, the bot will not fetch this media..."))
			}
		}()
	} else {
		log.Println(lg("API", "Twitter", color.RedString,
			"TWITTER AUTHENTICATION IS DISABLED IN SETTINGS..."))
	}

	// Instagram API
	if *config.Credentials.InstagramAuthEnabled {
		go func() {
			if config.Credentials.InstagramUsername != "" &&
				config.Credentials.InstagramPassword != "" {
				log.Println(lg("API", "Instagram", color.MagentaString, "Connecting..."))

				// Proxy
				instagramProxy := func(logIt bool) {
					if config.Credentials.InstagramProxy != "" {
						insecure := false
						if config.Credentials.InstagramProxyInsecure != nil {
							insecure = *config.Credentials.InstagramProxyInsecure
						}
						forceHTTP2 := false
						if config.Credentials.InstagramProxyForceHTTP2 != nil {
							forceHTTP2 = *config.Credentials.InstagramProxyForceHTTP2
						}
						err := instagramClient.SetProxy(config.Credentials.InstagramProxy, insecure, forceHTTP2)
						if err != nil {
							log.Println(lg("API", "Instagram", color.HiRedString, "Error setting proxy: %s", err.Error()))
						} else {
							log.Println(lg("API", "Instagram", color.HiMagentaString, "Proxy set to "+config.Credentials.InstagramProxy))
						}
					}
				}
				instagramProxy(true)

				// Login Loop
				instagramLoginCount := 0
			do_instagram_login:
				instagramLoginCount++
				if instagramLoginCount > 1 {
					time.Sleep(3 * time.Second)
				}
				if instagramClient, err = goinsta.Import(pathCacheInstagram); err != nil {
					instagramClient = goinsta.New(config.Credentials.InstagramUsername, config.Credentials.InstagramPassword)
					if err := instagramClient.Login(); err != nil {
						log.Println(lg("API", "Instagram", color.HiRedString, "Login Error: %s", err.Error()))
						if instagramLoginCount <= 3 {
							goto do_instagram_login
						} else {
							log.Println(lg("API", "Instagram", color.HiRedString,
								"Failed to login to Instagram, the bot will not fetch this media..."))
						}
					} else {
						log.Println(lg("API", "Instagram", color.HiMagentaString,
							"Connected to @%s via new login", instagramClient.Account.Username))
						instagramConnected = true
						defer instagramClient.Export(pathCacheInstagram)
					}
				} else {
					log.Println(lg("API", "Instagram", color.HiMagentaString,
						"Connected to @%s via cache", instagramClient.Account.Username))
					instagramConnected = true
				}
				if instagramConnected {
					instagramProxy(false)
				}
			} else {
				log.Println(lg("API", "Instagram", color.MagentaString,
					"Instagram login missing, the bot will not fetch this media..."))
			}
		}()
	} else {
		log.Println(lg("API", "Instagram", color.RedString,
			"INSTAGRAM AUTHENTICATION IS DISABLED IN SETTINGS..."))
	}

	mainWg.Done()
}

func botLoadDiscord() {
	var err error

	// Discord Login
	connectBot := func() {
		// Connect Bot
		bot.LogLevel = -1 // to ignore dumb wsapi error
		err = bot.Open()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "web socket already opened") {
			log.Println(lg("Discord", "", color.HiRedString, "Discord login failed:\t%s", err))
			properExit()
		}

		bot.LogLevel = config.DiscordLogLevel // reset
		bot.ShouldReconnectOnError = true
		dur, err := time.ParseDuration(fmt.Sprint(config.DiscordTimeout) + "s")
		if err != nil {
			dur, _ = time.ParseDuration("180s")
		}
		bot.Client.Timeout = dur

		bot.StateEnabled = true
		bot.State.MaxMessageCount = 100000
		bot.State.TrackChannels = true
		bot.State.TrackThreads = true
		bot.State.TrackMembers = true
		bot.State.TrackThreadMembers = true

		botUser, err = bot.User("@me")
		if err != nil {
			botUser = bot.State.User
		}
	}

	discord_login_count := 0
do_discord_login:
	discord_login_count++
	if discord_login_count > 1 {
		time.Sleep(3 * time.Second)
	}

	if config.Credentials.Token != "" && config.Credentials.Token != placeholderToken {
		// Login via Token (Bot or User)
		log.Println(lg("Discord", "", color.GreenString, "Connecting to Discord via Token..."))
		// attempt login without Bot prefix
		bot, err = discordgo.New(config.Credentials.Token)
		connectBot()
		if botUser.Bot { // is bot application, reconnect properly
			//log.Println(lg("Discord", "", color.GreenString, "Reconnecting as bot..."))
			bot, err = discordgo.New("Bot " + config.Credentials.Token)
		}

	} else if (config.Credentials.Email != "" && config.Credentials.Email != placeholderEmail) &&
		(config.Credentials.Password != "" && config.Credentials.Password != placeholderPassword) {
		// Login via Email+Password (User Only obviously)
		log.Println(lg("Discord", "", color.GreenString, "Connecting via Login..."))
		bot, err = discordgo.New(config.Credentials.Email, config.Credentials.Password)
	} else {
		if discord_login_count > 5 {
			log.Println(lg("Discord", "", color.HiRedString, "No valid credentials for Discord..."))
			properExit()
		} else {
			goto do_discord_login
		}
	}
	if err != nil {
		if discord_login_count > 5 {
			log.Println(lg("Discord", "", color.HiRedString, "Error logging in: %s", err))
			properExit()
		} else {
			goto do_discord_login
		}
	}

	connectBot()

	// Fetch Bot's User Info
	botUser, err = bot.User("@me")
	if err != nil {
		botUser = bot.State.User
		if botUser == nil {
			if discord_login_count > 5 {
				log.Println(lg("Discord", "", color.HiRedString, "Error obtaining user details: %s", err))
				properExit()
			} else {
				goto do_discord_login
			}
		}
	} else if botUser == nil {
		if discord_login_count > 5 {
			log.Println(lg("Discord", "", color.HiRedString, "No error encountered obtaining user details, but it's empty..."))
			properExit()
		} else {
			goto do_discord_login
		}
	} else {
		botReady = true
		log.Println(lg("Discord", "", color.HiGreenString, "Logged into %s", getUserIdentifier(*botUser)))
		if botUser.Bot {
			log.Println(lg("Discord", "Info", color.HiMagentaString, "GENUINE DISCORD BOT APPLICATION"))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ This is the safest way to use this bot."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ PRESENCE: Details don't work. Only activity and status."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ VISIBILITY: You can only see servers you have added the bot to, which requires you to be an admin or have an admin invite the bot."))
		} else {
			log.Println(lg("Discord", "Info", color.HiYellowString, "!!! USER ACCOUNT / SELF-BOT !!!"))
			log.Println(lg("Discord", "Info", color.HiMagentaString, "~ WARNING: Discord does NOT ALLOW automated user accounts (aka Self-Bots)."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~~~ By using this bot application with a user account, you potentially risk account termination."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~~~ See the GitHub page for link to Discord's official statement."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~~~ IF YOU WISH TO AVOID THIS, USE A BOT APPLICATION IF POSSIBLE."))
			log.Println(lg("Discord", "Info", color.HiMagentaString, "~ DISCORD API BUGS MAY OCCUR - KNOWN ISSUES:"))
			log.Println(lg("Discord", "Info", color.MagentaString, "~~~ Can't see active threads, only archived threads."))
			log.Println(lg("Discord", "Info", color.HiMagentaString, "~ VISIBILITY: You can download from any channels/servers this account has access to."))
		}
	}
	if bot.State.User != nil { // is selfbot
		selfbot = bot.State.User.Email != ""
	}

	// Event Handlers
	botCommands = handleCommands()
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

	// Start Presence
	timeLastUpdated = time.Now()
	go updateDiscordPresence()

	//(SV) Source Validation
	var invalidAdminChannels []string
	var invalidServers []string
	var invalidCategories []string
	var invalidChannels []string
	var missingPermsAdminChannels [][]string
	var missingPermsCategories [][]string
	var missingPermsChannels [][]string
	if config.Debug {
		log.Println(lg("Debug", "Discord Validation", color.GreenString, "Validating your configured Discord sources..."))
	}
	validateSource := func(checkFunc func(string) error, target string, label string, invalidStack *[]string) bool {
		if err := checkFunc(target); err != nil {
			*invalidStack = append(*invalidStack, target)
			log.Println(lg("Discord", "Validation", color.HiRedString,
				"Bot cannot access %s %s...\t%s", label, target, err))
			return false
		}
		return true
	}
	checkChannelPerm := func(perm int64, permName string, target string, label string, invalidStack *[][]string) {
		if perms, err := bot.State.UserChannelPermissions(botUser.ID, target); err == nil {
			if perms&perm == 0 { // lacks permission
				*invalidStack = append(*invalidStack, []string{target, permName})
				log.Println(lg("Discord", "Validation", color.HiRedString,
					"%s %s - Lacks <%s>...", strings.ToUpper(label), target, permName))
			}
		} else if config.Debug {
			log.Println(lg("Discord", "Validation", color.HiRedString,
				"Encountered error checking Discord permission <%s> in %s %s...\t%s", permName, label, target, err))
		}
	}

	//(SV) Check Admin Channels
	if config.AdminChannels != nil {
		for _, adminChannel := range config.AdminChannels {
			if adminChannel.ChannelIDs != nil {
				for _, subchannel := range *adminChannel.ChannelIDs {
					if validateSource(getChannelErr, subchannel, "admin subchannel", &invalidAdminChannels) {
						checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
							subchannel, "admin subchannel", &missingPermsChannels)
						checkChannelPerm(discordgo.PermissionSendMessages, "PermissionSendMessages",
							subchannel, "admin subchannel", &missingPermsChannels)
						checkChannelPerm(discordgo.PermissionEmbedLinks, "PermissionEmbedLinks",
							subchannel, "admin subchannel", &missingPermsChannels)
					}
				}
			} else {
				if validateSource(getChannelErr, adminChannel.ChannelID, "admin channel", &invalidAdminChannels) {
					checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
						adminChannel.ChannelID, "admin channel", &missingPermsChannels)
					checkChannelPerm(discordgo.PermissionSendMessages, "PermissionSendMessages",
						adminChannel.ChannelID, "admin channel", &missingPermsChannels)
					checkChannelPerm(discordgo.PermissionEmbedLinks, "PermissionEmbedLinks",
						adminChannel.ChannelID, "admin channel", &missingPermsChannels)
				}
			}
		}
	}
	//(SV) Check "servers" config.Servers
	for _, server := range config.Servers {
		if server.ServerIDs != nil {
			for _, subserver := range *server.ServerIDs {
				if validateSource(getServerErr, subserver, "subserver", &invalidServers) {
					// tbd?
				}
			}
		} else {
			if validateSource(getServerErr, server.ServerID, "server", &invalidServers) {
				// tbd?
			}
		}
	}
	//(SV) Check "categories" config.Categories
	for _, category := range config.Categories {
		if category.CategoryIDs != nil {
			for _, subcategory := range *category.CategoryIDs {
				if validateSource(getChannelErr, subcategory, "subcategory", &invalidCategories) {
					checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
						subcategory, "subcategory", &missingPermsChannels)
				}
			}

		} else {
			if validateSource(getChannelErr, category.CategoryID, "category", &invalidCategories) {
				checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
					category.CategoryID, "category", &missingPermsChannels)
			}
		}
	}
	//(SV) Check "channels" config.Channels
	for _, channel := range config.Channels {
		if channel.ChannelIDs != nil {
			for _, subchannel := range *channel.ChannelIDs {
				if validateSource(getChannelErr, subchannel, "subchannel", &invalidChannels) {
					checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
						subchannel, "subchannel", &missingPermsChannels)
					checkChannelPerm(discordgo.PermissionReadMessageHistory, "PermissionReadMessageHistory",
						subchannel, "subchannel", &missingPermsChannels)
					checkChannelPerm(discordgo.PermissionEmbedLinks, "PermissionEmbedLinks",
						subchannel, "subchannel", &missingPermsChannels)
					if channel.ReactWhenDownloaded != nil {
						if *channel.ReactWhenDownloaded {
							checkChannelPerm(discordgo.PermissionAddReactions, "PermissionAddReactions",
								subchannel, "subchannel", &missingPermsChannels)
						}
					}
				}
			}

		} else {
			if validateSource(getChannelErr, channel.ChannelID, "channel", &invalidChannels) {
				checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
					channel.ChannelID, "channel", &missingPermsChannels)
				checkChannelPerm(discordgo.PermissionReadMessageHistory, "PermissionReadMessageHistory",
					channel.ChannelID, "channel", &missingPermsChannels)
				checkChannelPerm(discordgo.PermissionEmbedLinks, "PermissionEmbedLinks",
					channel.ChannelID, "channel", &missingPermsChannels)
				if channel.ReactWhenDownloaded != nil {
					if *channel.ReactWhenDownloaded {
						checkChannelPerm(discordgo.PermissionAddReactions, "PermissionAddReactions",
							channel.ChannelID, "channel", &missingPermsChannels)
					}
				}
			}
		}
	}
	//(SV) NOTE: No validation for users because no way to do that by just user ID from what I've seen.

	//(SV) Output Invalid Sources
	invalidSources := len(invalidAdminChannels) + len(invalidChannels) + len(invalidCategories) + len(invalidServers)
	if invalidSources > 0 {
		log.Println(lg("Discord", "Validation", color.HiRedString,
			"Found %d invalid sources in configuration...", invalidSources))
		logMsg := fmt.Sprintf("Validation found %d invalid sources...\n", invalidSources)
		if len(invalidAdminChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Admin Channels: (%d)** - %s",
				len(invalidAdminChannels), strings.Join(invalidAdminChannels, ", "))
		}
		if len(invalidServers) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Servers: (%d)** - %s",
				len(invalidServers), strings.Join(invalidServers, ", "))
		}
		if len(invalidCategories) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Categories: (%d)** - %s",
				len(invalidCategories), strings.Join(invalidCategories, ", "))
		}
		if len(invalidChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Channels: (%d)** - %s",
				len(invalidChannels), strings.Join(invalidChannels, ", "))
		}
		sendErrorMessage(logMsg)
	} else if config.Debug {
		log.Println(lg("Debug", "Discord Validation", color.HiGreenString, "No source issues detected! Bot has access to all configured sources."))
	}
	//(SV) Output Discord Permission Issues
	missingPermsSources := len(missingPermsAdminChannels) + len(missingPermsCategories) + len(missingPermsChannels)
	if missingPermsSources > 0 {
		log.Println(lg("Discord", "Validation", color.HiRedString,
			"Found %d sources with insufficient Discord permissions...", missingPermsSources))
		// removing this part for now due to multidimensional array change
		/*logMsg := fmt.Sprintf("Validation found %d sources with insufficient Discord permissions...\n", missingPermsSources)
		if len(missingPermsAdminChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Admin Channels: (%d)** - %s",
				len(missingPermsAdminChannels), strings.Join(missingPermsAdminChannels, ", "))
		}
		if len(missingPermsCategories) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Categories: (%d)** - %s",
				len(missingPermsCategories), strings.Join(missingPermsCategories, ", "))
		}
		if len(missingPermsAdminChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Channels: (%d)** - %s",
				len(missingPermsAdminChannels), strings.Join(missingPermsAdminChannels, ", "))
		}
		sendErrorMessage(logMsg)*/
	} else if config.Debug {
		log.Println(lg("Debug", "Discord Validation", color.HiGreenString,
			"No permission issues detected! Bot seems to have all required Discord permissions."))
	}

	mainWg.Done()
}
