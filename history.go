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

var (
	historyStatus map[string]string
)

func handleHistory(commandingMessage *discordgo.Message, subjectChannelID string, before string, since string) int {
	// Identifier
	var commander string = "AUTORUN"
	if commandingMessage != nil {
		commander = getUserIdentifier(*commandingMessage.Author)
	}

	logPrefix := fmt.Sprintf("%s/%s: ", subjectChannelID, commander)

	// Check Read History perms
	if !hasPerms(subjectChannelID, discordgo.PermissionReadMessageHistory) {
		log.Println(logPrefixHistory, color.HiRedString(logPrefix+"BOT DOES NOT HAVE PERMISSION TO READ MESSAGE HISTORY!!!"))
		return 0
	}

	// Mark active
	historyStatus[subjectChannelID] = "downloading"

	var i int64 = 0
	var d int64 = 0
	var batch int = 0

	var beforeID string
	if before != "" {
		beforeID = before
	}
	var beforeTime time.Time

	var sinceID string
	if since != "" {
		sinceID = since
	}

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

	var err error
	var message *discordgo.Message = nil

	if isChannelRegistered(subjectChannelID) {
		channelConfig := getChannelConfig(subjectChannelID)

		// Open Cache File?
		if historyCachePath != "" {
			filepath := historyCachePath + string(os.PathSeparator) + subjectChannelID
			if f, err := ioutil.ReadFile(filepath); err == nil {
				beforeID = string(f)
				if commandingMessage != nil && config.DebugOutput {
					log.Println(logPrefixDebug, color.YellowString(logPrefix+"Found a cache file, picking up where we left off...", subjectChannelID, commander))
				}
			}
		}

		historyStartTime := time.Now()

		// Initial Status Message
		if commandingMessage != nil {
			if hasPerms(commandingMessage.ChannelID, discordgo.PermissionSendMessages) {
				message, err = replyEmbed(commandingMessage, "Command — History", "Starting to save channel history, please wait...")
				if err != nil {
					log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send command embed message:\t%s", err))
				}
			} else {
				log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm, commandingMessage.ChannelID))
			}
		}
		log.Println(logPrefixHistory, color.CyanString(logPrefix+"Began checking history for %s...", subjectChannelID))

	MessageRequestingLoop:
		for true {
			// Next 100
			if beforeTime != (time.Time{}) {
				batch++

				// Write to cache file
				if historyCachePath != "" {
					err := os.MkdirAll(historyCachePath, 0755)
					if err != nil {
						log.Println(logPrefixHistory, color.HiRedString("Error while creating history cache folder \"%s\": %s", historyCachePath, err))
					}

					filepath := historyCachePath + string(os.PathSeparator) + subjectChannelID
					f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
					if err != nil {
						log.Println(logPrefixHistory, color.RedString("Failed to open cache file:\t%s", err))
					}
					if _, err = f.WriteString(beforeID); err != nil {
						log.Println(logPrefixHistory, color.RedString("Failed to write cache file:\t%s", err))
					} else if commandingMessage != nil && config.DebugOutput {
						log.Println(logPrefixDebug, logPrefixHistory, color.YellowString(logPrefix+"Wrote to cache file."))
					}
					f.Close()
				}

				// Status Update
				if commandingMessage != nil {
					log.Println(logPrefixHistory, color.CyanString(logPrefix+"Requesting 100 more, %d downloaded, %d processed — Before %s",
						d, i, beforeTime))
					if message != nil {
						if hasPerms(message.ChannelID, discordgo.PermissionSendMessages) {
							content := fmt.Sprintf("``%s:`` **%s files downloaded**\n``%s messages processed``\n\n%s`(%d)` _Processing more messages, please wait..._",
								durafmt.ParseShort(time.Since(historyStartTime)).String(),
								formatNumber(d), formatNumber(i), rangeContent, batch)
							message, err = bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
								ID:      message.ID,
								Channel: message.ChannelID,
								Embed:   buildEmbed(message.ChannelID, "Command — History", content),
							})
							// Edit failure, so send replacement status
							if err != nil {
								log.Println(logPrefixHistory, color.RedString(logPrefix+"Failed to edit status message, sending new one:\t%s", err))
								message, err = replyEmbed(message, "Command — History", content)
								if err != nil {
									log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send replacement status message:\t%s", err))
								}
							}
						} else {
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm, message.ChannelID))
						}
					} else {
						log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Tried to edit status message but it doesn't exist.", subjectChannelID, commander))
					}
				}
				// Update presence
				timeLastUpdated = time.Now()
				if *channelConfig.UpdatePresence {
					updateDiscordPresence()
				}
			}

			// Request More
			messages, err := bot.ChannelMessages(subjectChannelID, 100, beforeID, sinceID, "")
			if err == nil {
				// No More Messages
				if len(messages) <= 0 {
					delete(historyStatus, subjectChannelID)
					break MessageRequestingLoop
				}
				// Go Back
				beforeID = messages[len(messages)-1].ID
				beforeTime, err = messages[len(messages)-1].Timestamp.Parse()
				if err != nil {
					log.Println(logPrefixHistory, color.RedString(logPrefix+"Failed to fetch message timestamp:\t%s", err))
				}
				sinceID = ""
				// Process Messages
				if *channelConfig.TypeWhileProcessing && hasPerms(commandingMessage.ChannelID, discordgo.PermissionSendMessages) {
					bot.ChannelTyping(commandingMessage.ChannelID)
				}
				for _, message := range messages {
					// Ordered to Cancel
					if historyStatus[message.ChannelID] == "cancel" {
						delete(historyStatus, message.ChannelID)
						break MessageRequestingLoop
					}

					// Check Before/Since
					message64, _ := strconv.ParseInt(message.ID, 10, 64)
					if before != "" {
						before64, _ := strconv.ParseInt(before, 10, 64)
						if message64 > before64 {
							continue
						}
					}
					if since != "" {
						since64, _ := strconv.ParseInt(since, 10, 64)
						if message64 < since64 {
							continue
						}
					}

					// Process
					downloadCount := handleMessage(message, false, true)
					if downloadCount > 0 {
						d += downloadCount
					}
					i++
				}
			} else {
				// Error requesting messages
				if message != nil {
					if hasPerms(message.ChannelID, discordgo.PermissionSendMessages) {
						_, err = replyEmbed(message, "Command — History", fmt.Sprintf("Encountered an error requesting messages: %s", err.Error()))
						if err != nil {
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send error message:\t%s", err))
						}
					} else {
						log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm, message.ChannelID))
					}
				}
				log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Error requesting messages:\t%s", err))
				delete(historyStatus, subjectChannelID)
				break MessageRequestingLoop
			}
		}

		// Final status update
		if commandingMessage != nil {
			if message != nil {
				if hasPerms(message.ChannelID, discordgo.PermissionSendMessages) {
					contentFinal := fmt.Sprintf("``%s:`` **%s total files downloaded!**\n``%s total messages processed``\n\nFinished cataloging history for ``%s``\n``%d`` message history requests\n\n%s_Duration was %s_",
						durafmt.ParseShort(time.Since(historyStartTime)).String(),
						formatNumber(int64(d)), formatNumber(int64(i)),
						subjectChannelID, batch,
						rangeContent,
						durafmt.Parse(time.Since(historyStartTime)).String(),
					)
					message, err = bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
						ID:      message.ID,
						Channel: message.ChannelID,
						Embed:   buildEmbed(message.ChannelID, "Command — History", contentFinal),
					})
					// Edit failure
					if err != nil {
						log.Println(logPrefixHistory, color.RedString(logPrefix+"Failed to edit status message, sending new one:\t%s", err))
						message, err = replyEmbed(message, "Command — History", contentFinal)
						if err != nil {
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send replacement status message:\t%s", err))
						}
					}
				} else {
					log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm, message.ChannelID))
				}
			} else {
				log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Tried to edit status message but it doesn't exist.", subjectChannelID, commander))
			}
		}

		// Final log
		log.Println(logPrefixHistory, color.HiCyanString(logPrefix+"Finished history, %s files", formatNumber(d)))

		// Delete Cache File
		if historyCachePath != "" {
			filepath := historyCachePath + string(os.PathSeparator) + subjectChannelID
			if _, err := os.Stat(filepath); err == nil {
				err = os.Remove(filepath)
				if err != nil {
					log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Encountered error deleting cache file:\t%s", err))
				} else if commandingMessage != nil && config.DebugOutput {
					log.Println(logPrefixDebug, logPrefixHistory, color.YellowString(logPrefix+"Deleted cache file."))
				}
			}
		}

	}

	return int(d)
}
