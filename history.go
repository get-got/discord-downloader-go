package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"mvdan.cc/xurls/v2"
)

var (
	historyStatus map[string]string
)

func handleHistory(commandingMessage *discordgo.Message, subjectChannelID string) int {
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
	var beforeTime time.Time

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
		log.Println(logPrefixHistory, color.CyanString(logPrefix+"Began checking history...",
			subjectChannelID, commander))

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
						log.Println(logPrefixDebug, color.YellowString(logPrefix+"Wrote to cache file."))
					}
					f.Close()
				}

				// Status Update
				if commandingMessage != nil {
					log.Println(logPrefixHistory, color.CyanString(logPrefix+"Requesting 100 more, %d downloaded, %d processed — Before %s",
						subjectChannelID, commander, d, i, beforeTime))
					if message != nil {
						if hasPerms(message.ChannelID, discordgo.PermissionSendMessages) {
							content := fmt.Sprintf("``%s:`` **%s files downloaded**\n``%s messages processed``\n\n`(%d)` _Processing more messages, please wait..._",
								durafmt.ParseShort(time.Since(historyStartTime)).String(),
								formatNumber(d), formatNumber(i), batch)
							message, err = bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
								ID:      message.ID,
								Channel: message.ChannelID,
								Embed:   buildEmbed(message.ChannelID, "Command — History", content),
							})
							// Edit failure, so send replacement status
							if err != nil {
								log.Println(logPrefixHistory, color.RedString(logPrefix+"Failed to edit status message, sending new one:\t%s", subjectChannelID, commander, err))
								message, err = replyEmbed(message, "Command — History", content)
								if err != nil {
									log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send replacement status message:\t%s", subjectChannelID, commander, err))
								}
							}
						} else {
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm,
								subjectChannelID, commander, message.ChannelID))
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
			messages, err := bot.ChannelMessages(subjectChannelID, 100, beforeID, "", "")
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
					log.Println(logPrefixHistory, color.RedString(logPrefix+"Failed to fetch message timestamp:\t%s", subjectChannelID, commander, err))
				}
				// Process Messages
				bot.ChannelTyping(commandingMessage.ChannelID)
				for _, message := range messages {
					fileTime := time.Now()
					if message.Timestamp != "" {
						fileTime, err = message.Timestamp.Parse()
						if err != nil {
							log.Println(logPrefixHistory, color.RedString(logPrefix+"Failed to parse message timestamp:\t%s", subjectChannelID, commander, err))
						}
					}
					// Ordered to Cancel
					if historyStatus[message.ChannelID] == "cancel" {
						delete(historyStatus, message.ChannelID)
						break MessageRequestingLoop
					}

					// Process
					if message.Author.ID != user.ID || config.ScanOwnMessages {
						for _, iAttachment := range message.Attachments {
							if len(dbFindDownloadByURL(iAttachment.URL)) == 0 {
								download := startDownload(
									iAttachment.URL,
									iAttachment.Filename,
									channelConfig.Destination,
									message,
									fileTime,
									true,
								)
								if download.Status == downloadSuccess {
									d++
								}
							}
						}
						foundUrls := xurls.Strict().FindAllString(message.Content, -1)
						for _, iFoundUrl := range foundUrls {
							links := getDownloadLinks(iFoundUrl, subjectChannelID)
							for link, filename := range links {
								if len(dbFindDownloadByURL(link)) == 0 {
									download := startDownload(
										link,
										filename,
										channelConfig.Destination,
										message,
										fileTime,
										true,
									)
									if download.Status == downloadSuccess {
										d++
									}
								}
							}
						}
						i++
					}
				}
			} else {
				// Error requesting messages
				if message != nil {
					if hasPerms(message.ChannelID, discordgo.PermissionSendMessages) {
						_, err = replyEmbed(message, "Command — History", fmt.Sprintf("Encountered an error requesting messages: %s", err.Error()))
						if err != nil {
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send error message:\t%s", subjectChannelID, commander, err))
						}
					} else {
						log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm,
							subjectChannelID, commander, message.ChannelID))
					}
				}
				log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Error requesting messages:\t%s", subjectChannelID, commander, err))
				delete(historyStatus, subjectChannelID)
				break MessageRequestingLoop
			}
		}

		// Final status update
		if commandingMessage != nil {
			if message != nil {
				if hasPerms(message.ChannelID, discordgo.PermissionSendMessages) {
					contentFinal := fmt.Sprintf("``%s:`` **%s total files downloaded!**\n``%s total messages processed``\n\nFinished cataloging history for ``%s``\n``%d`` message history requests\n\n_Duration was %s_",
						durafmt.ParseShort(time.Since(historyStartTime)).String(),
						formatNumber(int64(d)), formatNumber(int64(i)),
						subjectChannelID, batch,
						durafmt.Parse(time.Since(historyStartTime)).String(),
					)
					message, err = bot.ChannelMessageEditComplex(&discordgo.MessageEdit{
						ID:      message.ID,
						Channel: message.ChannelID,
						Embed:   buildEmbed(message.ChannelID, "Command — History", contentFinal),
					})
					// Edit failure
					if err != nil {
						log.Println(logPrefixHistory, color.RedString(logPrefix+"Failed to edit status message, sending new one:\t%s",
							subjectChannelID, commander, err))
						message, err = replyEmbed(message, "Command — History", contentFinal)
						if err != nil {
							log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Failed to send replacement status message:\t%s",
								subjectChannelID, commander, err))
						}
					}
				} else {
					log.Println(logPrefixHistory, color.HiRedString(logPrefix+fmtBotSendPerm,
						subjectChannelID, commander, message.ChannelID))
				}
			} else {
				log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Tried to edit status message but it doesn't exist.", subjectChannelID, commander))
			}
		}

		// Final log
		log.Println(logPrefixHistory, color.HiCyanString(logPrefix+"Finished history, %s files",
			subjectChannelID, commander, formatNumber(d)),
		)

		// Delete Cache File
		if historyCachePath != "" {
			filepath := historyCachePath + string(os.PathSeparator) + subjectChannelID
			if _, err := os.Stat(filepath); err == nil {
				err = os.Remove(filepath)
				if err != nil {
					log.Println(logPrefixHistory, color.HiRedString(logPrefix+"Encountered error deleting cache file:\t%s",
						subjectChannelID, commander, err))
				} else if commandingMessage != nil && config.DebugOutput {
					log.Println(logPrefixDebug, color.YellowString(logPrefix+"Deleted cache file.", subjectChannelID, commander))
				}
			}
		}

	}

	return int(d)
}
