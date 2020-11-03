package main

import (
	"encoding/json"
	"fmt"
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
	"github.com/hashicorp/go-version"
)

var (
	bot  *discordgo.Session
	user *discordgo.User
	myDB *db.DB

	loop chan os.Signal
)

func init() {
	startTime = time.Now()
	loop = make(chan os.Signal, 1)
	historyCommandActive = make(map[string]string)

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(color.HiCyanString("Welcome to %s v%s!",
		PROJECT_NAME,
		PROJECT_VERSION,
	))
	log.Println(color.CyanString("> discord-go v%s, Discord API v%s",
		discordgo.VERSION,
		discordgo.APIVersion,
	))
}

func main() {
	var err error

	// Config
	log.Println(color.YellowString("Loading settings from \"%s\"...", LOC_CONFIG_FILE))
	loadConfig()
	log.Println(color.HiYellowString("Settings loaded, bound to %d channel(s)", len(config.Channels)))

	if config.GithubUpdateChecking {
		if !isLatestRelease() {
			log.Println(color.HiCyanString("Update available on %s\n", PROJECT_RELEASE_URL))
			time.Sleep(5 * time.Second)
		}
	}

	// Initialize Database
	log.Println(color.YellowString("Opening database..."))
	myDB, err = db.OpenDB(LOC_DATABASE_DIR)
	if err != nil {
		log.Println(color.HiRedString("Unable to open database: %s", err))
		return
	}
	if myDB.Use("Downloads") == nil {
		log.Println(color.YellowString("Creating database, please wait..."))
		if err := myDB.Create("Downloads"); err != nil {
			log.Println(color.HiRedString("Error while trying to create database: %s", err))
			return
		} else {
			log.Println(color.HiYellowString("Created new database..."))
		}
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

	// For download counter
	cachedDownloadID = dbDownloadCount()

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

	// Initialize Regex
	err = compileRegex()
	if err != nil {
		log.Println(color.HiRedString("Error initializing Regex: %s", err.Error()))
		return
	}

	// Establish bot
	if config.Credentials.Token != "" && config.Credentials.Token != PLACEHOLDER_TOKEN {
		log.Println(color.GreenString("Connecting to Discord via Token..."))
		bot, err = discordgo.New("Bot " + config.Credentials.Token)
	} else if (config.Credentials.Email != "" && config.Credentials.Email != PLACEHOLDER_EMAIL) &&
		(config.Credentials.Password != "" && config.Credentials.Password != PLACEHOLDER_PASSWORD) {
		log.Println(color.GreenString("Connecting to Discord via Login..."))
		bot, err = discordgo.New(config.Credentials.Email, config.Credentials.Password)
	} else {
		log.Println(color.HiRedString("No valid credentials for Discord..."))
		delayedExit()
	}
	if err != nil {
		// Newer discordgo throws this error for some reason with Email/Password login
		if err.Error() != "Unable to fetch discord authentication token. <nil>" {
			log.Println(color.HiRedString("Error logging into Discord: %s", err))
			delayedExit()
		}
	}

	// Establish command router
	router := exrouter.New()

	// Utility Commands
	router.On("ping", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:ping]")
		if isCommandableChannel(ctx.Msg) {
			before_pong := time.Now()
			pong, err := ctx.Reply("Pong!")
			if err != nil {
				log.Println(logPrefixHere, color.HiRedString("Error sending pong message: ", err))
			} else {
				after_pong := time.Now()
				roundtrip := after_pong.Sub(before_pong)
				latency := bot.HeartbeatLatency()
				mention := ctx.Msg.Author.Mention()
				bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
					ID:      pong.ID,
					Channel: pong.ChannelID,
					Content: &mention,
					Embed: buildEmbed(ctx.Msg.ChannelID, "Command — Ping",
						fmt.Sprintf("**Latency:** ``%dms`` — **Roundtrip:** ``%dms``",
							latency.Milliseconds(),
							roundtrip.Milliseconds(),
						),
					),
				})
				//TODO: Message Edit error checking
				log.Println(logPrefixHere, color.HiCyanString("%s pinged bot - Latency: %dms, Roundtrip: %dms",
					userDisplay(*ctx.Msg.Author),
					latency.Milliseconds(),
					roundtrip.Milliseconds()),
				)
			}
		}
	}).Alias("test").Desc("Pings the bot").Cat("Utility")

	router.Default = router.On("help", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:help]")
		//TODO: PRE-1.0.0 - Update/Improve
		if isCommandableChannel(ctx.Msg) {
			text := ""
			for _, cmd := range router.Routes {
				if cmd.Category != "Admin" || adminCheck(ctx.Msg) {
					text += fmt.Sprintf("• %s : %s",
						cmd.Name,
						cmd.Description,
					)
					if len(cmd.Aliases) > 0 {
						text += fmt.Sprintf("\n— Aliases: %s", strings.Join(cmd.Aliases, ", "))
					}
					text += "\n\n"
				}
			}
			sendEmbed(ctx.Msg, "Command — Help", "```"+text+"```")
			//TODO: Message Send error checking
			log.Println(logPrefixHere, color.HiCyanString("%s asked for help", userDisplay(*ctx.Msg.Author)))
		}
	}).Alias("commands").Desc("Outputs this help menu").Cat("Utility")

	// Info Commands
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
				len(config.Channels),
				len(config.AdminChannels),
				bot.HeartbeatLatency().Milliseconds(),
			)
			if isChannelRegistered(ctx.Msg.ChannelID) {
				configJson, _ := json.MarshalIndent(getChannelConfig(ctx.Msg.ChannelID), "", "\t")
				message = message + fmt.Sprintf("\n• **Channel Settings...** ```%s```", string(configJson))
			}
			sendEmbed(ctx.Msg, "Command — Status", message)
			//TODO: Message Send error checking
			log.Println(logPrefixHere, color.HiCyanString("%s requested status report", userDisplay(*ctx.Msg.Author)))
		}
	}).Alias("info").Desc("Displays info regarding the current status of the bot").Cat("Info")

	router.On("stats", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:stats]")
		if isChannelRegistered(ctx.Msg.ChannelID) {
			channelConfig := getChannelConfig(ctx.Msg.ChannelID)
			if *channelConfig.AllowCommands {
				//TODO: Count in channel by users
				//TODO: Count in channel by domain
				sendEmbed(ctx.Msg, "Command — Stats",
					fmt.Sprintf("• **Total Downloads —** %s\n"+
						"• **Downloads in this Channel —** %s",
						formatNumber(int64(dbDownloadCount())),
						formatNumber(int64(dbDownloadCountByChannel(ctx.Msg.ChannelID))),
					),
				)
				//TODO: Message Send error checking
				log.Println(logPrefixHere, color.HiCyanString("%s requested stats", userDisplay(*ctx.Msg.Author)))
			}
		}
	}).Desc("Outputs statistics regarding this channel").Cat("Info")

	// Admin Commands
	router.On("history", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:history]")
		channel := ctx.Msg.ChannelID
		args := ctx.Args.After(1)
		if isChannelRegistered(channel) { // Local
			channelConfig := getChannelConfig(channel)
			if *channelConfig.AllowCommands {
				if adminCheckLocal(ctx.Msg) {
					// Cancel Local
					if historyCommandActive[channel] == "downloading" && strings.ToLower(strings.TrimSpace(args)) == "cancel" {
						historyCommandActive[channel] = "cancel"
						sendEmbed(ctx.Msg, "Command — History", CMDRESP_HISTORY_CANCELLED)
						log.Println(logPrefixHere, color.CyanString("%s cancelled history cataloging for %s", userDisplay(*ctx.Msg.Author), channel))
					} else { // Start Local
						_, historyCommandIsSet := historyCommandActive[channel]
						if !historyCommandIsSet || historyCommandActive[channel] == "" {
							historyCommandActive[channel] = ""
							handleHistory(ctx.Msg, channel, channel)
						} else {
							log.Println(logPrefixHere, color.CyanString("%s tried using history command but history is already running for %s...", userDisplay(*ctx.Msg.Author), channel))
						}
					}
				} else {
					sendEmbed(ctx.Msg, "Command — History", CMDERR_LACKING_LOCAL_ADMIN_PERMS)
					log.Println(logPrefixHere, color.CyanString("%s tried to cache history for %s but lacked local admin perms.", userDisplay(*ctx.Msg.Author), channel))
				}
			}
		} else if isAdminChannelRegistered(channel) { // Designated
			if adminCheck(ctx.Msg) {
				channels := strings.Split(args, ",")
				if len(channels) > 0 {
					// Cancel Designated
					if strings.ToLower(strings.TrimSpace(ctx.Args.Get(1))) == "cancel" {
						channels = strings.Split(ctx.Args.After(2), ",")
						for _, channelValue := range channels {
							channelValue = strings.TrimSpace(channelValue)
							if historyCommandActive[channelValue] == "downloading" {
								historyCommandActive[channelValue] = "cancel"
								sendEmbed(ctx.Msg, "Command — History", CMDRESP_HISTORY_CANCELLED)
								log.Println(logPrefixHere, color.CyanString("%s cancelled history cataloging for %s", userDisplay(*ctx.Msg.Author), channelValue))
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
								sendEmbed(ctx.Msg, "Command — History", CMDERR_CHANNEL_NOT_REGISTERED)
								log.Println(logPrefixHere, color.CyanString("%s tried to cache history for %s but channel is not registered...", userDisplay(*ctx.Msg.Author), channelValue))
							}
						}
					}
				} else {
					sendEmbed(ctx.Msg, "Command — History", "Please enter valid channel ID(s)...\n\n_Ex:_ ``<prefix>history <id1>,<id2>,<id3>``")
					log.Println(logPrefixHere, color.CyanString("%s tried to cache history but input no channels", userDisplay(*ctx.Msg.Author)))
				}
			} else {
				sendEmbed(ctx.Msg, "Command — History", CMDERR_LACKING_BOT_ADMIN_PERMS)
				log.Println(logPrefixHere, color.CyanString("%s tried to cache history for %s but lacked bot admin perms.", userDisplay(*ctx.Msg.Author), channel))
			}
		} else {
			log.Println(logPrefixHere, color.CyanString("%s tried to catalog history for %s but channel is not registered...", userDisplay(*ctx.Msg.Author), channel))
		}
	}).Alias("catalog", "cache").Desc("Catalogs history for this channel").Cat("Admin")

	router.On("exit", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:exit]")
		if isCommandableChannel(ctx.Msg) {
			if adminCheck(ctx.Msg) {
				sendEmbed(ctx.Msg, "Command — Exit", "Exiting...")
				log.Println(logPrefixHere, color.HiCyanString("%s requested exit, goodbye...", userDisplay(*ctx.Msg.Author)))
				loop <- syscall.SIGINT
			} else {
				sendEmbed(ctx.Msg, "Command — Exit", CMDERR_LACKING_BOT_ADMIN_PERMS)
				log.Println(logPrefixHere, color.HiCyanString("%s tried to exit but lacked bot admin perms.", userDisplay(*ctx.Msg.Author)))
			}
		}
	}).Alias("reload", "kill").Desc("Kills the bot").Cat("Admin")

	// Establish Commands
	bot.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(bot, config.CommandPrefix, bot.State.User.ID, m.Message)
	})
	//TODO: Case insensitive commands & prefix

	// Handle Events
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

	//TODO: Debug messages while establishing above

	// Check Bot
	err = bot.Open()
	if err == nil {
		user, err = bot.User("@me")
		if err != nil {
			log.Println(color.HiRedString("Error obtaining bot user details: %s", err))
		} else {
			log.Println(color.HiGreenString("Discord logged into %s", userDisplay(*user)))
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
		log.Println(color.HiRedString("Discord login failed: %s", err))
	}

	// Presence
	timeLastUpdated = time.Now()
	updateDiscordPresence()

	// Tickers
	if config.DebugOutput {
		log.Println(logPrefixDebug, color.YellowString("Starting background loops..."))
	}
	ticker_5m := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker_5m.C:
				// If bot experiences connection interruption the status will go blank until updated by message, this fixes that
				updateDiscordPresence()
			}
		}
	}()

	// Duration
	if config.DebugOutput {
		// fmt.Sprintf is used here to properly format Duration as a string, casting yields warnings.
		log.Println(logPrefixDebug, color.YellowString("Startup finished, took %s...", uptime()))
	}

	// Done
	log.Println(color.HiCyanString("%s is online! Connected to %d server(s)", PROJECT_LABEL, len(bot.State.Guilds)))
	// Loop & potentially logout
	log.Println(color.RedString("Ctrl+C to exit..."))
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop
	log.Println(color.GreenString("Logging out of discord..."))
	bot.Close()
	log.Println(color.YellowString("Closing database..."))
	myDB.Close()
	log.Println(color.HiRedString("Exiting..."))
}

func updateDiscordPresence() {
	if config.PresenceEnabled {
		// Vars
		countInt := int64(dbDownloadCount()) + *config.InflateCount
		count := formatNumber(countInt)
		countShort := formatNumberShort(countInt)
		timeShort := timeLastUpdated.Format("3:04pm")
		timeLong := timeLastUpdated.Format("3:04:05pm MST - January 1, 2006")

		// Defaults
		status_presence := fmt.Sprintf("%s - %s files", timeShort, countShort)
		status_details := timeLong
		status_state := fmt.Sprintf("%s files total", count)

		// Overwrite Presence
		if config.PresenceOverwrite != nil {
			status_presence = *config.PresenceOverwrite
			if status_presence != "" {
				status_presence = varKeyReplacement(status_presence)
			}
		}
		// Overwrite Details
		if config.PresenceOverwriteDetails != nil {
			status_details = *config.PresenceOverwriteDetails
			if status_details != "" {
				status_details = varKeyReplacement(status_details)
			}
		}
		// Overwrite State
		if config.PresenceOverwriteState != nil {
			status_state = *config.PresenceOverwriteState
			if status_state != "" {
				status_state = varKeyReplacement(status_state)
			}
		}

		// Update
		bot.UpdateStatusComplex(discordgo.UpdateStatusData{
			Game: &discordgo.Game{
				Name:    status_presence,
				Type:    config.PresenceType,
				Details: status_details, // Only visible if real user
				State:   status_state,   // Only visible if real user
			},
			Status: config.PresenceStatus,
		})
	}
}

func varKeyReplacement(input string) string {
	countInt := int64(dbDownloadCount()) + *config.InflateCount
	timeNow := time.Now()
	keys := [][]string{
		{"{{dgVersion}}", discordgo.VERSION},
		{"{{ddgVersion}}", PROJECT_VERSION},
		{"{{apiVersion}}", discordgo.APIVersion},
		{"{{count}}", formatNumber(countInt)},
		{"{{countShort}}", formatNumberShort(countInt)},
		{"{{numGuilds}}", fmt.Sprint(len(bot.State.Guilds))},
		{"{{numChannels}}", fmt.Sprint(len(config.Channels))},
		{"{{numAdminChannels}}", fmt.Sprint(len(config.AdminChannels))},
		{"{{numAdmins}}", fmt.Sprint(len(config.Admins))},
		{"{{timeSavedShort}}", timeLastUpdated.Format("3:04pm")},
		{"{{timeSavedLong}}", timeLastUpdated.Format("3:04:05pm MST - January 1, 2006")},
		{"{{timeSavedShort24}}", timeLastUpdated.Format("15:04")},
		{"{{timeSavedLong24}}", timeLastUpdated.Format("15:04:05 MST - 1 January, 2006")},
		{"{{timeNowShort}}", timeNow.Format("3:04pm")},
		{"{{timeNowLong}}", timeNow.Format("3:04:05pm MST - January 1, 2006")},
		{"{{timeNowShort24}}", timeNow.Format("15:04")},
		{"{{timeNowLong24}}", timeNow.Format("15:04:05 MST - 1 January, 2006")},
		{"{{uptime}}", durafmt.ParseShort(time.Since(startTime)).String()},
	}
	//TODO: Case-insensitive key replacement.
	for _, key := range keys {
		if strings.Contains(input, key[0]) {
			input = strings.ReplaceAll(input, key[0], key[1])
		}
	}
	return input
}

type GithubReleaseApiObject struct {
	TagName string `json:"tag_name"`
}

func isLatestRelease() bool {
	prefixHere := color.HiMagentaString("[Github Update Check]")

	githubReleaseApiObject := new(GithubReleaseApiObject)
	err := getJson(PROJECT_RELEASE_API_URL, githubReleaseApiObject)
	if err != nil {
		log.Println(prefixHere, color.RedString("Error fetching current Release JSON: %s", err))
		return true
	}

	thisVersion, err := version.NewVersion(PROJECT_VERSION)
	if err != nil {
		log.Println(prefixHere, color.RedString("Error parsing current version: %s", err))
		return true
	}

	latestVersion, err := version.NewVersion(githubReleaseApiObject.TagName)
	if err != nil {
		log.Println(prefixHere, color.RedString("Error parsing latest version: %s", err))
		return true
	}

	if latestVersion.GreaterThan(thisVersion) {
		return false
	}

	return true
}
