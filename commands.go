package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/kennygrant/sanitize"
)

// Multiple use messages to save space and make cleaner.
// TODO: Implement this for more?
const (
	cmderrLackingBotAdminPerms = "You do not have permission to use this command. Your User ID must be set as a bot administrator in the settings file."
	cmderrSendFailure          = "Failed to send command message (requested by %s)...\t%s"
)

func handleCommands() *exrouter.Route {
	router := exrouter.New()

	//#region Utility Commands

	go router.On("ping", func(ctx *exrouter.Context) {
		if isCommandableChannel(ctx.Msg) {
			if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
				log.Println(lg("Command", "Ping", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
			} else {
				beforePong := time.Now()
				pong, err := ctx.Reply("Pong!")
				if err != nil {
					log.Println(lg("Command", "Ping", color.HiRedString, "Error sending pong message:\t%s", err))
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
						if selfbot {
							bot.ChannelMessageEdit(pong.ChannelID, pong.ID, fmt.Sprintf("%s **Command — Ping**\n\n%s", mention, content))
						} else {
							bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
								ID:      pong.ID,
								Channel: pong.ChannelID,
								Content: &mention,
								Embed:   buildEmbed(ctx.Msg.ChannelID, "Command — Ping", content),
							})
						}
					}
					// Log
					log.Println(lg("Command", "Ping", color.HiCyanString, "%s pinged bot - Latency: %dms, Roundtrip: %dms",
						getUserIdentifier(*ctx.Msg.Author),
						latency,
						roundtrip),
					)
				}
			}
		}
	}).Cat("Utility").Alias("test").Desc("Pings the bot")

	go router.On("help", func(ctx *exrouter.Context) {
		if isCommandableChannel(ctx.Msg) {
			if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
				log.Println(lg("Command", "Help", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
			} else {
				content := ""
				for _, cmd := range router.Routes {
					if cmd.Category != "Admin" || isBotAdmin(ctx.Msg) {
						content += fmt.Sprintf("• \"%s\" : %s",
							cmd.Name,
							cmd.Description,
						)
						if len(cmd.Aliases) > 0 {
							content += fmt.Sprintf("\n— Aliases: \"%s\"", strings.Join(cmd.Aliases, "\", \""))
						}
						content += "\n\n"
					}
				}
				if _, err := replyEmbed(ctx.Msg, "Command — Help",
					fmt.Sprintf("Use commands as ``\"%s<command> <arguments?>\"``\n```%s```\n%s",
						config.CommandPrefix, content, projectRepoURL)); err != nil {
					log.Println(lg("Command", "Help", color.HiRedString, cmderrSendFailure, getUserIdentifier(*ctx.Msg.Author), err))
				}
				log.Println(lg("Command", "Help", color.HiCyanString, "%s asked for help", getUserIdentifier(*ctx.Msg.Author)))
			}
		}
	}).Cat("Utility").Alias("commands").Desc("Outputs this help menu")

	//#endregion

	//#region Info Commands

	go router.On("status", func(ctx *exrouter.Context) {
		if isCommandableChannel(ctx.Msg) {
			if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
				log.Println(lg("Command", "Status", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
			} else {
				message := fmt.Sprintf("• **Uptime —** %s\n"+
					"• **Started at —** %s\n"+
					"• **Joined Servers —** %d\n"+
					"• **Bound Channels —** %d\n"+
					"• **Bound Cagetories —** %d\n"+
					"• **Bound Servers —** %d\n"+
					"• **Bound Users —** %d\n"+
					"• **Admin Channels —** %d\n"+
					"• **Heartbeat Latency —** %dms",
					durafmt.Parse(time.Since(startTime)).String(),
					startTime.Format("03:04:05pm on Monday, January 2, 2006 (MST)"),
					len(bot.State.Guilds),
					getBoundChannelsCount(),
					getBoundCategoriesCount(),
					getBoundServersCount(),
					getBoundUsersCount(),
					len(config.AdminChannels),
					bot.HeartbeatLatency().Milliseconds(),
				)
				if sourceConfig := getSource(ctx.Msg); sourceConfig != emptyConfig {
					configJson, _ := json.MarshalIndent(sourceConfig, "", "\t")
					message = message + fmt.Sprintf("\n• **Channel Settings...** ```%s```", string(configJson))
				}
				if _, err := replyEmbed(ctx.Msg, "Command — Status", message); err != nil {
					log.Println(lg("Command", "Status", color.HiRedString, cmderrSendFailure, getUserIdentifier(*ctx.Msg.Author), err))
				}
				log.Println(lg("Command", "Status", color.HiCyanString, "%s requested status report", getUserIdentifier(*ctx.Msg.Author)))
			}
		}
	}).Cat("Info").Desc("Displays info regarding the current status of the bot")

	go router.On("stats", func(ctx *exrouter.Context) {
		if isCommandableChannel(ctx.Msg) {
			if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
				log.Println(lg("Command", "Stats", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
			} else {
				if sourceConfig := getSource(ctx.Msg); sourceConfig != emptyConfig {
					if *sourceConfig.AllowCommands {
						content := fmt.Sprintf("• **Total Downloads —** %s\n"+
							"• **Downloads in this Channel —** %s",
							formatNumber(int64(dbDownloadCount())),
							formatNumber(int64(dbDownloadCountByChannel(ctx.Msg.ChannelID))),
						)
						//TODO: Count in channel by users
						if _, err := replyEmbed(ctx.Msg, "Command — Stats", content); err != nil {
							log.Println(lg("Command", "Stats", color.HiRedString, cmderrSendFailure,
								getUserIdentifier(*ctx.Msg.Author), err))
						}
						log.Println(lg("Command", "Stats", color.HiCyanString, "%s requested stats",
							getUserIdentifier(*ctx.Msg.Author)))
					}
				}
			}
		}
	}).Cat("Info").Desc("Outputs statistics regarding this channel")

	go router.On("info", func(ctx *exrouter.Context) {
		if isCommandableChannel(ctx.Msg) {
			if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
				log.Println(lg("Command", "Info", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
			} else {
				content := fmt.Sprintf("Here is some useful info...\n\n"+
					"• **Your User ID —** `%s`\n"+
					"• **Bots User ID —** `%s`\n"+
					"• **This Channel ID —** `%s`\n"+
					"• **This Server ID —** `%s`\n\n"+
					"• **Versions —`%s, discordgo v%s (modified), Discord API v%s`"+
					"\n\nRemember to remove any spaces when copying to settings.",
					ctx.Msg.Author.ID, botUser.ID, ctx.Msg.ChannelID, ctx.Msg.GuildID, runtime.Version(), discordgo.VERSION, discordgo.APIVersion)
				if _, err := replyEmbed(ctx.Msg, "Command — Info", content); err != nil {
					log.Println(lg("Command", "Info", color.HiRedString, cmderrSendFailure, getUserIdentifier(*ctx.Msg.Author), err))
				}
				log.Println(lg("Command", "Info", color.HiCyanString, "%s requested info", getUserIdentifier(*ctx.Msg.Author)))
			}
		}
	}).Cat("Info").Alias("debug").Desc("Displays info regarding Discord IDs")

	//#endregion

	//#region Admin Commands

	go router.On("history", func(ctx *exrouter.Context) {
		if isCommandableChannel(ctx.Msg) {
			// Vars
			var channels []string

			var shouldAbort bool = false
			var shouldProcess bool = true

			var before string
			var beforeID string
			var since string
			var sinceID string

			//#region Parse Args
			for argKey, argValue := range ctx.Args {
				if argKey == 0 { // skip head
					continue
				}
				if strings.Contains(strings.ToLower(argValue), "cancel") ||
					strings.Contains(strings.ToLower(argValue), "stop") {
					shouldAbort = true
				} else if strings.Contains(strings.ToLower(argValue), "help") ||
					strings.Contains(strings.ToLower(argValue), "info") {
					shouldProcess = false

					if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
						//content := fmt.Sprintf("")
						_, err := replyEmbed(ctx.Msg, "Command — History Help", "TODO: this")
						if err != nil {
							log.Println(lg("Command", "History",
								color.HiRedString, cmderrSendFailure, getUserIdentifier(*ctx.Msg.Author), err))
						}
					} else {
						log.Println(lg("Command", "History", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
					}
					log.Println(lg("Command", "History", color.CyanString, "%s requested history help.", getUserIdentifier(*ctx.Msg.Author)))
				} else if strings.Contains(strings.ToLower(argValue), "list") ||
					strings.Contains(strings.ToLower(argValue), "status") ||
					strings.Contains(strings.ToLower(argValue), "output") {
					shouldProcess = false

					output := "Running history jobs...\n"
					for channelID, job := range historyJobs {
						channelLabel := channelID
						channelInfo, err := bot.State.Channel(channelID)
						if err == nil {
							channelLabel = "#" + channelInfo.Name
						}
						output += fmt.Sprintf("• _%s_ - (%s)`%s`, `updated %s ago, added %s ago`\n",
							historyStatusLabel(job.Status), job.OriginUser, channelLabel,
							durafmt.ParseShort(time.Since(job.Updated)).String(), durafmt.ParseShort(time.Since(job.Added)).String())
						log.Println(lg("Command", "History", color.HiCyanString, "History Job: %s - (%s)%s, updated %s ago, added %s ago",
							historyStatusLabel(job.Status), job.OriginUser, channelLabel,
							durafmt.ParseShort(time.Since(job.Updated)).String(), durafmt.ParseShort(time.Since(job.Added)).String()))
					}
					if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
						_, err := ctx.Reply(output)
						if err != nil {
							log.Println(lg("Command", "History", color.HiRedString, cmderrSendFailure, getUserIdentifier(*ctx.Msg.Author), err))
						}
					} else {
						log.Println(lg("Command", "History", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
					}
					log.Println(lg("Command", "History", color.HiRedString, "%s requested statuses of history jobs.",
						getUserIdentifier(*ctx.Msg.Author)))
				} else if strings.Contains(strings.ToLower(argValue), "--before=") { // before key
					before = strings.ReplaceAll(strings.ToLower(argValue), "--before=", "")
					if isDate(before) {
						beforeID = discordTimestampToSnowflake("2006-01-02", dateLocalToUTC(before))
					} else if isNumeric(before) {
						beforeID = before
					}
					if config.DebugOutput {
						log.Println(lg("Command", "History", color.CyanString, "Date before range applied, snowflake %s, converts back to %s",
							beforeID, discordSnowflakeToTimestamp(beforeID, "2006-01-02T15:04:05.000Z07:00")))
					}
				} else if strings.Contains(strings.ToLower(argValue), "--since=") { //  since key
					since = strings.ReplaceAll(strings.ToLower(argValue), "--since=", "")
					if isDate(since) {
						sinceID = discordTimestampToSnowflake("2006-01-02", dateLocalToUTC(since))
					} else if isNumeric(since) {
						sinceID = since
					}
					if config.DebugOutput {
						log.Println(lg("Command", "History", color.CyanString, "Date since range applied, snowflake %s, converts back to %s",
							sinceID, discordSnowflakeToTimestamp(sinceID, "2006-01-02T15:04:05.000Z07:00")))
					}
				} else {
					// Actual Source ID(s)
					targets := strings.Split(ctx.Args.Get(argKey), ",")
					for _, target := range targets {
						if isNumeric(target) {
							// Test/Use if number is guild
							guild, err := bot.State.Guild(target)
							if err == nil {
								if config.DebugOutput {
									log.Println(lg("Command", "History", color.YellowString,
										"Specified target %s is a guild: \"%s\", adding all channels...",
										target, guild.Name))
								}
								for _, ch := range guild.Channels {
									channels = append(channels, ch.ID)
									if config.DebugOutput {
										log.Println(lg("Command", "History", color.YellowString,
											"Added %s (#%s in \"%s\") to history queue",
											ch.ID, ch.Name, guild.Name))
									}
								}
							} else { // Test/Use if number is channel
								ch, err := bot.State.Channel(target)
								if err == nil {
									channels = append(channels, target)
									if config.DebugOutput {
										log.Println(lg("Command", "History", color.YellowString, "Added %s (#%s in %s) to history queue",
											ch.ID, ch.Name, ch.GuildID))
									}
								}
							}
						} else if strings.Contains(strings.ToLower(target), "all") {
							channels = getAllRegisteredChannels()
						}
					}
				}
			}
			//#endregion

			//#region Process Channels
			if shouldProcess {
				// Local
				if len(channels) == 0 {
					channels = append(channels, ctx.Msg.ChannelID)
				}
				// Foreach Channel
				for _, channel := range channels {
					if config.DebugOutput {
						nameGuild := channel
						if chinfo, err := bot.State.Channel(channel); err == nil {
							nameGuild = getGuildName(chinfo.GuildID)
						}
						nameCategory := getChannelCategoryName(channel)
						nameChannel := getChannelName(channel)
						nameDisplay := fmt.Sprintf("%s / %s", nameGuild, nameChannel)
						if nameCategory != "unknown" {
							nameDisplay = fmt.Sprintf("%s / %s / %s", nameGuild, nameCategory, nameChannel)
						}
						log.Println(lg("Command", "History", color.HiMagentaString,
							"Queueing history job for \"%s\" (%s)...", nameDisplay, channel))
					}
					if !isBotAdmin(ctx.Msg) {
						log.Println(lg("Command", "History", color.CyanString,
							"%s tried to cache history for %s but lacked proper permission.",
							getUserIdentifier(*ctx.Msg.Author), channel))
						if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
							log.Println(lg("Command", "History", color.HiRedString, fmtBotSendPerm, channel))
						} else {
							if _, err := replyEmbed(ctx.Msg, "Command — History", cmderrLackingBotAdminPerms); err != nil {
								log.Println(lg("Command", "History", color.HiRedString, cmderrSendFailure,
									getUserIdentifier(*ctx.Msg.Author), err))
							}
						}
					} else {
						// Run
						if !shouldAbort {
							if job, exists := historyJobs[channel]; !exists ||
								(job.Status != historyStatusDownloading && job.Status != historyStatusAbortRequested) {
								job.Status = historyStatusWaiting
								job.OriginChannel = ctx.Msg.ChannelID
								job.OriginUser = getUserIdentifier(*ctx.Msg.Author)
								job.TargetCommandingMessage = ctx.Msg
								job.TargetChannelID = channel
								job.TargetBefore = beforeID
								job.TargetSince = sinceID
								job.Updated = time.Now()
								job.Added = time.Now()
								historyJobs[channel] = job
							} else { // ALREADY RUNNING
								log.Println(lg("Command", "History", color.CyanString,
									"%s tried using history command but history is already running for %s...",
									getUserIdentifier(*ctx.Msg.Author), channel))
							}
						} else if historyJobs[channel].Status == historyStatusDownloading ||
							historyJobs[channel].Status == historyStatusWaiting { // requested abort while downloading
							if job, exists := historyJobs[channel]; exists {
								job.Status = historyStatusAbortRequested
								if historyJobs[channel].Status == historyStatusWaiting {
									job.Status = historyStatusAbortCompleted
								}
								historyJobs[channel] = job
							}
							log.Println(lg("Command", "History", color.CyanString,
								"%s cancelled history cataloging for \"%s\"",
								getUserIdentifier(*ctx.Msg.Author), channel))
						} else { // tried to stop but is not downloading
							log.Println(lg("Command", "History", color.CyanString,
								"%s tried to cancel history for \"%s\" but it's not running",
								getUserIdentifier(*ctx.Msg.Author), channel))
						}
					}
				}
			}
			//#endregion
		}
	}).Cat("Admin").Alias("catalog", "cache").Desc("Catalogs history for this channel")

	go router.On("exit", func(ctx *exrouter.Context) {
		if isCommandableChannel(ctx.Msg) {
			if isBotAdmin(ctx.Msg) {
				if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					log.Println(lg("Command", "Exit", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
				} else {
					if _, err := replyEmbed(ctx.Msg, "Command — Exit", "Exiting program..."); err != nil {
						log.Println(lg("Command", "Exit", color.HiRedString,
							cmderrSendFailure, getUserIdentifier(*ctx.Msg.Author), err))
					}
				}
				log.Println(lg("Command", "Exit", color.HiCyanString,
					"%s (bot admin) requested exit, goodbye...",
					getUserIdentifier(*ctx.Msg.Author)))
				properExit()
			} else {
				if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					log.Println(lg("Command", "Exit", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
				} else {
					if _, err := replyEmbed(ctx.Msg, "Command — Exit", cmderrLackingBotAdminPerms); err != nil {
						log.Println(lg("Command", "Exit", color.HiRedString,
							cmderrSendFailure, getUserIdentifier(*ctx.Msg.Author), err))
					}
				}
				log.Println(lg("Command", "Exit", color.HiCyanString,
					"%s tried to exit but lacked bot admin perms.", getUserIdentifier(*ctx.Msg.Author)))
			}
		}
	}).Cat("Admin").Alias("reload", "kill").Desc("Kills the bot")

	go router.On("emojis", func(ctx *exrouter.Context) {
		if isCommandableChannel(ctx.Msg) {
			if isBotAdmin(ctx.Msg) {
				if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					// Determine which guild(s)
					guilds := []string{ctx.Msg.GuildID}        // default to origin
					if args := ctx.Args.After(1); args != "" { // specifics
						guilds = nil
						_guilds := strings.Split(args, ",")
						if len(_guilds) > 0 {
							for _, guild := range _guilds {
								guild = strings.TrimSpace(guild)
								guilds = append(guilds, guild)
							}
						}
					}
					for _, guild := range guilds {
						i := 0
						s := 0

						guildName := guild
						guildNameO := guild
						if guildInfo, err := bot.Guild(guild); err == nil {
							guildName = sanitize.Name(guildInfo.Name)
							guildNameO = guildInfo.Name
						}

						destination := "emojis" + string(os.PathSeparator) + guildName + string(os.PathSeparator)
						if err = os.MkdirAll(destination, 0755); err != nil {
							log.Println(lg("Command", "Emojis", color.HiRedString, "Error while creating destination folder \"%s\": %s", destination, err))
						} else {
							emojis, err := bot.GuildEmojis(guild)
							if err != nil {
								log.Println(lg("Command", "Emojis", color.HiRedString, "Failed to get server emojis:\t%s", err))
							} else {
								for _, emoji := range emojis {
									var message discordgo.Message
									message.ChannelID = ctx.Msg.ChannelID
									url := "https://cdn.discordapp.com/emojis/" + emoji.ID

									status := handleDownload(
										downloadRequestStruct{
											InputURL:   url,
											Filename:   emoji.ID,
											Path:       destination,
											Message:    &message,
											FileTime:   time.Now(),
											HistoryCmd: false,
											EmojiCmd:   true,
											StartTime:  time.Now(),
										})

									if status.Status == downloadSuccess {
										i++
									} else {
										s++
										log.Println(lg("Command", "Emojis", color.HiRedString,
											"Failed to download emoji \"%s\": \t[%d - %s] %v",
											url, status.Status, getDownloadStatusString(status.Status), status.Error))
									}
								}
								destinationOut := destination
								abs, err := filepath.Abs(destination)
								if err == nil {
									destinationOut = abs
								}
								_, err = replyEmbed(ctx.Msg, "Command — Emojis",
									fmt.Sprintf("`%d` emojis downloaded, `%d` skipped or failed\n• Destination: `%s`\n• Server: `%s`",
										i, s, destinationOut, guildNameO,
									),
								)
								if err != nil {
									log.Println(lg("Command", "Emojis", color.HiRedString,
										"Failed to send status message for emoji downloads:\t%s", err))
								}
							}
						}
					}
				}
			} else {
				if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					log.Println(lg("Command", "Emojis", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
				} else {
					if _, err := replyEmbed(ctx.Msg, "Command — Emojis", cmderrLackingBotAdminPerms); err != nil {
						log.Println(lg("Command", "Emojis", color.HiRedString, cmderrSendFailure, getUserIdentifier(*ctx.Msg.Author), err))
					}
				}
				log.Println(lg("Command", "Emojis", color.HiCyanString,
					"%s tried to download emojis but lacked bot admin perms.", getUserIdentifier(*ctx.Msg.Author)))
			}
		}
	}).Cat("Admin").Desc("Saves all server emojis to download destination")

	//#endregion

	// Handler for Command Router
	go bot.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		//NOTE: This setup makes it case-insensitive but message content will be lowercase, currently case sensitivity is not necessary.
		router.FindAndExecute(bot, strings.ToLower(config.CommandPrefix), bot.State.User.ID, messageToLower(m.Message))
	})

	return router
}
