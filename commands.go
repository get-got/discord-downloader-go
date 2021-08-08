package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/kennygrant/sanitize"
)

// Multiple use messages to save space and make cleaner.
//TODO: Implement this for more?
const (
	cmderrLackingLocalAdminPerms = "You do not have permission to use this command.\n" +
		"\nTo use this command you must:" +
		"\n• Be set as a bot administrator (in the settings)" +
		"\n• Own this Discord Server" +
		"\n• Have Server Administrator Permissions"
	cmderrLackingBotAdminPerms = "You do not have permission to use this command. Your User ID must be set as a bot administrator in the settings file."
	cmderrChannelNotRegistered = "Specified channel is not registered in the bot settings."
	cmderrHistoryCancelled     = "History cataloging was cancelled."
)

func handleCommands() *exrouter.Route {
	router := exrouter.New()

	//#region Utility Commands

	router.On("ping", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:ping]")
		if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
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
		} else {
			log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
		}
	}).Cat("Utility").Alias("test").Desc("Pings the bot")

	router.On("help", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:help]")
		if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
			if isGlobalCommandAllowed(ctx.Msg) {
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
		} else {
			log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
		}
	}).Cat("Utility").Alias("commands").Desc("Outputs this help menu")

	//#endregion

	//#region Info Commands

	router.On("status", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:status]")
		if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
			if isCommandableChannel(ctx.Msg) {
				message := fmt.Sprintf("• **Uptime —** %s\n"+
					"• **Started at —** %s\n"+
					"• **Joined Servers —** %d\n"+
					"• **Bound Channels —** %d\n"+
					"• **Bound Servers —** %d\n"+
					"• **Admin Channels —** %d\n"+
					"• **Heartbeat Latency —** %dms",
					durafmt.Parse(time.Since(startTime)).String(),
					startTime.Format("03:04:05pm on Monday, January 2, 2006 (MST)"),
					len(bot.State.Guilds),
					getBoundChannelsCount(),
					getBoundServersCount(),
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
		} else {
			log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
		}
	}).Cat("Info").Desc("Displays info regarding the current status of the bot")

	router.On("stats", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:stats]")
		if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
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
		} else {
			log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
		}
	}).Cat("Info").Desc("Outputs statistics regarding this channel")

	router.On("info", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:info]")
		if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
			if isGlobalCommandAllowed(ctx.Msg) {
				content := fmt.Sprintf("Here is some useful info...\n\n"+
					"• **Your User ID —** `%s`\n"+
					"• **Bots User ID —** `%s`\n"+
					"• **This Channel ID —** `%s`\n"+
					"• **This Server ID —** `%s`"+
					"\n\nRemember to remove any spaces when copying to settings.",
					ctx.Msg.Author.ID, user.ID, ctx.Msg.ChannelID, ctx.Msg.GuildID)
				_, err := replyEmbed(ctx.Msg, "Command — Info", content)
				// Failed to send
				if err != nil {
					log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s requested info", getUserIdentifier(*ctx.Msg.Author)))
			} else {
				log.Println(logPrefixHere, color.HiRedString("%s tried using the info command but commands are disabled here", getUserIdentifier(*ctx.Msg.Author)))
			}
		} else {
			log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
		}
	}).Cat("Info").Desc("Displays info regarding Discord IDs")

	//#endregion

	//#region Admin Commands

	router.On("history", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:history]")
		// Vars
		var channels []string
		var before string
		var beforeID string
		var since string
		var sinceID string
		var stop bool
		// Keys
		beforeKey := "--before="
		sinceKey := "--since="
		// Parse Args
		for k, v := range ctx.Args {
			// Skip "history" segment
			if k == 0 {
				continue
			}
			// Actually Parse Args
			if strings.Contains(strings.ToLower(v), beforeKey) {
				before = strings.ReplaceAll(strings.ToLower(v), beforeKey, "")
				if isDate(before) {
					beforeID = discordTimestampToSnowflake("2006-01-02", before)
				} else if isNumeric(before) {
					beforeID = before
				}
				if config.DebugOutput {
					log.Println(logPrefixDebug, logPrefixHere, color.CyanString("Date range applied, before %s", beforeID))
				}
			} else if strings.Contains(strings.ToLower(v), sinceKey) {
				since = strings.ReplaceAll(strings.ToLower(v), sinceKey, "")
				if isDate(since) {
					sinceID = discordTimestampToSnowflake("2006-01-02", since)
				} else if isNumeric(since) {
					sinceID = since
				}
				if config.DebugOutput {
					log.Println(logPrefixDebug, logPrefixHere, color.CyanString("Date range applied, since %s", sinceID))
				}
			} else if strings.Contains(strings.ToLower(v), "cancel") || strings.Contains(strings.ToLower(v), "stop") {
				stop = true
			} else {
				// Actual Source ID(s)
				targets := strings.Split(ctx.Args.Get(k), ",")
				for _, target := range targets {
					if isNumeric(target) {
						// Test/Use if number is guild
						guild, err := bot.State.Guild(target)
						if err == nil {
							if config.DebugOutput {
								log.Println(logPrefixHere, logPrefixDebug, color.YellowString("Specified target %s is a guild: \"%s\", adding all channels...", target, guild.Name))
							}
							for _, ch := range guild.Channels {
								channels = append(channels, ch.ID)
								if config.DebugOutput {
									log.Println(logPrefixHere, logPrefixDebug, color.YellowString("Added %s (#%s in \"%s\") to history queue", ch.ID, ch.Name, guild.Name))
								}
							}
						} else { // Test/Use if number is channel
							ch, err := bot.State.Channel(target)
							if err == nil {
								channels = append(channels, target)
								if config.DebugOutput {
									log.Println(logPrefixHere, logPrefixDebug, color.YellowString("Added %s (#%s in %s) to history queue", ch.ID, ch.Name, ch.GuildID))
								}
							}
						}
					} else if strings.Contains(strings.ToLower(target), "all") {
						channels = getAllChannels()
					}
				}
			}
		}
		if len(channels) == 0 { // Local
			channels = append(channels, ctx.Msg.ChannelID)
		}
		// Foreach Channel
		for _, channel := range channels {
			if config.DebugOutput {
				log.Println(logPrefixHere, logPrefixDebug, color.YellowString("Processing %s...", channel))
			}
			// Registered check
			if isCommandableChannel(ctx.Msg) {
				// Permission check
				if isBotAdmin(ctx.Msg) || isLocalAdmin(ctx.Msg) {
					// Run
					if !stop {
						_, historyCommandIsSet := historyStatus[channel]
						if !historyCommandIsSet || historyStatus[channel] == "" {
							if config.AsynchronousHistory {
								go handleHistory(ctx.Msg, channel, beforeID, sinceID)
							} else {
								handleHistory(ctx.Msg, channel, beforeID, sinceID)
							}
						} else { // ALREADY RUNNING
							log.Println(logPrefixHere, color.CyanString("%s tried using history command but history is already running for %s...", getUserIdentifier(*ctx.Msg.Author), channel))
						}
					} else if historyStatus[channel] == "downloading" {
						historyStatus[channel] = "cancel"
						if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
							_, err := replyEmbed(ctx.Msg, "Command — History", cmderrHistoryCancelled)
							if err != nil {
								log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
							}
						} else {
							log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, channel))
						}
						log.Println(logPrefixHere, color.CyanString("%s cancelled history cataloging for \"%s\"", getUserIdentifier(*ctx.Msg.Author), channel))
					}
				} else { // DOES NOT HAVE PERMISSION
					if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
						_, err := replyEmbed(ctx.Msg, "Command — History", cmderrLackingLocalAdminPerms)
						if err != nil {
							log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
						}
					} else {
						log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, channel))
					}
					log.Println(logPrefixHere, color.CyanString("%s tried to cache history for %s but lacked proper permission.", getUserIdentifier(*ctx.Msg.Author), channel))
				}
			} else { // CHANNEL NOT REGISTERED
				log.Println(logPrefixHere, color.CyanString("%s tried to catalog history for \"%s\" but channel is not registered...", getUserIdentifier(*ctx.Msg.Author), channel))
			}
		}
	}).Alias("catalog", "cache").Cat("Admin").Desc("Catalogs history for this channel")

	router.On("exit", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:exit]")
		if isCommandableChannel(ctx.Msg) {
			if isBotAdmin(ctx.Msg) {
				if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					_, err := replyEmbed(ctx.Msg, "Command — Exit", "Exiting...")
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
					}
				} else {
					log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s (bot admin) requested exit, goodbye...", getUserIdentifier(*ctx.Msg.Author)))
				properExit()
			} else {
				if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					_, err := replyEmbed(ctx.Msg, "Command — Exit", cmderrLackingBotAdminPerms)
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
					}
				} else {
					log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s tried to exit but lacked bot admin perms.", getUserIdentifier(*ctx.Msg.Author)))
			}
		}
	}).Alias("reload", "kill").Cat("Admin").Desc("Kills the bot")

	router.On("emojis", func(ctx *exrouter.Context) {
		logPrefixHere := color.CyanString("[dgrouter:emojis]")
		if isGlobalCommandAllowed(ctx.Msg) {
			if isBotAdmin(ctx.Msg) {
				if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					args := ctx.Args.After(1)

					// Determine which guild(s)
					guilds := []string{ctx.Msg.GuildID}
					if args != "" {
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
						guildInfo, err := bot.Guild(guild)
						if err == nil {
							guildName = sanitize.Name(guildInfo.Name)
							guildNameO = guildInfo.Name
						}

						destination := "emojis" + string(os.PathSeparator) + guildName + string(os.PathSeparator)

						err = os.MkdirAll(destination, 0755)
						if err != nil {
							log.Println(logPrefixHere, color.HiRedString("Error while creating destination folder \"%s\": %s", destination, err))
						} else {
							emojis, err := bot.GuildEmojis(guild)
							if err == nil {
								for _, emoji := range emojis {
									var message discordgo.Message
									message.ChannelID = ctx.Msg.ChannelID
									url := "https://cdn.discordapp.com/emojis/" + emoji.ID

									status := startDownload(
										url,
										emoji.ID,
										destination,
										&message,
										time.Now(),
										false, true)

									if status.Status == downloadSuccess {
										i++
									} else {
										s++
										log.Println(logPrefixHere, color.HiRedString("Failed to download emoji \"%s\": \t[%d - %s] %v", url, status.Status, getDownloadStatusString(status.Status), status.Error))
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
									log.Println(logPrefixHere, color.HiRedString("Failed to send status message for emoji downloads:\t%s", err))
								}
							} else {
								log.Println(err)
							}
						}
					}
				}
			} else {
				if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					_, err := replyEmbed(ctx.Msg, "Command — Emojis", cmderrLackingBotAdminPerms)
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
					}
				} else {
					log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s tried to download emojis but lacked bot admin perms.", getUserIdentifier(*ctx.Msg.Author)))
			}
		}
	}).Cat("Admin").Desc("Saves all server emojis to download destination")

	//#endregion

	// Handler for Command Router
	bot.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		//NOTE: This setup makes it case-insensitive but message content will be lowercase, currently case sensitivity is not necessary.
		router.FindAndExecute(bot, strings.ToLower(config.CommandPrefix), bot.State.User.ID, messageToLower(m.Message))
	})

	return router
}
