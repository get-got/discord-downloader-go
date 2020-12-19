package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"syscall"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

func commandHandler() {
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
		} else {
			log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
		}
	}).Cat("Info").Alias("info").Desc("Displays info regarding the current status of the bot")

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

	//#endregion

	//#region Admin Commands

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
						if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
							_, err := replyEmbed(ctx.Msg, "Command — History", cmderrHistoryCancelled)
							if err != nil {
								log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
							}
						} else {
							log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, channel))
						}
						log.Println(logPrefixHere, color.CyanString("%s cancelled history cataloging for %s", getUserIdentifier(*ctx.Msg.Author), channel))
					} else { // Start Local
						_, historyCommandIsSet := historyCommandActive[channel]
						if !historyCommandIsSet || historyCommandActive[channel] == "" {
							handleHistory(ctx.Msg, channel)
						} else {
							log.Println(logPrefixHere, color.CyanString("%s tried using history command but history is already running for %s...", getUserIdentifier(*ctx.Msg.Author), channel))
						}
					}
				} else {
					if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
						_, err := replyEmbed(ctx.Msg, "Command — History", cmderrLackingLocalAdminPerms)
						if err != nil {
							log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
						}
					} else {
						log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, channel))
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
								if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
									_, err := replyEmbed(ctx.Msg, "Command — History", cmderrHistoryCancelled)
									if err != nil {
										log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
									}
								} else {
									log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, channel))
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
									handleHistory(ctx.Msg, channelValue)
								} else {
									log.Println(logPrefixHere, color.CyanString("Tried using history command but history is already running for %s...", channelValue))
								}
							} else {
								if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
									replyEmbed(ctx.Msg, "Command — History", cmderrChannelNotRegistered)
								} else {
									log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, channel))
								}
								log.Println(logPrefixHere, color.CyanString("%s tried to cache history for %s but channel is not registered...", getUserIdentifier(*ctx.Msg.Author), channelValue))
							}
						}
					}
				} else {
					if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
						_, err := replyEmbed(ctx.Msg, "Command — History", "Please enter valid channel ID(s)...\n\n_Ex:_ ``<prefix>history <id1>,<id2>,<id3>``")
						if err != nil {
							log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
						}
					} else {
						log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, channel))
					}
					log.Println(logPrefixHere, color.CyanString("%s tried to cache history but input no channels", getUserIdentifier(*ctx.Msg.Author)))
				}
			} else {
				if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					_, err := replyEmbed(ctx.Msg, "Command — History", cmderrLackingBotAdminPerms)
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
					}
				} else {
					log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, channel))
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
				if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
					_, err := replyEmbed(ctx.Msg, "Command — Exit", "Exiting...")
					if err != nil {
						log.Println(logPrefixHere, color.HiRedString("Failed to send command embed message (requested by %s)...\t%s", getUserIdentifier(*ctx.Msg.Author), err))
					}
				} else {
					log.Println(logPrefixHere, color.HiRedString(fmtBotSendPerm, ctx.Msg.ChannelID))
				}
				log.Println(logPrefixHere, color.HiCyanString("%s requested exit, goodbye...", getUserIdentifier(*ctx.Msg.Author)))
				loop <- syscall.SIGINT
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

	//#endregion

	// Handler for Command Router
	bot.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		//NOTE: This setup makes it case-insensitive but message content will be lowercase, currently case sensitivity is not necessary.
		router.FindAndExecute(bot, strings.ToLower(config.CommandPrefix), bot.State.User.ID, messageToLower(m.Message))
	})
}
