package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/hako/durafmt"
	"github.com/rivo/duplo"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

/* v2.0.0 REWRITE TODO:

* Logging System
*** Implement Log Leveling?
*** Truncate links to exact size?
*** Table/Indentation output?

* Better Message/Embed Send+Error Handling
*** Ensure USER Permission Check Compat

* Audit Settings/Config structure
*** Better Settings Insight / Corrective Suggestions

* Ensure 100% nil checks

* Fix Reddit

* Fix Mastodon

* Fix/Implement Instagram?

* Command: Reboot System

 */

var (
	err error
	// Bot
	bot         *discordgo.Session
	botUser     *discordgo.User
	botCommands *exrouter.Route
	selfbot     bool = false
	botReady    bool = false
	// Storage
	myDB     *db.DB
	imgStore *duplo.Store
	// APIs
	twitterConnected     bool
	googleDriveConnected bool
	googleDriveService   *drive.Service
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
	invalidServers       []string
)

func init() {
	loop = make(chan os.Signal, 1)
	startTime = time.Now()
	historyJobs = make(map[string]historyJob)

	if len(os.Args) > 1 {
		configFileBase = os.Args[1]
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(color.HiCyanString(wrapHyphensW(fmt.Sprintf("Welcome to %s v%s", projectName, projectVersion))))
	log.Println(lg("Version", "", color.CyanString,
		"%s / discord-go v%s (modified) / Discord API v%s", runtime.Version(), discordgo.VERSION, discordgo.APIVersion))

	// Github Update Check
	if config.GithubUpdateChecking {
		if !isLatestGithubRelease() {
			log.Println(lg("Version", "UPDATE", color.HiCyanString, "*** Update Available! ***"))
			log.Println(lg("Version", "UPDATE", color.CyanString, projectReleaseURL))
			log.Println(lg("Version", "UPDATE", color.HiCyanString, "*** See changelog for information ***"))
			log.Println(lg("Version", "UPDATE", color.HiCyanString, "CHECK ALL CHANGELOGS SINCE YOUR LAST UPDATE"))
			log.Println(lg("Version", "UPDATE", color.HiCyanString, "SOME SETTINGS MAY NEED TO BE UPDATED"))
			time.Sleep(5 * time.Second)
		}
	}

	log.Println(lg("Info", "", color.HiCyanString, "** Need help? Discord: https://discord.gg/6Z6FJZVaDV **"))
}

func main() {
	//#region Config
	loadConfig()
	log.Println(lg("Settings", "", color.HiYellowString, "Loaded - bound to %d channel%s and %d server%s",
		getBoundChannelsCount(), pluralS(getBoundChannelsCount()),
		getBoundServersCount(), pluralS(getBoundServersCount()),
	))
	//#endregion

	//#region Database Initialization

	// Database
	log.Println(lg("Database", "", color.YellowString, "Opening database..."))
	myDB, err = db.OpenDB(databasePath)
	if err != nil {
		log.Println(lg("Database", "", color.HiRedString, "Unable to open database: %s", err))
		return
	}
	if myDB.Use("Downloads") == nil {
		log.Println(lg("Database", "Setup", color.YellowString, "Creating database, please wait..."))
		if err := myDB.Create("Downloads"); err != nil {
			log.Println(lg("Database", "Setup", color.HiRedString, "Error while trying to create database: %s", err))
			return
		}
		log.Println(lg("Database", "Setup", color.HiYellowString, "Created new database..."))
		log.Println(lg("Database", "Setup", color.YellowString, "Indexing database, please wait..."))
		if err := myDB.Use("Downloads").Index([]string{"URL"}); err != nil {
			log.Println(lg("Database", "Setup", color.HiRedString, "Unable to create database index for URL: %s", err))
			return
		}
		if err := myDB.Use("Downloads").Index([]string{"ChannelID"}); err != nil {
			log.Println(lg("Database", "Setup", color.HiRedString, "Unable to create database index for ChannelID: %s", err))
			return
		}
		if err := myDB.Use("Downloads").Index([]string{"UserID"}); err != nil {
			log.Println(lg("Database", "Setup", color.HiRedString, "Unable to create database index for UserID: %s", err))
			return
		}
		log.Println(lg("Database", "Setup", color.HiYellowString, "Created new indexes..."))
	}
	// Cache download tally
	cachedDownloadID = dbDownloadCount()
	log.Println(lg("Database", "", color.HiYellowString, "Database opened, contains %d entries...", cachedDownloadID))

	//#region Duplicate Filter Storage
	if config.FilterDuplicateImages {
		imgStore = duplo.New()
		if _, err := os.Stat(imgStorePath); err == nil {
			log.Println(lg("Database", "", color.YellowString, "Opening image filter database..."))
			storeFile, err := ioutil.ReadFile(imgStorePath)
			if err != nil {
				log.Println(lg("Database", "", color.HiRedString, "Error opening imgStore file:\t%s", err))
			} else {
				err = imgStore.GobDecode(storeFile)
				if err != nil {
					log.Println(lg("Database", "", color.HiRedString, "Error decoding imgStore:\t%s", err))
				}
				if imgStore != nil {
					log.Println(lg("Database", "", color.HiYellowString, "filterDuplicateImages database opened %d", imgStore.Size()))
				}
			}
		}
	}
	//#endregion

	//#endregion

	//#region Component Initialization

	// Regex
	if err = compileRegex(); err != nil {
		log.Println(lg("Regex", "", color.HiRedString, "Error initializing:\t%s", err))
		return
	}

	botLoadAPIs()

	//#endregion

	//#region Discord Initialization

	botLoadDiscord()

	//#endregion

	//#region MAIN STARTUP COMPLETE

	if config.DebugOutput {
		log.Println(lg("Main", "", color.YellowString, "Startup finished, took %s...", uptime()))
	}
	log.Println(lg("Main", "", color.HiCyanString,
		wrapHyphensW(fmt.Sprintf("%s v%s is online and connected to %d server%s",
			projectLabel, projectVersion, len(bot.State.Guilds), pluralS(len(bot.State.Guilds))))))
	log.Println(lg("Main", "", color.RedString, "CTRL+C to exit..."))

	// Log Status
	sendStatusMessage(sendStatusStartup)

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
		} else if config.DebugOutput {
			log.Println(lg("Constants", "Debug", color.HiYellowString,
				"%s already cached (processing %s, has %s stored)",
				serverKey, server.ID, constants[serverKey]))
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
				} else if config.DebugOutput {
					log.Println(lg("Constants", "Debug", color.HiYellowString,
						"%s already cached (processing %s/%s, has %s stored)",
						channelKey, server.ID, channel.ID, constants[channelKey]))
				}
			}
		}
	}
	//--- Save constants
	os.MkdirAll(cachePath, 0755)
	if _, err := os.Stat(constantsPath); err == nil {
		err = os.Remove(constantsPath)
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
		err := ioutil.WriteFile(constantsPath, newJson, 0644)
		if err != nil {
			log.Println(lg("Constants", "", color.HiRedString, "Failed to save new constants file...\t%s", err))
		}
	}
	//#endregion

	//#region BG Tasks

	//#region BG Tasks - Tickers
	tickerCheckup := time.NewTicker(time.Duration(config.CheckupRate) * time.Minute)
	tickerPresence := time.NewTicker(time.Duration(config.PresenceRefreshRate) * time.Minute)
	tickerConnection := time.NewTicker(time.Duration(config.ConnectionCheckRate) * time.Minute)
	go func() {
		for {
			select {

			case <-tickerCheckup.C:
				if config.DebugOutput {
					str := fmt.Sprintf("Routine Checkup ... %dms latency, last discord heartbeat %s ago, %s uptime",
						bot.HeartbeatLatency().Milliseconds(),
						durafmt.ParseShort(time.Since(bot.LastHeartbeatSent)), durafmt.ParseShort(time.Since(startTime)))
					if !timeLastMessage.IsZero() {
						str += fmt.Sprintf(", last message %s ago", durafmt.ParseShort(time.Since(timeLastMessage)))
					}
					if !timeLastDownload.IsZero() {
						str += fmt.Sprintf(", last download %s ago", durafmt.ParseShort(time.Since(timeLastDownload)))
					}
					log.Println(lg("Main", "Checkup", color.CyanString, str))
				}

			case <-tickerPresence.C:
				// If bot experiences connection interruption the status will go blank until updated by message, this fixes that
				updateDiscordPresence()

			case <-tickerConnection.C:
				doReconnect := func() {
					log.Println(lg("Discord", "", color.YellowString, "Closing Discord connections..."))
					bot.Client.CloseIdleConnections()
					bot.CloseWithCode(1001)
					bot = nil
					log.Println(lg("Discord", "", color.RedString, "Discord connections closed!"))
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

	//#region BG Tasks - Autorun History
	type arh struct{ channel, before, since string }
	var autorunHistoryChannels []arh
	// Compile list of channels to autorun history
	for _, channel := range getAllRegisteredChannels() {
		channelConfig := getChannelConfig(channel)
		if channelConfig.OverwriteAutorunHistory != nil {
			if *channelConfig.OverwriteAutorunHistory {
				var autorunHistoryChannel arh
				autorunHistoryChannel.channel = channel
				if channelConfig.OverwriteAutorunHistoryBefore != nil {
					autorunHistoryChannel.before = *channelConfig.OverwriteAutorunHistoryBefore
				}
				if channelConfig.OverwriteAutorunHistorySince != nil {
					autorunHistoryChannel.since = *channelConfig.OverwriteAutorunHistorySince
				}
				autorunHistoryChannels = append(autorunHistoryChannels, autorunHistoryChannel)
			}
			continue
		}
		if config.AutorunHistory {
			var autorunHistoryChannel arh
			autorunHistoryChannel.channel = channel
			if config.AutorunHistoryBefore != "" {
				autorunHistoryChannel.before = config.AutorunHistoryBefore
			}
			if config.AutorunHistorySince != "" {
				autorunHistoryChannel.since = config.AutorunHistorySince
			}
			autorunHistoryChannels = append(autorunHistoryChannels, autorunHistoryChannel)
		}
	}
	// Process autorun history
	for _, arh := range autorunHistoryChannels {
		if job, exists := historyJobs[arh.channel]; !exists ||
			(job.Status != historyStatusDownloading && job.Status != historyStatusAbortRequested) {
			job.Status = historyStatusWaiting
			job.OriginChannel = "AUTORUN"
			job.OriginUser = "AUTORUN"
			job.TargetCommandingMessage = nil
			job.TargetChannelID = arh.channel
			job.TargetBefore = dateLocalToUTC(arh.before)
			job.TargetSince = dateLocalToUTC(arh.since)
			job.Updated = time.Now()
			job.Added = time.Now()
			historyJobs[arh.channel] = job
		}
	}
	if len(autorunHistoryChannels) > 0 {
		log.Println(lg("History", "Autorun", color.HiYellowString,
			"History Autoruns completed (for %d channel%s)",
			len(autorunHistoryChannels), pluralS(len(autorunHistoryChannels))))
		log.Println(lg("History", "Autorun", color.CyanString,
			"Waiting for something else to do..."))
	}
	//#endregion

	//#region BG Tasks - History Job Processing
	go func() {
		for {
			if !historyProcessing {
				anyRunning := false
				for _, job := range historyJobs {
					if job.Status == historyStatusDownloading {
						anyRunning = true
					}
				}
				if !anyRunning && len(historyJobs) > 0 {
					var job historyJob
					for _, _job := range historyJobs {
						if _job.Status == historyStatusWaiting {
							job = _job
							break
						}
					}
					// because of modifying the job while iterating historyJobs above
					go handleHistory(job.TargetCommandingMessage, job.TargetChannelID, job.TargetBefore, job.TargetSince)
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()
	//#endregion

	//#region BG Tasks - Settings Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(lg("Settings", "Watcher", color.HiRedString, "Error creating NewWatcher:\t%s", err))
	}
	defer watcher.Close()
	err = watcher.Add(configFile)
	if err != nil {
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
						loadConfig()
						log.Println(lg("Settings", "Watcher", color.HiYellowString,
							"Reloaded - bound to %d channel%s and %d server%s",
							getBoundChannelsCount(), pluralS(getBoundChannelsCount()),
							getBoundServersCount(), pluralS(getBoundServersCount()),
						))

						updateDiscordPresence()
					}
					configReloadLastTime = time.Now()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println(color.HiRedString("[Watchers] Error:\t%s", err))
			}
		}
	}()
	//#endregion

	//#endregion

	// ~~~ RUNNING

	//#region Exit...
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	sendStatusMessage(sendStatusExit)

	log.Println(lg("Discord", "", color.GreenString, "Logging out of discord..."))
	bot.Close()

	log.Println(lg("Database", "", color.YellowString, "Closing database..."))
	myDB.Close()

	log.Println(lg("Main", "", color.HiRedString, "Exiting... "))
	//#endregion
}

func botLoad() {
	botLoadAPIs()
	botLoadDiscord()
}

func botLoadAPIs() {
	// Twitter API
	if config.Credentials.TwitterAccessToken != "" &&
		config.Credentials.TwitterAccessTokenSecret != "" &&
		config.Credentials.TwitterConsumerKey != "" &&
		config.Credentials.TwitterConsumerSecret != "" {

		log.Println(lg("API", "Twitter", color.MagentaString, "Connecting to API..."))

		twitterClient = anaconda.NewTwitterApiWithCredentials(
			config.Credentials.TwitterAccessToken,
			config.Credentials.TwitterAccessTokenSecret,
			config.Credentials.TwitterConsumerKey,
			config.Credentials.TwitterConsumerSecret,
		)

		twitterSelf, err := twitterClient.GetSelf(url.Values{})
		if err != nil {
			log.Println(lg("API", "Twitter", color.HiRedString, "API Login Error: %s", err.Error()))
			log.Println(lg("API", "Twitter", color.MagentaString,
				"Error encountered while connecting to API, the bot won't use the Twitter API."))
		} else {
			log.Println(lg("API", "Twitter", color.HiMagentaString,
				"Connected to API @%s", twitterSelf.ScreenName))
			twitterConnected = true
		}
	} else {
		log.Println(lg("API", "Twitter", color.MagentaString,
			"API credentials missing, the bot won't use the Twitter API."))
	}

	// Google Drive Client
	if config.Credentials.GoogleDriveCredentialsJSON != "" {
		log.Println(lg("API", "Google", color.MagentaString, "Connecting..."))
		ctx := context.Background()
		authJson, err := ioutil.ReadFile(config.Credentials.GoogleDriveCredentialsJSON)
		if err != nil {
			log.Println(lg("API", "Google", color.HiRedString, "Error opening Google Credentials JSON:\t%s", err))
		} else {
			googleConfig, err := google.JWTConfigFromJSON(authJson, drive.DriveReadonlyScope)
			if err != nil {
				log.Println(lg("API", "Google", color.HiRedString, "Error parsing Google Credentials JSON:\t%s", err))
			} else {
				client := googleConfig.Client(ctx)
				googleDriveService, err = drive.New(client)
				if err != nil {
					log.Println(lg("API", "Google", color.HiRedString, "Error setting up client:\t%s", err))
				} else {
					log.Println(lg("API", "Google", color.HiMagentaString, "Connected!"))
					googleDriveConnected = true
				}
			}
		}
	}
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

	if config.Credentials.Token != "" && config.Credentials.Token != placeholderToken {
		// Login via Token (Bot or User)
		log.Println(lg("Discord", "", color.GreenString, "Connecting to Discord via Token..."))
		// attempt login without Bot prefix
		bot, err = discordgo.New(config.Credentials.Token)
		connectBot()
		if botUser.Bot {
			// is bot application, reconnect properly
			log.Println(lg("Discord", "", color.GreenString, "Reconnecting as bot..."))
			bot, err = discordgo.New("Bot " + config.Credentials.Token)
			connectBot()
		}

	} else if (config.Credentials.Email != "" && config.Credentials.Email != placeholderEmail) &&
		(config.Credentials.Password != "" && config.Credentials.Password != placeholderPassword) {
		// Login via Email+Password (User Only obviously)
		log.Println(lg("Discord", "", color.GreenString, "Connecting via Login..."))
		bot, err = discordgo.New(config.Credentials.Email, config.Credentials.Password)
	} else {
		log.Println(lg("Discord", "", color.HiRedString, "No valid credentials for Discord..."))
		properExit()
	}
	if err != nil {
		log.Println(lg("Discord", "", color.HiRedString, "Error logging in: %s", err))
		properExit()
	}
	connectBot()

	// Fetch Bot's User Info
	botUser, err = bot.User("@me")
	if err != nil {
		botUser = bot.State.User
		if botUser == nil {
			log.Println(lg("Discord", "", color.HiRedString, "Error obtaining user details: %s", err))
			loop <- syscall.SIGINT
		}
	} else if botUser == nil {
		log.Println(lg("Discord", "", color.HiRedString, "No error encountered obtaining user details, but it's empty..."))
		loop <- syscall.SIGINT
	} else {
		botReady = true
		log.Println(lg("Discord", "", color.HiGreenString, "Logged into %s", getUserIdentifier(*botUser)))
		if botUser.Bot {
			log.Println(lg("Discord", "", color.MagentaString, "This is a genuine Discord Bot Application"))
			log.Println(lg("Discord", "", color.MagentaString, "- Presence details & state are disabled, only status will work."))
			log.Println(lg("Discord", "", color.MagentaString, "- The bot can only see servers you have added it to."))
		} else {
			log.Println(lg("Discord", "", color.MagentaString, "This is a User Account (Self-Bot)"))
			log.Println(lg("Discord", "", color.MagentaString, "- Discord does not allow Automated User Accounts (Self-Bots), so by using this bot you potentially risk account termination."))
			log.Println(lg("Discord", "", color.MagentaString, "- See GitHub page for link to Discord's official statement."))
			log.Println(lg("Discord", "", color.MagentaString, "- If you wish to avoid this, use a Bot Application if possible."))
		}
	}
	if bot.State.User != nil { // is selfbot
		selfbot = bot.State.User.Email != ""
	}

	// Event Handlers
	botCommands = handleCommands()
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

	// Source Validation
	if config.DebugOutput {
		log.Println(lg("Discord", "Validation", color.HiYellowString, "Validating configured channels/servers..."))
	}
	//-
	if config.AdminChannels != nil {
		for _, adminChannel := range config.AdminChannels {
			if adminChannel.ChannelIDs != nil {
				for _, subchannel := range *adminChannel.ChannelIDs {
					_, err := bot.State.Channel(subchannel)
					if err != nil {
						invalidAdminChannels = append(invalidAdminChannels, subchannel)
						log.Println(lg("Discord", "Validation", color.HiRedString,
							"Bot cannot access admin subchannel %s...\t%s", subchannel, err))
					}
				}

			} else {
				_, err := bot.State.Channel(adminChannel.ChannelID)
				if err != nil {
					invalidAdminChannels = append(invalidAdminChannels, adminChannel.ChannelID)
					log.Println(lg("Discord", "Validation", color.HiRedString,
						"Bot cannot access admin channel %s...\t%s", adminChannel.ChannelID, err))
				}
			}
		}
	}
	//-
	for _, server := range config.Servers {
		if server.ServerIDs != nil {
			for _, subserver := range *server.ServerIDs {
				_, err := bot.State.Guild(subserver)
				if err != nil {
					invalidServers = append(invalidServers, subserver)
					log.Println(lg("Discord", "Validation", color.HiRedString,
						"Bot cannot access subserver %s...\t%s", subserver, err))
				}
			}
		} else {
			_, err := bot.State.Guild(server.ServerID)
			if err != nil {
				invalidServers = append(invalidServers, server.ServerID)
				log.Println(lg("Discord", "Validation", color.HiRedString,
					"Bot cannot access server %s...\t%s", server.ServerID, err))
			}
		}
	}
	for _, channel := range config.Channels {
		if channel.ChannelIDs != nil {
			for _, subchannel := range *channel.ChannelIDs {
				_, err := bot.State.Channel(subchannel)
				if err != nil {
					invalidChannels = append(invalidChannels, subchannel)
					log.Println(lg("Discord", "Validation", color.HiRedString,
						"Bot cannot access subchannel %s...\t%s", subchannel, err))
				}
			}

		} else {
			_, err := bot.State.Channel(channel.ChannelID)
			if err != nil {
				invalidChannels = append(invalidChannels, channel.ChannelID)
				log.Println(lg("Discord", "Validation", color.HiRedString,
					"Bot cannot access channel %s...\t%s", channel.ChannelID, err))
			}
		}
	}
	//-
	invalidSources := len(invalidAdminChannels) + len(invalidChannels) + len(invalidServers)
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
		if len(invalidChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Channels: (%d)** - %s",
				len(invalidChannels), strings.Join(invalidChannels, ", "))
		}
		sendErrorMessage(logMsg)
	} else if config.DebugOutput {
		log.Println(lg("Discord", "Validation", color.HiGreenString, "All channels/servers successfully validated!"))
	}

	// Start Presence
	timeLastUpdated = time.Now()
	updateDiscordPresence()
}
