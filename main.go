package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/rivo/duplo"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

var (
	bot      *discordgo.Session
	user     *discordgo.User
	dgr      *exrouter.Route
	myDB     *db.DB
	imgStore *duplo.Store
	loop     chan os.Signal

	twitterConnected     bool
	googleDriveConnected bool
	googleDriveService   *drive.Service

	startTime        time.Time
	timeLastUpdated  time.Time
	cachedDownloadID int

	configReloadLastTime time.Time
)

func init() {
	loop = make(chan os.Signal, 1)
	startTime = time.Now()
	historyStatus = make(map[string]string)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(color.HiCyanString("Welcome to %s v%s!", projectName, projectVersion))
	log.Println(color.CyanString("> discord-go v%s, Discord API v%s", discordgo.VERSION, discordgo.APIVersion))
}

func main() {
	var err error

	// Config
	log.Println(color.YellowString("Loading settings from \"%s\"...", configFile))
	loadConfig()
	log.Println(color.HiYellowString("Settings loaded, bound to %d channel(s) and %d server(s)", getBoundChannelsCount(), getBoundServersCount()))

	// Github Update Check
	if config.GithubUpdateChecking {
		if !isLatestGithubRelease() {
			log.Println(color.HiCyanString("Update Available!", projectReleaseURL))
			log.Println(color.CyanString("* " + projectReleaseURL))
			log.Println(color.HiCyanString("See changelog for information..."))
			time.Sleep(5 * time.Second)
		}
	}

	//#region Database/Cache Initialization

	// Database
	log.Println(color.YellowString("Opening database..."))
	myDB, err = db.OpenDB(databasePath)
	if err != nil {
		log.Println(color.HiRedString("Unable to open database: %s", err))
		return
	}
	if myDB.Use("Downloads") == nil {
		log.Println(logPrefixSetup, color.YellowString("Creating database, please wait..."))
		if err := myDB.Create("Downloads"); err != nil {
			log.Println(logPrefixSetup, color.HiRedString("Error while trying to create database: %s", err))
			return
		}
		log.Println(logPrefixSetup, color.HiYellowString("Created new database..."))
		log.Println(logPrefixSetup, color.YellowString("Indexing database, please wait..."))
		if err := myDB.Use("Downloads").Index([]string{"URL"}); err != nil {
			log.Println(logPrefixSetup, color.HiRedString("Unable to create database index for URL: %s", err))
			return
		}
		if err := myDB.Use("Downloads").Index([]string{"ChannelID"}); err != nil {
			log.Println(logPrefixSetup, color.HiRedString("Unable to create database index for ChannelID: %s", err))
			return
		}
		if err := myDB.Use("Downloads").Index([]string{"UserID"}); err != nil {
			log.Println(logPrefixSetup, color.HiRedString("Unable to create database index for UserID: %s", err))
			return
		}
		log.Println(logPrefixSetup, color.HiYellowString("Created database indexes..."))
	}
	// Cache download tally
	cachedDownloadID = dbDownloadCount()

	// Image Store
	if config.FilterDuplicateImages {
		imgStore = duplo.New()
		if _, err := os.Stat(imgStorePath); err == nil {
			log.Println(color.YellowString("Opening image filter database..."))
			storeFile, err := ioutil.ReadFile(imgStorePath)
			if err != nil {
				log.Println(color.HiRedString("Error opening imgStore file:\t%s", err))
			} else {
				err = imgStore.GobDecode(storeFile)
				if err != nil {
					log.Println(color.HiRedString("Error decoding imgStore:\t%s", err))
				}
			}
		}
	}

	//#endregion

	//#region Component Initialization

	// Regex
	err = compileRegex()
	if err != nil {
		log.Println(color.HiRedString("Error initializing Regex:\t%s", err))
		return
	}

	// Twitter API
	if config.Credentials.TwitterAccessToken != "" &&
		config.Credentials.TwitterAccessTokenSecret != "" &&
		config.Credentials.TwitterConsumerKey != "" &&
		config.Credentials.TwitterConsumerSecret != "" {

		log.Println(color.MagentaString("Connecting to Twitter API..."))

		twitterClient = anaconda.NewTwitterApiWithCredentials(
			config.Credentials.TwitterAccessToken,
			config.Credentials.TwitterAccessTokenSecret,
			config.Credentials.TwitterConsumerKey,
			config.Credentials.TwitterConsumerSecret,
		)

		twitterSelf, err := twitterClient.GetSelf(url.Values{})
		if err != nil {
			log.Println(color.HiRedString("Error encountered while connecting to Twitter API, the bot won't use the Twitter API. Error: %s", err.Error()))
		} else {
			log.Println(color.HiMagentaString("Connected to Twitter API (@%s)", twitterSelf.ScreenName))
			twitterConnected = true
		}
	} else {
		log.Println(color.MagentaString("Twitter API credentials missing, the bot won't use the Twitter API."))
	}

	// Google Drive Client
	if config.Credentials.GoogleDriveCredentialsJSON != "" {
		log.Println(color.MagentaString("Connecting to Google Drive Client..."))
		ctx := context.Background()
		authJson, err := ioutil.ReadFile(config.Credentials.GoogleDriveCredentialsJSON)
		if err != nil {
			log.Println(color.HiRedString("Error opening Google Credentials JSON:\t%s", err))
		} else {
			googleConfig, err := google.JWTConfigFromJSON(authJson, drive.DriveReadonlyScope)
			if err != nil {
				log.Println(color.HiRedString("Error parsing Google Credentials JSON:\t%s", err))
			} else {
				client := googleConfig.Client(ctx)
				googleDriveService, err = drive.New(client)
				if err != nil {
					log.Println(color.HiRedString("Error setting up Google Drive Client:\t%s", err))
				} else {
					log.Println(color.HiMagentaString("Connected to Google Drive Client"))
					googleDriveConnected = true
				}
			}
		}
	}

	//#endregion

	//#region Discord Initialization
	botLogin()

	// Event Handlers
	dgr = handleCommands()
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

	// Source Validation
	var invalidSources int
	if config.DebugOutput {
		log.Println(logPrefixDebug, color.HiYellowString("Validating configured channels/servers..."))
	}
	if config.AdminChannels != nil {
		for _, adminChannel := range config.AdminChannels {
			_, err := bot.State.Channel(adminChannel.ChannelID)
			if err != nil {
				invalidSources++
				log.Println(color.HiRedString("Bot cannot access admin channel %s...\t%s", adminChannel.ChannelID, err))
			}
		}
	}
	for _, server := range config.Servers {
		if server.ServerIDs != nil {
			for _, subserver := range *server.ServerIDs {
				_, err := bot.State.Guild(subserver)
				if err != nil {
					invalidSources++
					log.Println(color.HiRedString("Bot cannot access subserver %s...\t%s", subserver, err))
				}
			}
		} else {
			_, err := bot.State.Guild(server.ServerID)
			if err != nil {
				invalidSources++
				log.Println(color.HiRedString("Bot cannot access server %s...\t%s", server.ServerID, err))
			}
		}
	}
	for _, channel := range config.Channels {
		if channel.ChannelIDs != nil {
			for _, subchannel := range *channel.ChannelIDs {
				_, err := bot.State.Channel(subchannel)
				if err != nil {
					invalidSources++
					log.Println(color.HiRedString("Bot cannot access subchannel %s...\t%s", subchannel, err))
				}
			}

		} else {
			_, err := bot.State.Channel(channel.ChannelID)
			if err != nil {
				invalidSources++
				log.Println(color.HiRedString("Bot cannot access channel %s...\t%s", channel.ChannelID, err))
			}
		}
	}
	if config.DebugOutput {
		if invalidSources > 0 {
			log.Println(logPrefixDebug, color.HiRedString("Found %d invalid channels/servers in configuration...", invalidSources))
		} else {
			log.Println(logPrefixDebug, color.HiGreenString("All channels/servers successfully validated!"))
		}
	}

	// Start Presence
	timeLastUpdated = time.Now()
	updateDiscordPresence()

	//#endregion

	// Output Done
	if config.DebugOutput {
		log.Println(logPrefixDebug, color.YellowString("Startup finished, took %s...", uptime()))
	}
	log.Println(color.HiCyanString("%s is online! Connected to %d server(s)", projectLabel, len(bot.State.Guilds)))
	log.Println(color.RedString("CTRL+C to exit..."))

	// Log Status
	logStatusMessage(logStatusStartup)

	//#region Background Tasks

	// Tickers
	ticker5m := time.NewTicker(5 * time.Minute)
	ticker15s := time.NewTicker(15 * time.Second)
	go func() {
		for {
			select {
			case <-ticker5m.C:
				// If bot experiences connection interruption the status will go blank until updated by message, this fixes that
				updateDiscordPresence()
			case <-ticker15s.C:
				if time.Since(bot.LastHeartbeatAck).Seconds() > 180 {
					log.Println(color.HiRedString("Discord seems to have lost connection, reconnecting..."))
					log.Println(color.YellowString("Closing connections..."))
					bot.Client.CloseIdleConnections()
					bot.CloseWithCode(1001)
					log.Println(color.RedString("Connections closed!"))
					log.Println(color.GreenString("Logging in..."))
					botLogin()
					log.Println(color.HiGreenString("Reconnected! The bot *should* resume working..."))
					// Log Status
					logStatusMessage(logStatusReconnect)
				}
			}
		}
	}()

	// Compile list of channels to autorun history
	var autorunHistoryChannels []string
	for _, channel := range getAllChannels() {
		channelConfig := getChannelConfig(channel)
		if channelConfig.OverwriteAutorunHistory != nil {
			if *channelConfig.OverwriteAutorunHistory {
				autorunHistoryChannels = append(autorunHistoryChannels, channel)
			}
			continue
		}
		if config.AutorunHistory {
			autorunHistoryChannels = append(autorunHistoryChannels, channel)
		}
	}
	// Process autorun history
	for _, channel := range autorunHistoryChannels {
		if config.AsynchronousHistory {
			go handleHistory(nil, channel, "", "")
		} else {
			handleHistory(nil, channel, "", "")
		}
	}
	if len(autorunHistoryChannels) > 0 {
		log.Println(logPrefixHistory, color.HiYellowString("History Autoruns completed (for %d channel(s))", len(autorunHistoryChannels)))
		log.Println(color.CyanString("Waiting for something else to do..."))
	}

	// Settings Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(color.HiRedString("[Watchers] Error creating NewWatcher:\t%s", err))
	}
	defer watcher.Close()
	err = watcher.Add(configFile)
	if err != nil {
		log.Println(color.HiRedString("[Watchers] Error adding watcher for settings:\t%s", err))
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
					if time.Now().Sub(configReloadLastTime).Milliseconds() > 1 {
						time.Sleep(1 * time.Second)
						log.Println(color.YellowString("Detected changes in \"%s\", reloading settings...", configFile))
						loadConfig()
						log.Println(color.HiYellowString("Settings reloaded, bound to %d channel(s) and %d server(s)", getBoundChannelsCount(), getBoundServersCount()))

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

	// Infinite loop until interrupted
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	logStatusMessage(logStatusExit)

	log.Println(color.GreenString("Logging out of discord..."))
	bot.Close()

	log.Println(color.YellowString("Closing database..."))
	myDB.Close()

	log.Println(color.HiRedString("Exiting... "))
}

func botLogin() {
	var err error

	if config.Credentials.Token != "" && config.Credentials.Token != placeholderToken {
		log.Println(color.GreenString("Connecting to Discord via Token..."))
		if config.Credentials.UserBot {
			bot, err = discordgo.New(config.Credentials.Token)
		} else {
			bot, err = discordgo.New("Bot " + config.Credentials.Token)
		}
	} else if (config.Credentials.Email != "" && config.Credentials.Email != placeholderEmail) &&
		(config.Credentials.Password != "" && config.Credentials.Password != placeholderPassword) {
		log.Println(color.GreenString("Connecting to Discord via Login..."))
		bot, err = discordgo.New(config.Credentials.Email, config.Credentials.Password)
	} else {
		log.Println(color.HiRedString("No valid credentials for Discord..."))
		properExit()
	}
	if err != nil {
		// Newer discordgo throws this error for some reason with Email/Password login
		if err.Error() != "Unable to fetch discord authentication token. <nil>" {
			log.Println(color.HiRedString("Error logging into Discord: %s", err))
			properExit()
		}
	}

	// Connect Bot
	bot.LogLevel = -1 // to ignore dumb wsapi error
	err = bot.Open()
	if err != nil {
		log.Println(color.HiRedString("Discord login failed:\t%s", err))
		properExit()
	}
	bot.LogLevel = config.DiscordLogLevel // reset
	bot.ShouldReconnectOnError = true

	// Fetch Bot's User Info
	user, err = bot.User("@me")
	if err != nil {
		user = bot.State.User
		if user == nil {
			log.Println(color.HiRedString("Error obtaining bot user details: %s", err))
			loop <- syscall.SIGINT
		} else {
			log.Println(color.RedString("Using backup user data, hopefully it works..."))
		}
	} else if user == nil {
		log.Println(color.HiRedString("No error encountered obtaining bot user details, but it's empty..."))
		loop <- syscall.SIGINT
	} else {
		log.Println(color.HiGreenString("Discord logged into %s", getUserIdentifier(*user)))
		if user.Bot {
			log.Println(logPrefixHelper, color.MagentaString("This is a Bot User"))
			log.Println(logPrefixHelper, color.MagentaString("- Status presence details are limited."))
			log.Println(logPrefixHelper, color.MagentaString("- Server access is restricted to servers you have permission to add the bot to."))
		} else {
			log.Println(logPrefixHelper, color.MagentaString("This is a User Account (Self-Bot)"))
			log.Println(logPrefixHelper, color.MagentaString("- Discord does not allow Automated User Accounts (Self-Bots), so by using this bot you potentially risk account termination."))
			log.Println(logPrefixHelper, color.MagentaString("- See GitHub page for link to Discord's official statement."))
			log.Println(logPrefixHelper, color.MagentaString("- If you wish to avoid this, use a Bot account if possible."))
		}
	}
}
