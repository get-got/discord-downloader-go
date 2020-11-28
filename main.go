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
	"github.com/hako/durafmt"
	"github.com/rivo/duplo"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

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
)

func init() {
	loop = make(chan os.Signal, 1)
	startTime = time.Now()
	historyCommandActive = make(map[string]string)

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
			log.Println(color.HiCyanString("Update available on %s\n", projectReleaseURL))
			time.Sleep(5 * time.Second)
		}
	}

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

	// Regex
	err = compileRegex()
	if err != nil {
		log.Println(color.HiRedString("Error initializing Regex:\t%s", err))
		return
	}

	// Bot Login
	if config.Credentials.Token != "" && config.Credentials.Token != placeholderToken {
		log.Println(color.GreenString("Connecting to Discord via Token..."))
		bot, err = discordgo.New("Bot " + config.Credentials.Token)
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

	// Open Bot, Fetch User
	err = bot.Open()
	if err == nil {
		user, err = bot.User("@me")
		if err != nil {
			log.Println(color.HiRedString("Error obtaining bot user details: %s", err))
		} else {
			log.Println(color.HiGreenString("Discord logged into %s", getUserIdentifier(*user)))
			if user.Bot {
				log.Println(logPrefixHelper, color.MagentaString("This is a Bot User..."))
				log.Println(logPrefixHelper, color.MagentaString("- Status presence details are limited."))
				log.Println(logPrefixHelper, color.MagentaString("- Server access is restricted to servers you have permission to add the bot to."))
			} else {
				log.Println(logPrefixHelper, color.MagentaString("This is a User Account (Self-Bot)..."))
				log.Println(logPrefixHelper, color.MagentaString("- Discord does not allow Automated User Accounts (Self-Bots), so by using this bot you potentially risk account termination."))
				log.Println(logPrefixHelper, color.MagentaString("- See GitHub page for link to Discord's official statement."))
				log.Println(logPrefixHelper, color.MagentaString("- If you wish to avoid this, use a Bot account if possible."))
			}
		}
	} else {
		log.Println(color.HiRedString("Discord login failed:\t%s", err))
	}

	// Command Router
	router := exrouter.New()

	// Commands: Utility
	router.On("ping", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:ping]")
		if isCommandableChannel(ctx.Msg) {
			beforePong := time.Now()
			pong, err := ctx.Reply("Pong!")
			if err != nil {
				log.Println(logPrefixHere, color.HiRedString("Error sending pong message:\t%s", err))
			} else {
				afterPong := time.Now()
				latency := bot.HeartbeatLatency().Milliseconds()
				roundtrip := afterPong.Sub(beforePong).Milliseconds()
				mention := ctx.Msg.Author.Mention()
				content := fmt.Sprintf("**Latency:** ``%dms`` — **Roundtrip:** ``%dms``",
					latency,
					roundtrip,
				)
				if pong != nil {
					_, err := bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
						ID:      pong.ID,
						Channel: pong.ChannelID,
						Content: &mention,
						Embed:   buildEmbed(ctx.Msg.ChannelID, "Command — Ping", content),
					})
					// Failed to edit pong
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to edit pong message, sending new one:\t%s", err))
						_, err := replyEmbed(pong, "Command — Ping", content)
						// Failed to send new pong
						if err != nil {
							log.Println(logPrefixHere, color.HiRedString("Failed to send replacement pong message:\t%s", err))
						}
					}
				}
				// Log
				log.Println(logPrefixHere, color.HiCyanString("%s pinged bot - Latency: %dms, Roundtrip: %dms",
					getUserIdentifier(*ctx.Msg.Author),
					latency,
					roundtrip),
				)
			}
		}
	}).Cat("Utility").Alias("test").Desc("Pings the bot")

	router.Default = router.On("help", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:help]")
		if isCommandableChannel(ctx.Msg) {
			text := ""
			for _, cmd := range router.Routes {
				if cmd.Category != "Admin" || isBotAdmin(ctx.Msg) {
					text += fmt.Sprintf("• \"%s\" : %s",
						cmd.Name,
						cmd.Description,
					)
					if len(cmd.Aliases) > 0 {
						text += fmt.Sprintf("\n— Aliases: \"%s\"", strings.Join(cmd.Aliases, "\", \""))
					}
					text += "\n\n"
				}
			}
			_, err := replyEmbed(ctx.Msg, "Command — Help", fmt.Sprintf("Use commands as ``\"%s<command> <arguments?>\"``\n```%s```\n%s", config.CommandPrefix, text, projectRepoURL))
			// Failed to send
			if err != nil {
				log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
			}
			log.Println(logPrefixHere, color.HiCyanString("%s asked for help", getUserIdentifier(*ctx.Msg.Author)))
		}
	}).Cat("Utility").Alias("commands").Desc("Outputs this help menu")

	// Commands: Info
	router.On("status", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:status]")
		if isCommandableChannel(ctx.Msg) {
			message := fmt.Sprintf("• **Uptime —** %s\n"+
				"• **Started at —** %s\n"+
				"• **Joined Servers —** %d\n"+
				"• **Bound Channels —** %d\n"+
				"• **Admin Channels —** %d\n"+
				"• **Heartbeat Latency —** %dms",
				durafmt.Parse(time.Since(startTime)).String(),
				startTime.Format("03:04:05pm on Monday, January 2, 2006 (MST)"),
				len(bot.State.Guilds),
				getBoundChannelsCount(),
				len(config.AdminChannels),
				bot.HeartbeatLatency().Milliseconds(),
			)
			if isChannelRegistered(ctx.Msg.ChannelID) {
				configJson, _ := json.MarshalIndent(getChannelConfig(ctx.Msg.ChannelID), "", "\t")
				message = message + fmt.Sprintf("\n• **Channel Settings...** ```%s```", string(configJson))
			}
			_, err := replyEmbed(ctx.Msg, "Command — Status", message)
			// Failed to send
			if err != nil {
				log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
			}
			log.Println(logPrefixHere, color.HiCyanString("%s requested status report", getUserIdentifier(*ctx.Msg.Author)))
		}
	}).Cat("Info").Alias("info").Desc("Displays info regarding the current status of the bot")

	router.On("stats", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:stats]")
		if isChannelRegistered(ctx.Msg.ChannelID) {
			channelConfig := getChannelConfig(ctx.Msg.ChannelID)
			if *channelConfig.AllowCommands {
				content := fmt.Sprintf("• **Total Downloads —** %s\n"+
					"• **Downloads in this Channel —** %s",
					formatNumber(int64(dbDownloadCount())),
					formatNumber(int64(dbDownloadCountByChannel(ctx.Msg.ChannelID))),
				)
				//TODO: Count in channel by users
				_, err := replyEmbed(ctx.Msg, "Command — Stats", content)
				// Failed to send
				if err != nil {
					log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s requested stats", getUserIdentifier(*ctx.Msg.Author)))
			}
		}
	}).Cat("Info").Desc("Outputs statistics regarding this channel")

	// Commands: Admin
	router.On("history", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:history]")
		channel := ctx.Msg.ChannelID
		args := ctx.Args.After(1)
		if isChannelRegistered(channel) { // Local
			channelConfig := getChannelConfig(channel)
			if *channelConfig.AllowCommands {
				if isLocalAdmin(ctx.Msg) {
					// Cancel Local
					if historyCommandActive[channel] == "downloading" && strings.ToLower(strings.TrimSpace(args)) == "cancel" {
						historyCommandActive[channel] = "cancel"
						_, err := replyEmbed(ctx.Msg, "Command — History", cmderrHistoryCancelled)
						if err != nil {
							log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
						}
						log.Println(logPrefixHere, color.CyanString("%s cancelled history cataloging for %s", getUserIdentifier(*ctx.Msg.Author), channel))
					} else { // Start Local
						_, historyCommandIsSet := historyCommandActive[channel]
						if !historyCommandIsSet || historyCommandActive[channel] == "" {
							historyCommandActive[channel] = ""
							handleHistory(ctx.Msg, channel, channel)
						} else {
							log.Println(logPrefixHere, color.CyanString("%s tried using history command but history is already running for %s...", getUserIdentifier(*ctx.Msg.Author), channel))
						}
					}
				} else {
					_, err := replyEmbed(ctx.Msg, "Command — History", cmderrLackingLocalAdminPerms)
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
					}
					log.Println(logPrefixHere, color.CyanString("%s tried to cache history for %s but lacked local admin perms.", getUserIdentifier(*ctx.Msg.Author), channel))
				}
			}
		} else if isAdminChannelRegistered(channel) { // Designated
			if isBotAdmin(ctx.Msg) {
				channels := strings.Split(args, ",")
				if len(channels) > 0 {
					// Cancel Designated
					if strings.ToLower(strings.TrimSpace(ctx.Args.Get(1))) == "cancel" {
						channels = strings.Split(ctx.Args.After(2), ",")
						for _, channelValue := range channels {
							channelValue = strings.TrimSpace(channelValue)
							if historyCommandActive[channelValue] == "downloading" {
								historyCommandActive[channelValue] = "cancel"
								_, err := replyEmbed(ctx.Msg, "Command — History", cmderrHistoryCancelled)
								if err != nil {
									log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
								}
								log.Println(logPrefixHere, color.CyanString("%s cancelled history cataloging for %s", getUserIdentifier(*ctx.Msg.Author), channelValue))
							}
						}
					} else { // Start Designated
						for _, channelValue := range channels {
							channelValue = strings.TrimSpace(channelValue)
							if isChannelRegistered(channelValue) {
								_, historyCommandIsSet := historyCommandActive[channelValue]
								if !historyCommandIsSet || historyCommandActive[channelValue] == "" {
									historyCommandActive[channelValue] = ""
									handleHistory(ctx.Msg, channel, channelValue)
								} else {
									log.Println(logPrefixHere, color.CyanString("Tried using history command but history is already running for %s...", channelValue))
								}
							} else {
								replyEmbed(ctx.Msg, "Command — History", cmderrChannelNotRegistered)
								log.Println(logPrefixHere, color.CyanString("%s tried to cache history for %s but channel is not registered...", getUserIdentifier(*ctx.Msg.Author), channelValue))
							}
						}
					}
				} else {
					_, err := replyEmbed(ctx.Msg, "Command — History", "Please enter valid channel ID(s)...\n\n_Ex:_ ``<prefix>history <id1>,<id2>,<id3>``")
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
					}
					log.Println(logPrefixHere, color.CyanString("%s tried to cache history but input no channels", getUserIdentifier(*ctx.Msg.Author)))
				}
			} else {
				_, err := replyEmbed(ctx.Msg, "Command — History", cmderrLackingBotAdminPerms)
				if err != nil {
					log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
				}
				log.Println(logPrefixHere, color.CyanString("%s tried to cache history for %s but lacked bot admin perms.", getUserIdentifier(*ctx.Msg.Author), channel))
			}
		} else {
			log.Println(logPrefixHere, color.CyanString("%s tried to catalog history for %s but channel is not registered...", getUserIdentifier(*ctx.Msg.Author), channel))
		}
	}).Alias("catalog", "cache").Cat("Admin").Desc("Catalogs history for this channel")

	router.On("exit", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:exit]")
		if isCommandableChannel(ctx.Msg) {
			if isBotAdmin(ctx.Msg) {
				_, err := replyEmbed(ctx.Msg, "Command — Exit", "Exiting...")
				if err != nil {
					log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s requested exit, goodbye...", getUserIdentifier(*ctx.Msg.Author)))
				loop <- syscall.SIGINT
			} else {
				_, err := replyEmbed(ctx.Msg, "Command — Exit", cmderrLackingBotAdminPerms)
				if err != nil {
					log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s tried to exit but lacked bot admin perms.", getUserIdentifier(*ctx.Msg.Author)))
			}
		}
	}).Alias("reload", "kill").Cat("Admin").Desc("Kills the bot")

	//TODO: add_channel command
	//TODO: edit_channel command
	//TODO: delete_channel command
	//NOTE: The problem with these is opeaning, modifying, then saving the JSON without adding unwanted junk.
	// Also the hoops the user will have to jump through to edit these via commands.

	// Handler for Command Router
	bot.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		//NOTE: This setup makes it case-insensitive but message content will be lowercase, currently case sensitivity is not necessary.
		router.FindAndExecute(bot, strings.ToLower(config.CommandPrefix), bot.State.User.ID, messageToLower(m.Message))
	})

	// Event Handlers
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

	// Start Presence
	timeLastUpdated = time.Now()
	updateDiscordPresence()

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

	// Output Startup Duration
	if config.DebugOutput {
		log.Println(logPrefixDebug, color.YellowString("Startup finished, took %s...", uptime()))
	}

	// Output Done
	log.Println(color.HiCyanString("%s is online! Connected to %d server(s)", projectLabel, len(bot.State.Guilds)))

	// Infinite loop until interrupted
	log.Println(color.RedString("Ctrl+C to exit..."))
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	log.Println(color.GreenString("Logging out of discord..."))
	bot.Close()

	log.Println(color.YellowString("Closing database..."))
	myDB.Close()

	log.Println(color.HiRedString("Exiting..."))
}
