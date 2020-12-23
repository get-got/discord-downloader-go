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
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/rivo/duplo"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

/* TODO:

- Permission checks are very messy.
	- Solution may be to wrap all editing and sending within functions that check permissions universally.

- Consolidate history command variations

*/

var (
	bot      *discordgo.Session
	user     *discordgo.User
	myDB     *db.DB
	imgStore *duplo.Store
	loop     chan os.Signal

	googleDriveService *drive.Service

	startTime        time.Time
	timeLastUpdated  time.Time
	cachedDownloadID int

	configReloadLastTime time.Time
)

func init() {
	loop = make(chan os.Signal, 1)
	startTime = time.Now()
	historyStatus = make(map[string]string)

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(color.HiCyanString("Welcome to %s v%s!", projectName, projectVersion))
	log.Println(color.CyanString("> discord-go v%s, Discord API v%s", discordgo.VERSION, discordgo.APIVersion))
}

func main() {
	var err error

	// Config
	log.Println(color.YellowString("Loading settings from \"%s\"...", configPath))
	loadConfig()
	log.Println(color.HiYellowString("Settings loaded, bound to %d channel(s)", getBoundChannelsCount()))

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
		log.Println(color.YellowString("Creating database, please wait..."))
		if err := myDB.Create("Downloads"); err != nil {
			log.Println(color.HiRedString("Error while trying to create database: %s", err))
			return
		}
		log.Println(color.HiYellowString("Created new database..."))
		log.Println(color.YellowString("Indexing database, please wait..."))
		if err := myDB.Use("Downloads").Index([]string{"URL"}); err != nil {
			log.Println(color.HiRedString("Unable to create database index for URL: %s", err))
			return
		}
		if err := myDB.Use("Downloads").Index([]string{"ChannelID"}); err != nil {
			log.Println(color.HiRedString("Unable to create database index for ChannelID: %s", err))
			return
		}
		if err := myDB.Use("Downloads").Index([]string{"UserID"}); err != nil {
			log.Println(color.HiRedString("Unable to create database index for UserID: %s", err))
			return
		}
		log.Println(color.HiYellowString("Created database indexes..."))
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
				}
			}
		}
	}

	//#endregion

	//#region Discord Initialization

	// Bot Login
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
	bot.LogLevel = 0 // reset

	// Fetch Bot's User Info
	user, err = bot.User("@me")
	if err != nil {
		log.Println(color.HiRedString("Error obtaining bot user details: %s", err))
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

	// Commands
	handleCommands()

	// Event Handlers
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

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

	//#region Background Tasks

	// Tickers
	if config.DebugOutput {
		log.Println(logPrefixDebug, color.YellowString("Starting background loops..."))
	}
	ticker5m := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker5m.C:
				// If bot experiences connection interruption the status will go blank until updated by message, this fixes that
				updateDiscordPresence()
			}
		}
	}()

	// Autorun History on Launch
	var autorunHistoryChannels []configurationChannel
	for _, item := range config.Channels {
		if item.OverwriteAutorunHistory != nil {
			if *item.OverwriteAutorunHistory {
				autorunHistoryChannels = append(autorunHistoryChannels, item)
			}
			continue
		}
		if config.AutorunHistory {
			autorunHistoryChannels = append(autorunHistoryChannels, item)
		}
	}
	for _, item := range autorunHistoryChannels {
		if item.ChannelID != "" {
			handleHistory(nil, item.ChannelID)
		} else if *item.ChannelIDs != nil {
			for _, subchannel := range *item.ChannelIDs {
				handleHistory(nil, subchannel)
			}
		}
	}
	if len(autorunHistoryChannels) > 0 {
		log.Println(logPrefixHistory, color.HiYellowString("History Autoruns completed"))
		log.Println(color.CyanString("Waiting for something else to do..."))
	}

	// Settings Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(color.HiRedString("[Watchers] Error creating NewWatcher:\t%s", err))
	}
	defer watcher.Close()
	err = watcher.Add(configPath)
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
						log.Println(color.YellowString("Detected changes in \"%s\", reloading settings...", configPath))
						loadConfig()
						log.Println(color.HiYellowString("Settings reloaded, bound to %d channel(s)", getBoundChannelsCount()))

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

	log.Println(color.GreenString("Logging out of discord..."))
	bot.Close()

	log.Println(color.YellowString("Closing database..."))
	myDB.Close()

	log.Println(color.HiRedString("Exiting..."))
}
