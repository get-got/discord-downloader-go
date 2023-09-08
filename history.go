package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type historyStatus int

const (
	historyStatusWaiting historyStatus = iota
	historyStatusRunning
	historyStatusAbortRequested
	historyStatusAbortCompleted
	historyStatusErrorReadMessageHistoryPerms
	historyStatusErrorRequesting
	historyStatusCompletedNoMoreMessages
	historyStatusCompletedToBeforeFilter
	historyStatusCompletedToSinceFilter
)

func historyStatusLabel(status historyStatus) string {
	switch status {
	case historyStatusWaiting:
		return "Waiting..."
	case historyStatusRunning:
		return "Currently Downloading..."
	case historyStatusAbortRequested:
		return "Abort Requested..."
	case historyStatusAbortCompleted:
		return "Aborted..."
	case historyStatusErrorReadMessageHistoryPerms:
		return "ERROR: Cannot Read Message History"
	case historyStatusErrorRequesting:
		return "ERROR: Message Requests Failed"
	case historyStatusCompletedNoMoreMessages:
		return "COMPLETE: No More Messages"
	case historyStatusCompletedToBeforeFilter:
		return "COMPLETE: Exceeded Before Date Filter"
	case historyStatusCompletedToSinceFilter:
		return "COMPLETE: Exceeded Since Date Filter"
	default:
		return "Unknown"
	}
}

type historyJob struct {
	Status                  historyStatus
	OriginUser              string
	OriginChannel           string
	TargetCommandingMessage *discordgo.Message
	TargetChannelID         string
	TargetBefore            string
	TargetSince             string
	DownloadCount           int64
	DownloadSize            int64
	Updated                 time.Time
	Added                   time.Time
}

var (
	historyJobs            *orderedmap.OrderedMap[string, historyJob]
	historyJobCnt          int
	historyJobCntWaiting   int
	historyJobCntRunning   int
	historyJobCntAborted   int
	historyJobCntErrored   int
	historyJobCntCompleted int
)

// TODO: cleanup
type historyCache struct {
	Updated        time.Time
	Running        bool
	RunningBefore  string // messageID for last before range attempted if interrupted
	CompletedSince string // messageID for last message the bot has 100% assumed completion on (since start of channel)
}

func handleHistory(commandingMessage *discordgo.Message, subjectChannelID string, before string, since string) int {
	var err error

	historyStartTime := time.Now()

	var historyDownloadDuration time.Duration

	// Log Prefix
	var commander string = "AUTORUN"
	var autorun bool = true
	if commandingMessage != nil { // Only time commandingMessage is nil is Autorun
		commander = getUserIdentifier(*commandingMessage.Author)
		autorun = false
	}
	logPrefix := fmt.Sprintf("%s/%s: ", subjectChannelID, commander)

	// Skip Requested
	if job, exists := historyJobs.Get(subjectChannelID); exists && job.Status != historyStatusWaiting {
		log.Println(lg("History", "", color.RedString, logPrefix+"History job skipped, Status: %s", historyStatusLabel(job.Status)))
		return -1
	}

	// Vars
	var totalMessages int64 = 0
	var totalDownloads int64 = 0
	var totalFilesize int64 = 0
	var messageRequestCount int = 0
	var responseMsg *discordgo.Message = &discordgo.Message{} // dummy message
	responseMsg.ID = ""
	responseMsg.ChannelID = subjectChannelID
	responseMsg.GuildID = ""

	baseChannelInfo, err := bot.State.Channel(subjectChannelID)
	if err != nil {
		baseChannelInfo, err = bot.Channel(subjectChannelID)
		if err != nil {
			log.Println(lg("History", "", color.HiRedString, logPrefix+"Error fetching channel data from discordgo:\t%s", err))
		}
	}

	guildName := getServerLabel(baseChannelInfo.GuildID)
	categoryName := getCategoryLabel(baseChannelInfo.ID)

	subjectChannels := []discordgo.Channel{}

	// Check channel type
	baseChannelIsForum := true
	if baseChannelInfo.Type != discordgo.ChannelTypeGuildCategory &&
		baseChannelInfo.Type != discordgo.ChannelTypeGuildForum &&
		baseChannelInfo.Type != discordgo.ChannelTypeGuildNews &&
		baseChannelInfo.Type != discordgo.ChannelTypeGuildStageVoice &&
		baseChannelInfo.Type != discordgo.ChannelTypeGuildVoice &&
		baseChannelInfo.Type != discordgo.ChannelTypeGuildStore {
		subjectChannels = append(subjectChannels, *baseChannelInfo)
		baseChannelIsForum = false
	}

	// Index Threads
	now := time.Now()
	if threads, err := bot.ThreadsArchived(subjectChannelID, &now, 0); err == nil {
		for _, thread := range threads.Threads {
			subjectChannels = append(subjectChannels, *thread)
		}
	}
	if threads, err := bot.ThreadsActive(subjectChannelID); err == nil {
		for _, thread := range threads.Threads {
			subjectChannels = append(subjectChannels, *thread)
		}
	}

	// Send Status?
	var sendStatus bool = true
	if (autorun && !config.SendAutoHistoryStatus) || (!autorun && !config.SendHistoryStatus) {
		sendStatus = false
	}

	// Check Read History perms
	if !baseChannelIsForum && !hasPerms(subjectChannelID, discordgo.PermissionReadMessageHistory) {
		if job, exists := historyJobs.Get(subjectChannelID); exists {
			job.Status = historyStatusRunning
			job.Updated = time.Now()
			historyJobs.Set(subjectChannelID, job)
		}
		log.Println(lg("History", "", color.HiRedString, logPrefix+"BOT DOES NOT HAVE PERMISSION TO READ MESSAGE HISTORY!!!"))
	}

	// Update Job Status to Downloading
	if job, exists := historyJobs.Get(subjectChannelID); exists {
		job.Status = historyStatusRunning
		job.Updated = time.Now()
		historyJobs.Set(subjectChannelID, job)
	}

	//#region Cache Files

	openHistoryCache := func(channel string) historyCache {
		if f, err := os.ReadFile(pathCacheHistory + string(os.PathSeparator) + channel + ".json"); err == nil {
			var ret historyCache
			if err = json.Unmarshal(f, &ret); err != nil {
				log.Println(lg("Debug", "History", color.RedString,
					logPrefix+"Failed to unmarshal json for cache:\t%s", err))
			} else {
				return ret
			}
		}
		return historyCache{}
	}

	writeHistoryCache := func(channel string, cache historyCache) {
		cacheJson, err := json.Marshal(cache)
		if err != nil {
			log.Println(lg("Debug", "History", color.RedString,
				logPrefix+"Failed to format cache into json:\t%s", err))
		} else {
			if err := os.MkdirAll(pathCacheHistory, 0755); err != nil {
				log.Println(lg("Debug", "History", color.HiRedString,
					logPrefix+"Error while creating history cache folder \"%s\": %s", pathCacheHistory, err))
			}
			f, err := os.OpenFile(
				pathCacheHistory+string(os.PathSeparator)+channel+".json",
				os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				log.Println(lg("Debug", "History", color.RedString,
					logPrefix+"Failed to open cache file:\t%s", err))
			}
			if _, err = f.WriteString(string(cacheJson)); err != nil {
				log.Println(lg("Debug", "History", color.RedString,
					logPrefix+"Failed to write cache file:\t%s", err))
			} else if !autorun && config.Debug {
				log.Println(lg("Debug", "History", color.YellowString,
					logPrefix+"Wrote to cache file."))
			}
			f.Close()
		}
	}

	deleteHistoryCache := func(channel string) {
		fp := pathCacheHistory + string(os.PathSeparator) + channel + ".json"
		if _, err := os.Stat(fp); err == nil {
			err = os.Remove(fp)
			if err != nil {
				log.Println(lg("Debug", "History", color.HiRedString,
					logPrefix+"Encountered error deleting cache file:\t%s", err))
			} else if commandingMessage != nil && config.Debug {
				log.Println(lg("Debug", "History", color.HiRedString,
					logPrefix+"Deleted cache file."))
			}
		}
	}

	//#endregion

	for _, channel := range subjectChannels {
		logPrefix = fmt.Sprintf("%s/%s: ", channel.ID, commander)

		sourceConfig := getSource(responseMsg, &channel)

		// Invalid Source?
		if sourceConfig == emptyConfig {
			log.Println(lg("History", "", color.HiRedString,
				logPrefix+"Invalid source: "+channel.ID))
			if job, exists := historyJobs.Get(subjectChannelID); exists {
				job.Status = historyStatusErrorRequesting
				job.Updated = time.Now()
				historyJobs.Set(subjectChannelID, job)
			}
			continue
		} else { // Process

			// Overwrite Send Status
			if sourceConfig.SendAutoHistoryStatus != nil {
				if autorun && !*sourceConfig.SendAutoHistoryStatus {
					sendStatus = false
				}
			}
			if sourceConfig.SendHistoryStatus != nil {
				if !autorun && !*sourceConfig.SendHistoryStatus {
					sendStatus = false
				}
			}

			hasPermsToRespond := hasPerms(channel.ID, discordgo.PermissionSendMessages)
			if !autorun {
				hasPermsToRespond = hasPerms(commandingMessage.ChannelID, discordgo.PermissionSendMessages)
			}

			// Date Range Vars
			rangeContent := ""
			var beforeTime time.Time
			var beforeID = before
			var sinceID = ""

			// Handle Cache File
			if cache := openHistoryCache(channel.ID); cache != (historyCache{}) {
				if cache.CompletedSince != "" {
					if config.Debug {
						log.Println(lg("Debug", "History", color.GreenString,
							logPrefix+"Assuming history is completed prior to "+cache.CompletedSince))
					}
					since = cache.CompletedSince
				}
				if cache.Running {
					if config.Debug {
						log.Println(lg("Debug", "History", color.YellowString,
							logPrefix+"Job was interrupted last run, picking up from "+beforeID))
					}
					beforeID = cache.RunningBefore
				}
			}

			//#region Date Range Output

			var beforeRange = before
			if beforeRange != "" {
				if isDate(beforeRange) {
					beforeRange = discordTimestampToSnowflake(beforeID, "2006-01-02")
				}
				if isNumeric(beforeRange) {
					rangeContent += fmt.Sprintf("**Before:** `%s`\n", beforeRange)
				}
				before = beforeRange
			}

			var sinceRange = since
			if sinceRange != "" {
				if isDate(sinceRange) {
					sinceRange = discordTimestampToSnowflake(sinceRange, "2006-01-02")
				}
				if isNumeric(sinceRange) {
					rangeContent += fmt.Sprintf("**Since:** `%s`\n", sinceRange)
				}
				since = sinceRange
			}

			if rangeContent != "" {
				rangeContent += "\n"
			}

			//#endregion

			channelName := getChannelLabel(channel.ID, &channel)
			if channel.ParentID != "" {
				channelName = getChannelLabel(channel.ParentID, nil) + " \"" + getChannelLabel(channel.ID, &channel) + "\""
			}
			sourceName := fmt.Sprintf("%s / %s", guildName, channelName)
			msgSourceDisplay := fmt.Sprintf("`Server:` **%s**\n`Channel:` #%s", guildName, channelName)
			if categoryName != "unknown" {
				sourceName = fmt.Sprintf("%s / %s / %s", guildName, categoryName, channelName)
				msgSourceDisplay = fmt.Sprintf("`Server:` **%s**\n`Category:` _%s_\n`Channel:` #%s",
					guildName, categoryName, channelName)
			}

			// Initial Status Message
			if sendStatus {
				if hasPermsToRespond {
					responseMsg, err = replyEmbed(commandingMessage, "Command — History", msgSourceDisplay)
					if err != nil {
						log.Println(lg("History", "", color.HiRedString,
							logPrefix+"Failed to send command embed message:\t%s", err))
					}
				} else {
					log.Println(lg("History", "", color.HiRedString,
						logPrefix+fmtBotSendPerm, commandingMessage.ChannelID))
				}
			}
			log.Println(lg("History", "", color.HiCyanString, logPrefix+"Began checking history for \"%s\"...", sourceName))

			lastMessageID := ""
		MessageRequestingLoop:
			for {
				// Next 100
				if beforeTime != (time.Time{}) {
					messageRequestCount++
					writeHistoryCache(channel.ID, historyCache{
						Updated:       time.Now(),
						Running:       true,
						RunningBefore: beforeID,
					})

					// Update Status
					log.Println(lg("History", "", color.CyanString,
						logPrefix+"Requesting more, \t%d downloaded (%s), \t%d processed, \tsearching before %s ago (%s)",
						totalDownloads, humanize.Bytes(uint64(totalFilesize)), totalMessages, timeSinceShort(beforeTime), beforeTime.String()[:10]))
					if sendStatus {
						var status string
						if totalDownloads == 0 {
							status = fmt.Sprintf(
								"``%s:`` **No files downloaded...**\n"+
									"_%s messages processed, avg %d msg/s_\n\n"+
									"%s\n\n"+
									"%s`(%d)` _Processing more messages, please wait..._\n",
								timeSinceShort(historyStartTime),
								formatNumber(totalMessages), int(float64(totalMessages)/time.Since(historyStartTime).Seconds()),
								msgSourceDisplay, rangeContent, messageRequestCount,
							)
						} else {
							status = fmt.Sprintf(
								"``%s:`` **%s files downloaded...**\n`%s so far, avg %1.1f MB/s`\n"+
									"_%s messages processed, avg %d msg/s_\n\n"+
									"%s\n\n"+
									"%s`(%d)` _Processing more messages, please wait..._\n",
								timeSinceShort(historyStartTime), formatNumber(totalDownloads),
								humanize.Bytes(uint64(totalFilesize)), float64(totalFilesize/humanize.MByte)/historyDownloadDuration.Seconds(),
								formatNumber(totalMessages), int(float64(totalMessages)/time.Since(historyStartTime).Seconds()),
								msgSourceDisplay, rangeContent, messageRequestCount,
							)
						}
						if responseMsg == nil {
							log.Println(lg("History", "", color.RedString,
								logPrefix+"Tried to edit status message but it doesn't exist, sending new one."))
							if responseMsg, err = replyEmbed(responseMsg, "Command — History", status); err != nil { // Failed to Edit Status, Send New Message
								log.Println(lg("History", "", color.HiRedString,
									logPrefix+"Failed to send replacement status message:\t%s", err))
							}
						} else {
							if !hasPermsToRespond {
								log.Println(lg("History", "", color.HiRedString,
									logPrefix+fmtBotSendPerm+" - %s", responseMsg.ChannelID, status))
							} else {
								// Edit Status
								if selfbot {
									responseMsg, err = bot.ChannelMessageEdit(responseMsg.ChannelID, responseMsg.ID,
										fmt.Sprintf("**Command — History**\n\n%s", status))
								} else {
									responseMsg, err = bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
										ID:      responseMsg.ID,
										Channel: responseMsg.ChannelID,
										Embed:   buildEmbed(responseMsg.ChannelID, "Command — History", status),
									})
								}
								// Failed to Edit Status
								if err != nil {
									log.Println(lg("History", "", color.HiRedString,
										logPrefix+"Failed to edit status message, sending new one:\t%s", err))
									if responseMsg, err = replyEmbed(responseMsg, "Command — History", status); err != nil { // Failed to Edit Status, Send New Message
										log.Println(lg("History", "", color.HiRedString,
											logPrefix+"Failed to send replacement status message:\t%s", err))
									}
								}
							}
						}
					}

					// Update presence
					timeLastUpdated = time.Now()
					if *sourceConfig.PresenceEnabled {
						go updateDiscordPresence()
					}
				}

				// Request More Messages
				msg_rq_cnt := 0
			request_messages:
				msg_rq_cnt++
				if config.HistoryRequestDelay > 0 {
					log.Println(lg("History", "", color.YellowString, "Delaying next batch request for %d seconds...", config.HistoryRequestDelay))
					time.Sleep(time.Second * time.Duration(config.HistoryRequestDelay))
				}
				if messages, err := bot.ChannelMessages(channel.ID, config.HistoryRequestCount, beforeID, sinceID, ""); err != nil {
					// Error requesting messages
					if sendStatus {
						if !hasPermsToRespond {
							log.Println(lg("History", "", color.HiRedString,
								logPrefix+fmtBotSendPerm, responseMsg.ChannelID))
						} else {
							_, err = replyEmbed(responseMsg, "Command — History",
								fmt.Sprintf("Encountered an error requesting messages for %s: %s", channel.ID, err.Error()))
							if err != nil {
								log.Println(lg("History", "", color.HiRedString,
									logPrefix+"Failed to send error message:\t%s", err))
							}
						}
					}
					log.Println(lg("History", "", color.HiRedString, logPrefix+"Error requesting messages:\t%s", err))
					if job, exists := historyJobs.Get(subjectChannelID); exists {
						job.Status = historyStatusErrorRequesting
						job.Updated = time.Now()
						historyJobs.Set(subjectChannelID, job)
					}
					//TODO: delete cahce or handle it differently?
					break MessageRequestingLoop
				} else {
					// No More Messages
					if len(messages) <= 0 {
						if msg_rq_cnt > 3 {
							if job, exists := historyJobs.Get(subjectChannelID); exists {
								job.Status = historyStatusCompletedNoMoreMessages
								job.Updated = time.Now()
								historyJobs.Set(subjectChannelID, job)
							}
							writeHistoryCache(channel.ID, historyCache{
								Updated:        time.Now(),
								Running:        false,
								RunningBefore:  "",
								CompletedSince: lastMessageID,
							})
							break MessageRequestingLoop
						} else { // retry to make sure no more
							time.Sleep(10 * time.Millisecond)
							goto request_messages
						}
					}

					// Set New Range, this shouldn't be changed regardless of before/since filters. The bot will always go latest to oldest.
					beforeID = messages[len(messages)-1].ID
					beforeTime = messages[len(messages)-1].Timestamp
					sinceID = ""

					// Process Messages
					if sourceConfig.HistoryTyping != nil && !autorun {
						if *sourceConfig.HistoryTyping && hasPermsToRespond {
							bot.ChannelTyping(commandingMessage.ChannelID)
						}
					}
					for _, message := range messages {
						// Ordered to Cancel
						if job, exists := historyJobs.Get(subjectChannelID); exists {
							if job.Status == historyStatusAbortRequested {
								job.Status = historyStatusAbortCompleted
								job.Updated = time.Now()
								historyJobs.Set(subjectChannelID, job)
								deleteHistoryCache(channel.ID) //TODO: Replace with different variation of writing cache?
								break MessageRequestingLoop
							}
						}

						lastMessageID = message.ID

						// Check Message Range
						message64, _ := strconv.ParseInt(message.ID, 10, 64)
						if before != "" {
							before64, _ := strconv.ParseInt(before, 10, 64)
							if message64 > before64 { // keep scrolling back in messages
								continue
							}
						}
						if since != "" {
							since64, _ := strconv.ParseInt(since, 10, 64)
							if message64 < since64 { // message too old, kill loop
								if job, exists := historyJobs.Get(subjectChannelID); exists {
									job.Status = historyStatusCompletedToSinceFilter
									job.Updated = time.Now()
									historyJobs.Set(subjectChannelID, job)
								}
								deleteHistoryCache(channel.ID) // unsure of consequences of caching when using filters, so deleting to be safe for now.
								break MessageRequestingLoop
							}
						}

						// Process Message
						timeStartingDownload := time.Now()
						downloadedFiles := handleMessage(message, &channel, false, true)
						if len(downloadedFiles) > 0 {
							totalDownloads += int64(len(downloadedFiles))
							for _, file := range downloadedFiles {
								totalFilesize += file.Filesize
							}
							historyDownloadDuration += time.Since(timeStartingDownload)
						}
						totalMessages++
					}
				}
			}

			// Final log
			log.Println(lg("History", "", color.HiGreenString, logPrefix+"Finished history for \"%s\", %s files, %s total",
				sourceName, formatNumber(totalDownloads), humanize.Bytes(uint64(totalFilesize))))
			// Final status update
			if sendStatus {
				jobStatus := "Unknown"
				if job, exists := historyJobs.Get(subjectChannelID); exists {
					jobStatus = historyStatusLabel(job.Status)
				}
				var status string
				if totalDownloads == 0 {
					status = fmt.Sprintf(
						"``%s:`` **No files found...**\n"+
							"_%s total messages processed, avg %d msg/s_\n\n"+
							"%s\n\n"+ // msgSourceDisplay^
							"**DONE!** - %s\n"+
							"Ran ``%d`` message history requests\n\n"+
							"%s_Duration was %s_",
						timeSinceShort(historyStartTime),
						formatNumber(int64(totalMessages)), int(float64(totalMessages)/time.Since(historyStartTime).Seconds()),
						msgSourceDisplay,
						jobStatus,
						messageRequestCount,
						rangeContent, timeSince(historyStartTime),
					)
				} else {
					status = fmt.Sprintf(
						"``%s:`` **%s total files downloaded!**\n`%s total, avg %1.1f MB/s`\n"+
							"_%s total messages processed, avg %d msg/s_\n\n"+
							"%s\n\n"+ // msgSourceDisplay^
							"**DONE!** - %s\n"+
							"Ran ``%d`` message history requests\n\n"+
							"%s_Duration was %s_",
						timeSinceShort(historyStartTime), formatNumber(int64(totalDownloads)),
						humanize.Bytes(uint64(totalFilesize)), float64(totalFilesize/humanize.MByte)/historyDownloadDuration.Seconds(),
						formatNumber(int64(totalMessages)), int(float64(totalMessages)/time.Since(historyStartTime).Seconds()),
						msgSourceDisplay,
						jobStatus,
						messageRequestCount,
						rangeContent, timeSince(historyStartTime),
					)
				}
				if !hasPermsToRespond {
					log.Println(lg("History", "", color.HiRedString, logPrefix+fmtBotSendPerm, responseMsg.ChannelID))
				} else {
					if responseMsg == nil {
						log.Println(lg("History", "", color.RedString,
							logPrefix+"Tried to edit status message but it doesn't exist, sending new one."))
						if _, err = replyEmbed(responseMsg, "Command — History", status); err != nil { // Failed to Edit Status, Send New Message
							log.Println(lg("History", "", color.HiRedString,
								logPrefix+"Failed to send replacement status message:\t%s", err))
						}
					} else {
						if selfbot {
							responseMsg, err = bot.ChannelMessageEdit(responseMsg.ChannelID, responseMsg.ID,
								fmt.Sprintf("**Command — History**\n\n%s", status))
						} else {
							responseMsg, err = bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
								ID:      responseMsg.ID,
								Channel: responseMsg.ChannelID,
								Embed:   buildEmbed(responseMsg.ChannelID, "Command — History", status),
							})
						}
						// Edit failure
						if err != nil {
							log.Println(lg("History", "", color.RedString,
								logPrefix+"Failed to edit status message, sending new one:\t%s", err))
							if _, err = replyEmbed(responseMsg, "Command — History", status); err != nil {
								log.Println(lg("History", "", color.HiRedString,
									logPrefix+"Failed to send replacement status message:\t%s", err))
							}
						}
					}
				}
			}
		}
	}

	return int(totalDownloads)
}
