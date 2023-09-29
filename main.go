package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	twitterscraper "github.com/n0madic/twitter-scraper"
	"github.com/rivo/duplo"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

var (
	err error

	mainWg sync.WaitGroup

	// Bot
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

	// Gen
	loop                 chan os.Signal
	startTime            time.Time
	timeLastUpdated      time.Time
	timeLastDownload     time.Time
	timeLastMessage      time.Time
	cachedDownloadID     int
	configReloadLastTime time.Time

	// Validation
	invalidAdminChannels []string
	invalidChannels      []string
	invalidCategories    []string
	invalidServers       []string
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
	log.Println(color.HiCyanString(wrapHyphensW(fmt.Sprintf("Welcome to %s v%s", projectName, projectVersion))))
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
		if !isLatestGithubRelease() {
			log.Println(lg("Version", "UPDATE", color.HiCyanString, "***\tUPDATE AVAILABLE\t***"))
			log.Println(lg("Version", "UPDATE", color.CyanString, projectRepoURL+"/releases/latest"))
			log.Println(lg("Version", "UPDATE", color.HiCyanString,
				fmt.Sprintf("You are on v%s, latest is %s", projectVersion, latestGithubRelease),
			))
			log.Println(lg("Version", "UPDATE", color.HiCyanString, "*** See changelog for information ***"))
			log.Println(lg("Version", "UPDATE", color.HiCyanString, "CHECK ALL CHANGELOGS SINCE YOUR LAST UPDATE"))
			log.Println(lg("Version", "UPDATE", color.HiCyanString, "SOME SETTINGS MAY NEED TO BE UPDATED"))
			time.Sleep(5 * time.Second)
		}
	}
	//#endregion

	log.Println(lg("Version", "", color.CyanString, versions(false)))

	log.Println(lg("Info", "", color.HiCyanString, "** Need help? Discord: https://discord.gg/6Z6FJZVaDV **"))
}

func main() {

	//#region Critical to functionality
	loadConfig()
	openDatabase()
	//#endregion

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

	//#region MAIN STARTUP COMPLETE, BOT IS FUNCTIONAL

	if config.Debug {
		log.Println(lg("Main", "", color.YellowString, "Startup finished, took %s...", uptime()))
	}
	log.Println(lg("Main", "", color.HiCyanString,
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
		channelConfig := getSource(&discordgo.Message{ChannelID: channel}, nil)
		if channelConfig.AutoHistory != nil {
			if *channelConfig.AutoHistory {
				var autoHistoryChannel arh
				autoHistoryChannel.channel = channel
				autoHistoryChannel.before = *channelConfig.AutoHistoryBefore
				autoHistoryChannel.since = *channelConfig.AutoHistorySince
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
	constantsStruct := constStruct{}
	constantsStruct.Constants = constants
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

	if config.Debug {
		log.Println(lg("Main", "", color.YellowString, "Background startup tasks finished, took %s...", uptime()))
	}

	//#endregion

	// ~~~ RUNNING ~~~

	//#region ----------- TEST ENV / main

	//#endregion ------------------------

	//#region Exit...
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	sendStatusMessage(sendStatusExit) // not goroutine because we want to wait to send this before logout

	// Log out of twitter if authenticated.
	if twitterScraper.IsLoggedIn() {
		twitterScraper.Logout()
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
	log.Println(lg("Database", "", color.YellowString, "Opening database...\t(this can take a second...)"))
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
		log.Println(lg("Database", "Setup", color.YellowString, "Indexing database, please wait..."))
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
		log.Println(lg("Database", "Setup", color.HiYellowString, "Created new indexes...\t(took %s)", timeSinceShort(createT)))
	}
	// Cache download tally
	cachedDownloadID = dbDownloadCount()
	log.Println(lg("Database", "", color.HiYellowString, "Database opened, contains %d entries...\t(took %s)", cachedDownloadID, timeSinceShort(openT)))

	// Duplo
	if config.Duplo || sourceHasDuplo {
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
			log.Println(lg("Discord", "Info", color.MagentaString, "This is a genuine Discord Bot Application"))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ Presence details & state are disabled, only activity label will work."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ The bot can only see servers you have added it to. Usually you need to be admin or know an admin."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ Nothing is wrong, this is just info :)"))
		} else {
			log.Println(lg("Discord", "Info", color.MagentaString, "This is a User Account (Self-Bot)"))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ Discord does not allow Automated User Accounts (Self-Bots), so by using this bot you potentially risk account termination."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ See GitHub page for link to Discord's official statement."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ If you wish to avoid this, use a Bot Application if possible."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ But since you're using a self-bot, you can download from any channels this account can access."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ Nothing is wrong, this is just info :)"))
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
	if config.Debug {
		log.Println(lg("Discord", "Validation", color.HiYellowString, "Validating configured sources..."))
	}
	//(SV) Check Admin Channels
	if config.AdminChannels != nil {
		for _, adminChannel := range config.AdminChannels {
			if adminChannel.ChannelIDs != nil {
				for _, subchannel := range *adminChannel.ChannelIDs {
					_, err := bot.State.Channel(subchannel)
					if err != nil {
						_, err := bot.Channel(subchannel)
						if err != nil {
							invalidAdminChannels = append(invalidAdminChannels, subchannel)
							log.Println(lg("Discord", "Validation", color.HiRedString,
								"Bot cannot access admin subchannel %s...\t%s", subchannel, err))
						}
					}
				}

			} else {
				_, err := bot.State.Channel(adminChannel.ChannelID)
				if err != nil {
					_, err := bot.Channel(adminChannel.ChannelID)
					if err != nil {
						invalidAdminChannels = append(invalidAdminChannels, adminChannel.ChannelID)
						log.Println(lg("Discord", "Validation", color.HiRedString,
							"Bot cannot access admin channel %s...\t%s", adminChannel.ChannelID, err))
					}
				}
			}
		}
	}
	//(SV) Check Sources
	for _, server := range config.Servers {
		if server.ServerIDs != nil {
			for _, subserver := range *server.ServerIDs {
				_, err := bot.State.Guild(subserver)
				if err != nil {
					_, err := bot.Guild(subserver)
					if err != nil {
						invalidServers = append(invalidServers, subserver)
						log.Println(lg("Discord", "Validation", color.HiRedString,
							"Bot cannot access subserver %s...\t%s", subserver, err))
					}
				}
			}
		} else {
			_, err := bot.State.Guild(server.ServerID)
			if err != nil {
				_, err := bot.Guild(server.ServerID)
				if err != nil {
					invalidServers = append(invalidServers, server.ServerID)
					log.Println(lg("Discord", "Validation", color.HiRedString,
						"Bot cannot access server %s...\t%s", server.ServerID, err))
				}
			}
		}
	}
	for _, category := range config.Categories {
		if category.CategoryIDs != nil {
			for _, subcategory := range *category.CategoryIDs {
				_, err := bot.State.Channel(subcategory)
				if err != nil {
					_, err := bot.Channel(subcategory)
					if err != nil {
						invalidCategories = append(invalidCategories, subcategory)
						log.Println(lg("Discord", "Validation", color.HiRedString,
							"Bot cannot access subcategory %s...\t%s", subcategory, err))
					}
				}
			}

		} else {
			_, err := bot.State.Channel(category.CategoryID)
			if err != nil {
				_, err := bot.Channel(category.CategoryID)
				if err != nil {
					invalidCategories = append(invalidCategories, category.CategoryID)
					log.Println(lg("Discord", "Validation", color.HiRedString,
						"Bot cannot access category %s...\t%s", category.CategoryID, err))
				}
			}
		}
	}
	for _, channel := range config.Channels {
		if channel.ChannelIDs != nil {
			for _, subchannel := range *channel.ChannelIDs {
				_, err := bot.State.Channel(subchannel)
				if err != nil {
					_, err := bot.Channel(subchannel)
					if err != nil {
						invalidChannels = append(invalidChannels, subchannel)
						log.Println(lg("Discord", "Validation", color.HiRedString,
							"Bot cannot access subchannel %s...\t%s", subchannel, err))
					}
				}
			}

		} else {
			_, err := bot.State.Channel(channel.ChannelID)
			if err != nil {
				_, err := bot.Channel(channel.ChannelID)
				if err != nil {
					invalidChannels = append(invalidChannels, channel.ChannelID)
					log.Println(lg("Discord", "Validation", color.HiRedString,
						"Bot cannot access channel %s...\t%s", channel.ChannelID, err))
				}
			}
		}
	}
	//(SV) Output
	invalidSources := len(invalidAdminChannels) + len(invalidChannels) + len(invalidCategories) + len(invalidServers)
	if invalidSources > 0 {
		log.Println(lg("Discord", "Validation", color.HiRedString,
			"Found %d invalid channels/servers in configuration...", invalidSources))
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
		log.Println(lg("Discord", "Validation", color.HiGreenString, "All sources successfully validated!"))
	}

	mainWg.Done()
}
