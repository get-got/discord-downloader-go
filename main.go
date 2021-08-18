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
	"strings"
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

	invalidAdminChannels []string
	invalidChannels      []string
	invalidServers       []string

	configReloadLastTime time.Time
)

func init() {
	loop = make(chan os.Signal, 1)
	startTime = time.Now()
	historyStatus = make(map[string]string)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(color.HiCyanString(wrapHyphensW(fmt.Sprintf("Welcome to %s v%s", projectName, projectVersion))))
	log.Println(logPrefixVersion, color.CyanString("discord-go v%s using Discord API v%s", discordgo.VERSION, discordgo.APIVersion))
}

func main() {
	var err error

	// Config
	loadConfig()
	log.Println(logPrefixSettings, color.HiYellowString("Loaded - bound to %d channel%s and %d server%s",
		getBoundChannelsCount(), pluralS(getBoundChannelsCount()),
		getBoundServersCount(), pluralS(getBoundServersCount()),
	))

	// Github Update Check
	if config.GithubUpdateChecking {
		if !isLatestGithubRelease() {
			log.Println(logPrefixVersion, color.HiCyanString("*** Update Available! ***"))
			log.Println(logPrefixVersion, color.CyanString(projectReleaseURL))
			log.Println(logPrefixVersion, color.HiCyanString("*** See changelog for information ***"))
			time.Sleep(5 * time.Second)
		}
	}

	//#region Database/Cache Initialization

	// Database
	log.Println(logPrefixDatabase, color.YellowString("Opening database..."))
	myDB, err = db.OpenDB(databasePath)
	if err != nil {
		log.Println(logPrefixDatabase, color.HiRedString("Unable to open database: %s", err))
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
	log.Println(logPrefixDatabase, color.HiYellowString("Database opened, contains %d entries...", cachedDownloadID))

	// Image Store
	if config.FilterDuplicateImages {
		imgStore = duplo.New()
		if _, err := os.Stat(imgStorePath); err == nil {
			log.Println(logPrefixDatabase, color.YellowString("Opening image filter database..."))
			storeFile, err := ioutil.ReadFile(imgStorePath)
			if err != nil {
				log.Println(logPrefixDatabase, color.HiRedString("Error opening imgStore file:\t%s", err))
			} else {
				err = imgStore.GobDecode(storeFile)
				if err != nil {
					log.Println(logPrefixDatabase, color.HiRedString("Error decoding imgStore:\t%s", err))
				}
				if imgStore != nil {
					log.Println(logPrefixDatabase, color.HiYellowString("filterDuplicateImages database opened", imgStore.Size()))
				}
			}
		}
	}

	//#endregion

	//#region Component Initialization

	// Regex
	err = compileRegex()
	if err != nil {
		log.Println(logPrefixRegex, color.HiRedString("Error initializing:\t%s", err))
		return
	}

	// Twitter API
	if config.Credentials.TwitterAccessToken != "" &&
		config.Credentials.TwitterAccessTokenSecret != "" &&
		config.Credentials.TwitterConsumerKey != "" &&
		config.Credentials.TwitterConsumerSecret != "" {

		log.Println(logPrefixTwitter, color.MagentaString("Connecting to API..."))

		twitterClient = anaconda.NewTwitterApiWithCredentials(
			config.Credentials.TwitterAccessToken,
			config.Credentials.TwitterAccessTokenSecret,
			config.Credentials.TwitterConsumerKey,
			config.Credentials.TwitterConsumerSecret,
		)

		twitterSelf, err := twitterClient.GetSelf(url.Values{})
		if err != nil {
			log.Println(logPrefixTwitter, color.HiRedString("API Login Error: %s", err.Error()))
			log.Println(logPrefixTwitter, color.MagentaString("Error encountered while connecting to API, the bot won't use the Twitter API."))
		} else {
			log.Println(logPrefixTwitter, color.HiMagentaString("Connected to API @%s", twitterSelf.ScreenName))
			twitterConnected = true
		}
	} else {
		log.Println(logPrefixTwitter, color.MagentaString("API credentials missing, the bot won't use the Twitter API."))
	}

	// Google Drive Client
	if config.Credentials.GoogleDriveCredentialsJSON != "" {
		log.Println(logPrefixGoogleDrive, color.MagentaString("Connecting..."))
		ctx := context.Background()
		authJson, err := ioutil.ReadFile(config.Credentials.GoogleDriveCredentialsJSON)
		if err != nil {
			log.Println(logPrefixGoogleDrive, color.HiRedString("Error opening Google Credentials JSON:\t%s", err))
		} else {
			googleConfig, err := google.JWTConfigFromJSON(authJson, drive.DriveReadonlyScope)
			if err != nil {
				log.Println(logPrefixGoogleDrive, color.HiRedString("Error parsing Google Credentials JSON:\t%s", err))
			} else {
				client := googleConfig.Client(ctx)
				googleDriveService, err = drive.New(client)
				if err != nil {
					log.Println(logPrefixGoogleDrive, color.HiRedString("Error setting up client:\t%s", err))
				} else {
					log.Println(logPrefixGoogleDrive, color.HiMagentaString("Connected!"))
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
	if config.DebugOutput {
		log.Println(logPrefixDebugLabel("Validation"), color.HiYellowString("Validating configured channels/servers..."))
	}
	//-
	if config.AdminChannels != nil {
		for _, adminChannel := range config.AdminChannels {
			if adminChannel.ChannelIDs != nil {
				for _, subchannel := range *adminChannel.ChannelIDs {
					_, err := bot.State.Channel(subchannel)
					if err != nil {
						invalidAdminChannels = append(invalidAdminChannels, subchannel)
						log.Println(logPrefixErrorLabel("Validation"), color.HiRedString("Bot cannot access admin subchannel %s...\t%s", subchannel, err))
					}
				}

			} else {
				_, err := bot.State.Channel(adminChannel.ChannelID)
				if err != nil {
					invalidAdminChannels = append(invalidAdminChannels, adminChannel.ChannelID)
					log.Println(logPrefixErrorLabel("Validation"), color.HiRedString("Bot cannot access admin channel %s...\t%s", adminChannel.ChannelID, err))
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
					log.Println(logPrefixErrorLabel("Validation"), color.HiRedString("Bot cannot access subserver %s...\t%s", subserver, err))
				}
			}
		} else {
			_, err := bot.State.Guild(server.ServerID)
			if err != nil {
				invalidServers = append(invalidServers, server.ServerID)
				log.Println(logPrefixErrorLabel("Validation"), color.HiRedString("Bot cannot access server %s...\t%s", server.ServerID, err))
			}
		}
	}
	for _, channel := range config.Channels {
		if channel.ChannelIDs != nil {
			for _, subchannel := range *channel.ChannelIDs {
				_, err := bot.State.Channel(subchannel)
				if err != nil {
					invalidChannels = append(invalidChannels, subchannel)
					log.Println(logPrefixErrorLabel("Validation"), color.HiRedString("Bot cannot access subchannel %s...\t%s", subchannel, err))
				}
			}

		} else {
			_, err := bot.State.Channel(channel.ChannelID)
			if err != nil {
				invalidChannels = append(invalidChannels, channel.ChannelID)
				log.Println(logPrefixErrorLabel("Validation"), color.HiRedString("Bot cannot access channel %s...\t%s", channel.ChannelID, err))
			}
		}
	}
	//-
	invalidSources := len(invalidAdminChannels) + len(invalidChannels) + len(invalidServers)
	if invalidSources > 0 {
		log.Println(logPrefixErrorLabel("Validation"), color.HiRedString("Found %d invalid channels/servers in configuration...", invalidSources))
		logMsg := fmt.Sprintf("Validation found %d invalid sources...\n", invalidSources)
		if len(invalidAdminChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Admin Channels: (%d)** - %s", len(invalidAdminChannels), strings.Join(invalidAdminChannels, ", "))
		}
		if len(invalidServers) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Servers: (%d)** - %s", len(invalidServers), strings.Join(invalidServers, ", "))
		}
		if len(invalidChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Channels: (%d)** - %s", len(invalidChannels), strings.Join(invalidChannels, ", "))
		}
		logErrorMessage(logMsg)
	} else if config.DebugOutput {
		log.Println(logPrefixDebugLabel("Validation"), color.HiGreenString("All channels/servers successfully validated!"))
	}

	// Start Presence
	timeLastUpdated = time.Now()
	updateDiscordPresence()

	//#endregion

	// Output Done
	if config.DebugOutput {
		log.Println(color.YellowString("Startup finished, took %s...", uptime()))
	}
	log.Println(color.HiCyanString(wrapHyphensW(fmt.Sprintf("%s v%s is online and connected to %d server%s", projectLabel, projectVersion, len(bot.State.Guilds), pluralS(len(bot.State.Guilds))))))
	log.Println(color.RedString("CTRL+C to exit..."))

	// Log Status
	logStatusMessage(logStatusStartup)

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
			log.Println(logPrefixDebug, "[Constants]", color.HiYellowString("%s already cached (processing %s, has %s stored)", serverKey, server.ID, constants[serverKey]))
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
				channelKey := fmt.Sprintf("CHANNEL_%s_%s_%s", stripSymbols(server.Name), stripSymbols(categoryName), stripSymbols(channel.Name))
				channelKey = strings.ReplaceAll(channelKey, " ", "_")
				for strings.Contains(channelKey, "__") {
					channelKey = strings.ReplaceAll(channelKey, "__", "_")
				}
				channelKey = strings.ToUpper(channelKey)
				if constants[channelKey] == "" {
					constants[channelKey] = channel.ID
				} else if config.DebugOutput {
					log.Println(logPrefixDebug, "[Constants]", color.HiYellowString("%s already cached (processing %s/%s, has %s stored)", channelKey, server.ID, channel.ID, constants[channelKey]))
				}
			}
		}
	}
	//--- Save constants
	os.MkdirAll(cachePath, 0755)
	if _, err := os.Stat(constantsPath); err == nil {
		err = os.Remove(constantsPath)
		if err != nil {
			log.Println("[Constants]", color.HiRedString("Encountered error deleting cache file:\t%s", err))
		}
	}
	constantsStruct := constStruct{}
	constantsStruct.Constants = constants
	newJson, err := json.MarshalIndent(constantsStruct, "", "\t")
	if err != nil {
		log.Println("[Constants]", color.HiRedString("Failed to format constants...\t%s", err))
	} else {
		err := ioutil.WriteFile(constantsPath, newJson, 0644)
		if err != nil {
			log.Println("[Constants]", color.HiRedString("Failed to save new constants file...\t%s", err))
		}
	}
	//#endregion

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
		log.Println(logPrefixHistory, color.HiYellowString("History Autoruns completed (for %d channel%s)", len(autorunHistoryChannels), pluralS(len(autorunHistoryChannels))))
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
						log.Println(logPrefixSettings, color.YellowString("Detected changes in \"%s\", reloading...", configFile))
						loadConfig()
						log.Println(logPrefixSettings, color.HiYellowString("Reloaded - bound to %d channel%s and %d server%s",
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

	// Infinite loop until interrupted
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	logStatusMessage(logStatusExit)

	log.Println(logPrefixDiscord, color.GreenString("Logging out of discord..."))
	bot.Close()

	log.Println(logPrefixDatabase, color.YellowString("Closing database..."))
	myDB.Close()

	log.Println(color.HiRedString("Exiting... "))
}

func botLogin() {
	var err error

	if config.Credentials.Token != "" && config.Credentials.Token != placeholderToken {
		log.Println(logPrefixDiscord, color.GreenString("Connecting to Discord via Token..."))
		if config.Credentials.UserBot {
			bot, err = discordgo.New(config.Credentials.Token)
		} else {
			bot, err = discordgo.New("Bot " + config.Credentials.Token)
		}
	} else if (config.Credentials.Email != "" && config.Credentials.Email != placeholderEmail) &&
		(config.Credentials.Password != "" && config.Credentials.Password != placeholderPassword) {
		log.Println(logPrefixDiscord, color.GreenString("Connecting via Login..."))
		bot, err = discordgo.New(config.Credentials.Email, config.Credentials.Password)
	} else {
		log.Println(logPrefixDiscord, color.HiRedString("No valid credentials for Discord..."))
		properExit()
	}
	if err != nil {
		// Newer discordgo throws this error for some reason with Email/Password login
		if err.Error() != "Unable to fetch discord authentication token. <nil>" {
			log.Println(logPrefixDiscord, color.HiRedString("Error logging in: %s", err))
			properExit()
		}
	}

	// Connect Bot
	bot.LogLevel = -1 // to ignore dumb wsapi error
	err = bot.Open()
	if err != nil {
		log.Println(logPrefixDiscord, color.HiRedString("Discord login failed:\t%s", err))
		properExit()
	}
	bot.LogLevel = config.DiscordLogLevel // reset
	bot.ShouldReconnectOnError = true

	// Fetch Bot's User Info
	user, err = bot.User("@me")
	if err != nil {
		user = bot.State.User
		if user == nil {
			log.Println(logPrefixDiscord, color.HiRedString("Error obtaining user details: %s", err))
			loop <- syscall.SIGINT
		} else {
			log.Println(logPrefixDiscord, color.RedString("Using backup user data, hopefully it works..."))
		}
	} else if user == nil {
		log.Println(logPrefixDiscord, color.HiRedString("No error encountered obtaining user details, but it's empty..."))
		loop <- syscall.SIGINT
	} else {
		log.Println(logPrefixDiscord, color.HiGreenString("Logged into %s", getUserIdentifier(*user)))
		if user.Bot {
			log.Println(logPrefixDiscord, color.MagentaString("This is a Bot User"))
			log.Println(logPrefixDiscord, color.MagentaString("- Status presence details are limited."))
			log.Println(logPrefixDiscord, color.MagentaString("- Access is restricted to servers you have permission to add the bot to."))
		} else {
			log.Println(logPrefixDiscord, color.MagentaString("This is a User Account (Self-Bot)"))
			log.Println(logPrefixDiscord, color.MagentaString("- Discord does not allow Automated User Accounts (Self-Bots), so by using this bot you potentially risk account termination."))
			log.Println(logPrefixDiscord, color.MagentaString("- See GitHub page for link to Discord's official statement."))
			log.Println(logPrefixDiscord, color.MagentaString("- If you wish to avoid this, use a Bot account if possible."))
		}
	}
}
