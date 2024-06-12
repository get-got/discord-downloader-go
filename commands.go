package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

// TODO: Implement this for more?
const (
	cmderrLackingBotAdminPerms = "You do not have permission to use this command. Your User ID must be set as a bot administrator in the settings file."
	cmderrSendFailure          = "Failed to send command message (requested by %s)...\t%s"
)

// safe = logs errors
func safeReply(ctx *exrouter.Context, content string) bool {
	if hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
		if _, err := ctx.Reply(content); err != nil {
			log.Println(lg("Command", "", color.HiRedString, cmderrSendFailure, getUserIdentifier(*ctx.Msg.Author), err))
			return false
		} else {
			return true
		}
	} else {
		log.Println(lg("Command", "", color.HiRedString, fmtBotSendPerm, ctx.Msg.ChannelID))
		return false
	}
}

// TODO: function for handling perm error messages, etc etc to reduce clutter
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
					if !config.CommandTagging { // Erase mention if tagging disabled
						mention = ""
					}
					content := fmt.Sprintf("**Latency:** ``%dms`` — **Roundtrip:** ``%dms``",
						latency,
						roundtrip,
					)
					if pong != nil {
						if selfbot {
							if mention != "" { // Add space if mentioning
								mention += " "
							}
							bot.ChannelMessageEdit(pong.ChannelID, pong.ID, fmt.Sprintf("%s**Command — Ping**\n\n%s", mention, content))
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
					timeSince(startTime),
					startTime.Format("03:04:05pm on Monday, January 2, 2006 (MST)"),
					len(bot.State.Guilds),
					getBoundChannelsCount(),
					getBoundCategoriesCount(),
					getBoundServersCount(),
					getBoundUsersCount(),
					len(config.AdminChannels),
					bot.HeartbeatLatency().Milliseconds(),
				)
				if sourceConfig := getSource(ctx.Msg); sourceConfig != emptySourceConfig {
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
				if sourceConfig := getSource(ctx.Msg); sourceConfig != emptySourceConfig {
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

	//#endregion

	//#region Admin Commands

	go router.On("history", func(ctx *exrouter.Context) {
		if isCommandableChannel(ctx.Msg) {
			// Vars
			var all = false
			var channels []string

			var shouldAbort bool = false
			var shouldProcess bool = true
			var shouldWipeDB bool = false
			var shouldWipeCache bool = false

			var before string
			var beforeID string
			var since string
			var sinceID string

			if len(bot.State.Guilds) == 0 {
				log.Println(lg("Command", "History", color.HiRedString, "WARNING: Something is wrong with your Discord cache. This can result in missed channels..."))
			}

			//#region Parse Args
			for argKey, argValue := range ctx.Args {
				if argKey == 0 { // skip head
					continue
				}
				//SUBCOMMAND: cancel
				if strings.Contains(strings.ToLower(argValue), "cancel") ||
					strings.Contains(strings.ToLower(argValue), "stop") {
					shouldAbort = true
				} else if strings.Contains(strings.ToLower(argValue), "dbwipe") ||
					strings.Contains(strings.ToLower(argValue), "wipedb") { //SUBCOMMAND: dbwipe
					shouldProcess = false
					shouldWipeDB = true
				} else if strings.Contains(strings.ToLower(argValue), "cachewipe") ||
					strings.Contains(strings.ToLower(argValue), "wipecache") { //SUBCOMMAND: cachewipe
					shouldProcess = false
					shouldWipeCache = true
				} else if strings.Contains(strings.ToLower(argValue), "help") ||
					strings.Contains(strings.ToLower(argValue), "info") { //SUBCOMMAND: help
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
					strings.Contains(strings.ToLower(argValue), "output") { //SUBCOMMAND: list
					shouldProcess = false
					//MARKER: history jobs list

					// 1st
					output := fmt.Sprintf("**CURRENT HISTORY JOBS** ~ `%d total, %d running",
						historyJobCnt, historyJobCntRunning)
					outputC := fmt.Sprintf("CURRENT HISTORY JOBS ~ %d total, %d running",
						historyJobCnt, historyJobCntRunning)
					if historyJobCntCompleted > 0 {
						t := fmt.Sprintf(", %d completed", historyJobCntCompleted)
						output += t
						outputC += t
					}
					if historyJobCntWaiting > 0 {
						t := fmt.Sprintf(", %d waiting", historyJobCntWaiting)
						output += t
						outputC += t
					}
					if historyJobCntAborted > 0 {
						t := fmt.Sprintf(", %d cancelled", historyJobCntAborted)
						output += t
						outputC += t
					}
					if historyJobCntErrored > 0 {
						t := fmt.Sprintf(", %d failed", historyJobCntErrored)
						output += t
						outputC += t
					}
					safeReply(ctx, output+"`")
					log.Println(lg("Command", "History", color.HiCyanString, outputC))

					// Following
					output = ""
					for pair := historyJobs.Oldest(); pair != nil; pair = pair.Next() {
						channelID := pair.Key
						job := pair.Value
						jobSourceName, jobChannelName := channelDisplay(channelID)

						newline := fmt.Sprintf("• _%s_ (%s) `%s - %s`, `updated %s ago, added %s ago`\n",
							historyStatusLabel(job.Status), job.OriginUser, jobSourceName, jobChannelName,
							timeSinceShort(job.Updated),
							timeSinceShort(job.Added))
					redothismath: // bad way but dont care right now
						if len(output)+len(newline) > limitMsg {
							// send batch
							safeReply(ctx, output)
							output = ""
							goto redothismath
						}
						output += newline
						log.Println(lg("Command", "History", color.HiCyanString,
							fmt.Sprintf("%s (%s) %s - %s, updated %s ago, added %s ago",
								historyStatusLabel(job.Status), job.OriginUser, jobSourceName, jobChannelName,
								timeSinceShort(job.Updated),
								timeSinceShort(job.Added)))) // no batching
					}
					// finish off
					if output != "" {
						safeReply(ctx, output)
					}
					// done
					log.Println(lg("Command", "History", color.HiRedString, "%s requested statuses of history jobs.",
						getUserIdentifier(*ctx.Msg.Author)))
				} else if strings.Contains(strings.ToLower(argValue), "--before=") { // before key
					before = strings.ReplaceAll(strings.ToLower(argValue), "--before=", "")
					if isDate(before) {
						beforeID = discordTimestampToSnowflake("2006-01-02", before)
					} else if isNumeric(before) {
						beforeID = before
					} else { // try to parse duration
						dur, err := time.ParseDuration(before)
						if err == nil {
							beforeID = discordTimestampToSnowflake("2006-01-02 15:04:05.999999999 -0700 MST", time.Now().Add(-dur).Format("2006-01-02 15:04:05.999999999 -0700 MST"))
						}
					}
					if config.Debug && beforeID != "" {
						log.Println(lg("Command", "History", color.CyanString, "Date before range applied, snowflake %s, converts back to %s",
							beforeID, discordSnowflakeToTimestamp(beforeID, "2006-01-02T15:04:05.000Z07:00")))
					}
				} else if strings.Contains(strings.ToLower(argValue), "--since=") { //  since key
					since = strings.ReplaceAll(strings.ToLower(argValue), "--since=", "")
					if isDate(since) {
						sinceID = discordTimestampToSnowflake("2006-01-02", since)
					} else if isNumeric(since) {
						sinceID = since
					} else { // try to parse duration
						dur, err := time.ParseDuration(since)
						if err == nil {
							sinceID = discordTimestampToSnowflake("2006-01-02 15:04:05.999999999 -0700 MST", time.Now().Add(-dur).Format("2006-01-02 15:04:05.999999999 -0700 MST"))
						}
					}
					if config.Debug && sinceID != "" {
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
							if err != nil {
								guild, err = bot.Guild(target)
							}
							if err == nil {
								if config.Debug {
									log.Println(lg("Command", "History", color.YellowString,
										"Specified target %s is a guild: \"%s\", adding all channels...",
										target, guild.Name))
								}
								for _, ch := range guild.Channels {
									if ch.Type != discordgo.ChannelTypeGuildCategory &&
										ch.Type != discordgo.ChannelTypeGuildStageVoice &&
										ch.Type != discordgo.ChannelTypeGuildVoice {
										channels = append(channels, ch.ID)
										if config.Debug {
											log.Println(lg("Command", "History", color.YellowString,
												"Added %s (#%s in \"%s\") to history queue",
												ch.ID, ch.Name, guild.Name))
										}
									}
								}
							} else { // Test/Use if number is channel or category
								ch, err := bot.State.Channel(target)
								if err != nil {
									ch, err = bot.Channel(target)
								}
								if err == nil {
									if ch.Type == discordgo.ChannelTypeGuildCategory {
										// Category
										for _, guild := range bot.State.Guilds {
											for _, ch := range guild.Channels {
												if ch.ParentID == target {
													channels = append(channels, ch.ID)
													if config.Debug {
														log.Println(lg("Command", "History", color.YellowString, "Added %s (#%s in %s) to history queue",
															ch.ID, ch.Name, ch.GuildID))
													}
												}
											}
										}
									} else { // Standard Channel
										channels = append(channels, target)
										if config.Debug {
											log.Println(lg("Command", "History", color.YellowString, "Added %s (#%s in %s) to history queue",
												ch.ID, ch.Name, ch.GuildID))
										}
									}
								} else if config.Debug {
									log.Println(lg("Command", "History", color.HiRedString, "All attempts to identify target \"%s\" have failed...",
										target))
								}
							}
						} else if strings.Contains(strings.ToLower(target), "all") {
							for _, channel := range getAllRegisteredChannels() {
								channels = append(channels, channel.ChannelID)
							}
							all = true
						} else { // Aliasing
							for _, channel := range getAllRegisteredChannels() {
								if channel.Source.Aliases != nil {
									for _, alias := range *channel.Source.Aliases {
										if alias != "" && alias != " " {
											if strings.EqualFold(alias, target) {
												channels = append(channels, channel.ChannelID)
												if config.Debug {
													log.Println(lg("Command", "History", color.YellowString,
														"Added %s to history queue by alias \"%s\"",
														channel.ChannelID, alias))
												}
												break
											}
										}
									}
								} else if channel.Source.Alias != nil {
									alias := *channel.Source.Alias
									if alias != "" && alias != " " {
										if strings.EqualFold(alias, target) {
											channels = append(channels, channel.ChannelID)
											if config.Debug {
												log.Println(lg("Command", "History", color.YellowString,
													"Added %s to history queue by alias \"%s\"",
													channel.ChannelID, alias))
											}
										}
									}
								}
							}
						}
					}
				}
			}
			//#endregion

			// Local
			if len(channels) == 0 {
				channels = append(channels, ctx.Msg.ChannelID)
			}
			// Foreach Channel
			for _, channel := range channels {
				//#region Process Channels
				if shouldProcess && config.Debug {
					nameGuild := channel
					chinfo, err := bot.State.Channel(channel)
					if err != nil {
						chinfo, err = bot.Channel(channel)
					}
					if err == nil {
						nameGuild = getServerLabel(chinfo.GuildID)
					}
					nameCategory := getCategoryLabel(channel)
					nameChannel := getChannelLabel(channel, nil)
					nameDisplay := fmt.Sprintf("%s / #%s", nameGuild, nameChannel)
					if nameCategory != "Category" {
						nameDisplay = fmt.Sprintf("%s / %s / #%s", nameGuild, nameCategory, nameChannel)
					}
					log.Println(lg("Command", "History", color.HiMagentaString,
						"Queueing history job for \"%s\"\t\t(%s) ...", nameDisplay, channel))
				}
				if !isBotAdmin(ctx.Msg) {
					log.Println(lg("Command", "History", color.CyanString,
						"%s tried to handle history for %s but lacked proper permission.",
						getUserIdentifier(*ctx.Msg.Author), channel))
					if !hasPerms(ctx.Msg.ChannelID, discordgo.PermissionSendMessages) {
						log.Println(lg("Command", "History", color.HiRedString, fmtBotSendPerm, channel))
					} else {
						if _, err := replyEmbed(ctx.Msg, "Command — History", cmderrLackingBotAdminPerms); err != nil {
							log.Println(lg("Command", "History", color.HiRedString, cmderrSendFailure,
								getUserIdentifier(*ctx.Msg.Author), err))
						}
					}
				} else { // IS BOT ADMIN
					if shouldProcess { // PROCESS TREE; MARKER: history queue via cmd
						if shouldAbort { // ABORT
							if job, exists := historyJobs.Get(channel); exists &&
								(job.Status == historyStatusRunning || job.Status == historyStatusWaiting) {
								// DOWNLOADING, ABORTING
								job.Status = historyStatusAbortRequested
								if job.Status == historyStatusWaiting {
									job.Status = historyStatusAbortCompleted
								}
								historyJobs.Set(channel, job)
								log.Println(lg("Command", "History", color.CyanString,
									"%s cancelled history cataloging for \"%s\"",
									getUserIdentifier(*ctx.Msg.Author), channel))
							} else { // NOT DOWNLOADING, ABORTING
								log.Println(lg("Command", "History", color.CyanString,
									"%s tried to cancel history for \"%s\" but it's not running",
									getUserIdentifier(*ctx.Msg.Author), channel))
							}
						} else { // RUN
							if job, exists := historyJobs.Get(channel); !exists ||
								(job.Status != historyStatusRunning && job.Status != historyStatusAbortRequested) {
								job.Status = historyStatusWaiting
								job.OriginChannel = ctx.Msg.ChannelID
								job.OriginUser = getUserIdentifier(*ctx.Msg.Author)
								job.TargetCommandingMessage = ctx.Msg
								job.TargetChannelID = channel
								job.TargetBefore = beforeID
								job.TargetSince = sinceID
								job.Updated = time.Now()
								job.Added = time.Now()
								historyJobs.Set(channel, job)
							} else { // ALREADY RUNNING
								log.Println(lg("Command", "History", color.CyanString,
									"%s tried using history command but history is already running for %s...",
									getUserIdentifier(*ctx.Msg.Author), channel))
							}
						}
					}
					if shouldWipeDB {
						if all {
							myDB.Close()
							time.Sleep(1 * time.Second)
							if _, err := os.Stat(pathDatabaseBase); err == nil {
								err = os.RemoveAll(pathDatabaseBase)
								if err != nil {
									log.Println(lg("Command", "History", color.HiRedString,
										"Encountered error deleting database folder:\t%s", err))
								} else {
									log.Println(lg("Command", "History", color.HiGreenString,
										"Deleted database."))
								}
								time.Sleep(1 * time.Second)
								mainWg.Add(1)
								go openDatabase()
								break
							} else {
								log.Println(lg("Command", "History", color.HiRedString,
									"Database folder inaccessible:\t%s", err))
							}
						} else {
							dbDeleteByChannelID(channel)
						}
					}
					if shouldWipeCache {
						if all {
							if _, err := os.Stat(pathCacheHistory); err == nil {
								err = os.RemoveAll(pathCacheHistory)
								if err != nil {
									log.Println(lg("Command", "History", color.HiRedString,
										"Encountered error deleting database folder:\t%s", err))
								} else {
									log.Println(lg("Command", "History", color.HiGreenString,
										"Deleted database."))
									break
								}
							} else {
								log.Println(lg("Command", "History", color.HiRedString,
									"Cache folder inaccessible:\t%s", err))
							}
						} else {
							fp := pathCacheHistory + string(os.PathSeparator) + channel + ".json"
							if _, err := os.Stat(fp); err == nil {
								err = os.RemoveAll(fp)
								if err != nil {
									log.Println(lg("Debug", "History", color.HiRedString,
										"Encountered error deleting cache file for %s:\t%s", channel, err))
								} else {
									log.Println(lg("Debug", "History", color.HiGreenString,
										"Deleted cache file for %s.", channel))
								}
							} else {
								log.Println(lg("Command", "History", color.HiRedString,
									"Cache folder inaccessible:\t%s", err))
							}
						}
					}
				}
				//#endregion
			}
			if shouldWipeDB {
				cachedDownloadID = dbDownloadCount()
			}
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

	//#endregion

	// Handler for Command Router
	go bot.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {

		// Override Prefix per-Source
		prefix := config.CommandPrefix
		if _, err := getChannel(m.ChannelID); err == nil {
			if messageConfig := getSource(m.Message); messageConfig != emptySourceConfig {
				if messageConfig.CommandPrefix != nil {
					prefix = *messageConfig.CommandPrefix
				}
			} else {
				if messageAdminConfig := getAdminChannelConfig(m.Message.ChannelID); messageAdminConfig != emptyAdminChannelConfig {
					if messageAdminConfig.CommandPrefix != nil {
						prefix = *messageAdminConfig.CommandPrefix
					}
				}
			}
		}

		//NOTE: This setup makes it case-insensitive but message content will be lowercase, currently case sensitivity is not necessary.
		router.FindAndExecute(bot, strings.ToLower(prefix), bot.State.User.ID, messageToLower(m.Message))
	})

	return router
}
