package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

type historyStatus int

const (
	historyStatusWaiting historyStatus = iota
	historyStatusDownloading
	historyStatusAbortRequested
	historyStatusAbortCompleted
	historyStatusErrorRequesting
	historyStatusCompletedNoMoreMessages
	historyStatusCompletedToBeforeFilter
	historyStatusCompletedToSinceFilter
)

func historyStatusLabel(status historyStatus) string {
	switch status {
	case historyStatusWaiting:
		return "Waiting..."
	case historyStatusDownloading:
		return "Downloading..."
	case historyStatusAbortRequested:
		return "Abort Requested..."
	case historyStatusAbortCompleted:
		return "Aborted..."
	case historyStatusErrorRequesting:
		return "ERROR REQUESTING MESSAGES"
	case historyStatusCompletedNoMoreMessages:
		return "Completed! No More Messages"
	case historyStatusCompletedToBeforeFilter:
		return "Completed! Exceeded Before Date Filter"
	case historyStatusCompletedToSinceFilter:
		return "Completed! Exceeded Since Date Filter"
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
	Updated                 time.Time
}

var (
	historyJobs map[string]historyJob
)

func handleHistory(commandingMessage *discordgo.Message, subjectChannelID string, before string, since string) int {
	if job, exists := historyJobs[subjectChannelID]; exists && job.Status != historyStatusWaiting {
		log.Println(logPrefixHistory, color.RedString("History job skipped, Status: %s", historyStatusLabel(job.Status)))
		return -1
	}

	var err error

	var totalMessages int64 = 0
	var totalDownloads int64 = 0
	var messageRequestCount int = 0

	var responseMsg *discordgo.Message = &discordgo.Message{}
	responseMsg.ID = ""
	responseMsg.ChannelID = subjectChannelID

	var commander string = "AUTORUN"
	var autorun bool = true
	if commandingMessage != nil { // Only time commandingMessage is nil is Autorun
		commander = getUserIdentifier(*commandingMessage.Author)
		autorun = false
	}
	logPrefix := fmt.Sprintf("%s/%s: ", subjectChannelID, commander)

	// Send Status?
	var sendStatus bool = true
	if (autorun && !config.SendAutorunHistoryStatus) || (!autorun && !config.SendHistoryStatus) {
		sendStatus = false
	}

	// Check Read History perms
	if !hasPerms(subjectChannelID, discordgo.PermissionReadMessageHistory) {
		log.Println(logPrefixHistory, color.HiRedString(logPrefix+"BOT DOES NOT HAVE PERMISSION TO READ MESSAGE HISTORY!!!"))
		return -1
	}
	hasPermsToRespond := hasPerms(subjectChannelID, discordgo.PermissionSendMessages)
	if !autorun {
		hasPermsToRespond = hasPerms(commandingMessage.ChannelID, discordgo.PermissionSendMessages)
	}

	// Update Job Status to Downloading
	if job, exists := historyJobs[subjectChannelID]; exists {
		job.Status = historyStatusDownloading
		job.Updated = time.Now()
		historyJobs[subjectChannelID] = job
	}

	// Date Range Vars
	var beforeID = before
	var sinceID = since
	var beforeTime time.Time

	//#region Date Range Output
	rangeContent := ""
	if since != "" {
		if isDate(since) {
			rangeContent += fmt.Sprintf("**Since:** `%s`\n", discordSnowflakeToTimestamp(since, "2006-01-02"))
		} else if isNumeric(since) {
			rangeContent += fmt.Sprintf("**Since:** `%s`\n", since)
		}
	}
	if before != "" {
		if isDate(before) {
			rangeContent += fmt.Sprintf("**Before:** `%s`", discordSnowflakeToTimestamp(before, "2006-01-02"))
		} else if isNumeric(before) {
			rangeContent += fmt.Sprintf("**Before:** `%s`\n", before)
		}
	}
	if rangeContent != "" {
		rangeContent += "\n\n"
	}
	//#endregion

	ch := channelRegistered(responseMsg)
	if ch != "" {
		channelConfig := getChannelConfig(ch)

		// Overwrite Send Status
		if channelConfig.OverwriteSendAutorunHistoryStatus != nil {
			if autorun && !*channelConfig.OverwriteSendAutorunHistoryStatus {
				sendStatus = false
			}
		}
		if channelConfig.OverwriteSendHistoryStatus != nil {
			if !autorun && !*channelConfig.OverwriteSendHistoryStatus {
				sendStatus = false
			}
		}

		// Open Cache File?
		if historyCachePath != "" {
			if f, err := ioutil.ReadFile(historyCachePath + string(os.PathSeparator) + subjectChannelID); err == nil {
				beforeID = string(f)
				if !autorun && config.DebugOutput {
					log.Println(logPrefixDebug, color.YellowString(logPrefix+"Found a cache file, picking up where we left off...", subjectChannelID, commander))
				}
			}
		}

		historyStartTime := time.Now()

		// Initial Status Message
		if sendStatus {
			if hasPermsToRespond {
				responseMsg, err = replyEmbed(commandingMessage, "Command — History", fmt.Sprintf("Starting to save history, please wait...\n\n`Server:` **%s**\n`Channel:` _#%s_\n\n",
					getGuildName(getChannelGuildID(subjectChannelID)),
					getChannelName(subjectChannelID),
				))
				if err != nil {
					log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send command embed message:\t%s", err))
				}
			} else {
				log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm, commandingMessage.ChannelID))
			}
		}
		log.Println(logPrefixHistory, color.CyanString(logPrefix+"Began checking history for %s...", subjectChannelID))

	MessageRequestingLoop:
		for {
			// Next 100
			if beforeTime != (time.Time{}) {
				messageRequestCount++

				// Write to cache file
				if historyCachePath != "" {
					if err := os.MkdirAll(historyCachePath, 0755); err != nil {
						log.Println(logPrefixHistory, color.HiRedString("Error while creating history cache folder \"%s\": %s", historyCachePath, err))
					}
					filepath := historyCachePath + string(os.PathSeparator) + subjectChannelID
					f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
					if err != nil {
						log.Println(logPrefixHistory, color.RedString("Failed to open cache file:\t%s", err))
					}
					if _, err = f.WriteString(beforeID); err != nil {
						log.Println(logPrefixHistory, color.RedString("Failed to write cache file:\t%s", err))
					} else if !autorun && config.DebugOutput {
						log.Println(logPrefixDebug, logPrefixHistory, color.YellowString(logPrefix+"Wrote to cache file."))
					}
					f.Close()
				}

				// Update Status
				log.Println(logPrefixHistory,
					color.CyanString(logPrefix+"Requesting 100 more, %d downloaded, %d processed — Before %s",
						totalDownloads, totalMessages, beforeTime))
				if sendStatus {
					status := fmt.Sprintf(
						"``%s:`` **%s files downloaded**\n``"+
							"%s messages processed``\n\n"+
							"`Server:` **%s**\n"+
							"`Channel:` _#%s_\n\n"+
							"%s`(%d)` _Processing more messages, please wait..._",
						durafmt.ParseShort(time.Since(historyStartTime)).String(), formatNumber(totalDownloads),
						formatNumber(totalMessages),
						getGuildName(getChannelGuildID(subjectChannelID)),
						getChannelName(subjectChannelID),
						rangeContent, messageRequestCount)
					if responseMsg == nil {
						log.Println(logPrefixHistory, color.RedString(logPrefix+"Tried to edit status message but it doesn't exist, sending new one."))
						if responseMsg, err = replyEmbed(responseMsg, "Command — History", status); err != nil { // Failed to Edit Status, Send New Message
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send replacement status message:\t%s", err))
						}
					} else {
						if !hasPermsToRespond {
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm+" - %s",
								responseMsg.ChannelID, status))
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
								log.Println(logPrefixHistory, color.RedString(logPrefix+"Failed to edit status message, sending new one:\t%s", err))
								if responseMsg, err = replyEmbed(responseMsg, "Command — History", status); err != nil { // Failed to Edit Status, Send New Message
									log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send replacement status message:\t%s", err))
								}
							}
						}
					}
				}

				// Update presence
				timeLastUpdated = time.Now()
				if *channelConfig.UpdatePresence {
					updateDiscordPresence()
				}
			}

			// Request More Messages
			if messages, err := bot.ChannelMessages(subjectChannelID, 100, beforeID, sinceID, ""); err != nil {
				// Error requesting messages
				if sendStatus {
					if !hasPermsToRespond {
						log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm, responseMsg.ChannelID))
					} else {
						_, err = replyEmbed(responseMsg, "Command — History", fmt.Sprintf("Encountered an error requesting messages for %s: %s", subjectChannelID, err.Error()))
						if err != nil {
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send error message:\t%s", err))
						}
					}
				}
				log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Error requesting messages:\t%s", err))
				if job, exists := historyJobs[subjectChannelID]; exists {
					job.Status = historyStatusErrorRequesting
					job.Updated = time.Now()
					historyJobs[subjectChannelID] = job
				}
				break MessageRequestingLoop
			} else {
				// No More Messages
				if len(messages) <= 0 {
					if job, exists := historyJobs[subjectChannelID]; exists {
						job.Status = historyStatusCompletedNoMoreMessages
						job.Updated = time.Now()
						historyJobs[subjectChannelID] = job
					}
					break MessageRequestingLoop
				}

				// Set New Range
				beforeID = messages[len(messages)-1].ID
				beforeTime = messages[len(messages)-1].Timestamp
				sinceID = ""

				// Process Messages
				if channelConfig.TypeWhileProcessing != nil && !autorun {
					if *channelConfig.TypeWhileProcessing && hasPermsToRespond {
						bot.ChannelTyping(commandingMessage.ChannelID)
					}
				}
				for _, message := range messages {
					// Ordered to Cancel
					if historyJobs[subjectChannelID].Status == historyStatusAbortRequested {
						if job, exists := historyJobs[subjectChannelID]; exists {
							job.Status = historyStatusAbortCompleted
							job.Updated = time.Now()
							historyJobs[subjectChannelID] = job
						}
						break MessageRequestingLoop
						break
					}

					// Check Message Range
					message64, _ := strconv.ParseInt(message.ID, 10, 64)
					if before != "" && since != "" {
						//
					} else if before != "" {
						before64, _ := strconv.ParseInt(before, 10, 64)
						if message64 > before64 {
							if job, exists := historyJobs[subjectChannelID]; exists {
								job.Status = historyStatusCompletedToBeforeFilter
								job.Updated = time.Now()
								historyJobs[subjectChannelID] = job
							}
							break MessageRequestingLoop
						}
					} else if since != "" {
						since64, _ := strconv.ParseInt(since, 10, 64)
						if message64 < since64 {
							if job, exists := historyJobs[subjectChannelID]; exists {
								job.Status = historyStatusCompletedToSinceFilter
								job.Updated = time.Now()
								historyJobs[subjectChannelID] = job
							}
							break MessageRequestingLoop
						}
					}

					// Process Message
					downloadCount := handleMessage(message, false, true)
					if downloadCount > 0 {
						totalDownloads += downloadCount
					}
					totalMessages++
				}
			}
		}

		// Final log
		log.Println(logPrefixHistory, color.HiCyanString(logPrefix+"Finished history, %s files", formatNumber(totalDownloads)))
		// Final status update
		if sendStatus {
			status := fmt.Sprintf(
				"``%s:`` **%s total files downloaded!**\n"+
					"``%s total messages processed``\n\n"+
					"`Server:` **%s**\n"+
					"`Channel:` _#%s_\n\n"+
					"**FINISHED!**\n"+
					"Ran ``%d`` message history requests\n\n"+
					"%s_Duration was %s_",
				durafmt.ParseShort(time.Since(historyStartTime)).String(), formatNumber(int64(totalDownloads)),
				formatNumber(int64(totalMessages)),
				getGuildName(getChannelGuildID(subjectChannelID)),
				getChannelName(subjectChannelID),
				messageRequestCount, rangeContent,
				durafmt.Parse(time.Since(historyStartTime)).String(),
			)
			if !hasPermsToRespond {
				log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm, responseMsg.ChannelID))
			} else {
				if responseMsg == nil {
					log.Println(logPrefixHistory, color.RedString(logPrefix+"Tried to edit status message but it doesn't exist, sending new one."))
					if _, err = replyEmbed(responseMsg, "Command — History", status); err != nil { // Failed to Edit Status, Send New Message
						log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send replacement status message:\t%s", err))
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
						log.Println(logPrefixHistory, color.RedString(logPrefix+"Failed to edit status message, sending new one:\t%s", err))
						if _, err = replyEmbed(responseMsg, "Command — History", status); err != nil {
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send replacement status message:\t%s", err))
						}
					}
				}
			}
		}
	}

	return int(totalDownloads)
}
