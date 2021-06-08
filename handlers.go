package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

type fileItem struct {
	Link     string
	Filename string
	Time     time.Time
}

var (
	skipCommands = []string{
		"skip",
		"ignore",
		"don't save",
		"no save",
	}
)

//#region Events

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	handleMessage(m.Message, false, false)
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.EditedTimestamp != discordgo.Timestamp("") {
		handleMessage(m.Message, true, false)
	}
}

func handleMessage(m *discordgo.Message, edited bool, history bool) int64 {
	// Ignore own messages unless told not to
	if m.Author.ID == user.ID && !config.ScanOwnMessages {
		return -1
	}

	// Admin Channel
	if isAdminChannelRegistered(m.ChannelID) {
		//TODO: Make this its own function
		// If message content is empty (likely due to userbot/selfbot)
		ubIssue := "Message is corrupted due to endpoint restriction"
		if m.Content == "" && len(m.Attachments) == 0 {
			// Get message history
			mCache, err := bot.ChannelMessages(m.ChannelID, 25, "", "", "")
			if err == nil {
				if len(mCache) > 0 {
					for _, mCached := range mCache {
						if mCached.ID == m.ID {
							// Fix original message having empty Guild ID
							guildID := m.GuildID
							// Replace message
							m = mCached
							// ^^
							if m.GuildID == "" && guildID != "" {
								m.GuildID = guildID
							}
							// Parse commands
							dgr.FindAndExecute(bot, strings.ToLower(config.CommandPrefix), bot.State.User.ID, messageToLower(m))

							break
						}
					}
				} else if config.DebugOutput {
					log.Println(logPrefixDebug, color.RedString("%s, and an attempt to get channel messages found nothing...", ubIssue))
				}
			} else if config.DebugOutput {
				log.Println(logPrefixDebug, color.HiRedString("%s, and an attempt to get channel messages encountered an error:\t%s", ubIssue, err))
			}
		}
		if m.Content == "" && len(m.Attachments) == 0 {
			if config.DebugOutput {
				log.Println(logPrefixDebug, color.YellowString("%s, and attempts to fix seem to have failed...", ubIssue))
			}
		}

		// Log
		if !config.ShowMessages && !config.DebugOutput {
			//nothing
		} else {
			var sendLabel string
			if config.DebugOutput {
				sendLabel = fmt.Sprintf("%s/%s/%s", m.GuildID, m.ChannelID, m.Author.ID)
			} else {
				sendLabel = fmt.Sprintf("%s in %s", getUserIdentifier(*m.Author), getSourceName(m.GuildID, m.ChannelID))
			}
			content := m.Content
			if len(m.Attachments) > 0 {
				content = content + fmt.Sprintf(" (%d attachments)", len(m.Attachments))
			}
			if edited {
				log.Println(color.CyanString("Edited Message [%s]: %s", sendLabel, content))
			} else {
				log.Println(color.CyanString("Message [%s]: %s", sendLabel, content))
			}
		}
	}

	// Registered Channel
	if isChannelRegistered(m.ChannelID) {
		channelConfig := getChannelConfig(m.ChannelID)
		// Ignore bots if told to do so
		if m.Author.Bot && *channelConfig.IgnoreBots {
			return -1
		}
		// Ignore if told so by config
		if !*channelConfig.Enabled || (edited && !*channelConfig.ScanEdits) {
			return -1
		}

		//TODO: Make this its own function
		// If message content is empty (likely due to userbot/selfbot)
		if m.Content == "" && len(m.Attachments) == 0 {
			nms, err := bot.ChannelMessages(m.ChannelID, 10, "", "", "")
			if err == nil {
				if len(nms) > 0 {
					for _, nm := range nms {
						if nm.ID == m.ID {
							id := m.GuildID
							m = nm
							if m.GuildID == "" && id != "" {
								m.GuildID = id
							}
							dgr.FindAndExecute(bot, strings.ToLower(config.CommandPrefix), bot.State.User.ID, messageToLower(m))
						}
					}
				}
			}
		}

		// Log
		if !config.ShowMessages && !config.DebugOutput {
			//nothing
		} else {
			var sendLabel string
			if config.DebugOutput {
				sendLabel = fmt.Sprintf("%s/%s/%s", m.GuildID, m.ChannelID, m.Author.ID)
			} else {
				sendLabel = fmt.Sprintf("%s in %s", getUserIdentifier(*m.Author), getSourceName(m.GuildID, m.ChannelID))
			}
			content := m.Content
			if len(m.Attachments) > 0 {
				content = content + fmt.Sprintf(" (%d attachments)", len(m.Attachments))
			}
			if edited {
				log.Println(color.CyanString("Edited Message [%s]: %s", sendLabel, content))
			} else {
				log.Println(color.CyanString("Message [%s]: %s", sendLabel, content))
			}
		}

		// Log Messages to File
		if channelConfig.LogMessages != nil {
			if channelConfig.LogMessages.Destination != "" {
				logPath := channelConfig.LogMessages.Destination
				if *channelConfig.LogMessages.DestinationIsFolder == true {
					if !strings.HasSuffix(logPath, string(os.PathSeparator)) {
						logPath += string(os.PathSeparator)
					}
					err := os.MkdirAll(logPath, 0755)
					if err == nil {
						logPath += "Log_Messages"
						if *channelConfig.LogMessages.DivideLogsByServer == true {
							if m.GuildID == "" {
								ch, err := bot.State.Channel(m.ChannelID)
								if err == nil {
									if ch.Type == discordgo.ChannelTypeDM {
										logPath += " DM"
									} else if ch.Type == discordgo.ChannelTypeGroupDM {
										logPath += " GroupDM"
									} else {
										logPath += " Unknown"
									}
								} else {
									logPath += " Unknown"
								}
							} else {
								logPath += " SID_" + m.GuildID
							}
						}
						if *channelConfig.LogMessages.DivideLogsByChannel == true {
							logPath += " CID_" + m.ChannelID
						}
						if *channelConfig.LogMessages.DivideLogsByUser == true {
							logPath += " UID_" + m.Author.ID
						}
					}
					logPath += ".txt"
				}
				// Read
				currentLog, err := ioutil.ReadFile(logPath)
				currentLogS := ""
				if err == nil {
					currentLogS = string(currentLog)
				}
				canLog := true
				// Filter Duplicates
				if channelConfig.LogMessages.FilterDuplicates != nil {
					if *channelConfig.LogMessages.FilterDuplicates {
						if strings.Contains(currentLogS, fmt.Sprintf("[%s/%s/%s]", m.GuildID, m.ChannelID, m.ID)) {
							canLog = false
						}
					}
				}

				if canLog {
					// Writer
					f, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
					if err != nil {
						log.Println(color.RedString("[channelConfig.LogMessages] Failed to open log file:\t%s", err))
						f.Close()
					}
					defer f.Close()

					var newLine string
					// Prepend
					prefix := ""
					if channelConfig.LogMessages.Prefix != nil {
						prefix = *channelConfig.LogMessages.Prefix
					}
					// More Data
					additionalInfo := ""
					if channelConfig.LogMessages.UserData != nil {
						if *channelConfig.LogMessages.UserData == true {
							additionalInfo = fmt.Sprintf("[%s/%s/%s] \"%s\"#%s (%s) @ %s: ", m.GuildID, m.ChannelID, m.ID, m.Author.Username, m.Author.Discriminator, m.Author.ID, m.Timestamp)
						}
					}
					if len(m.Attachments) > 0 {
						additionalInfo += fmt.Sprintf("<%d ATTACHMENTS> ", len(m.Attachments))
					}
					// Append
					suffix := ""
					if channelConfig.LogMessages.Suffix != nil {
						suffix = *channelConfig.LogMessages.Suffix
					}
					// New Line
					contentFmt, err := m.ContentWithMoreMentionsReplaced(bot)
					if err == nil {
						newLine += "\n" + prefix + additionalInfo + contentFmt + suffix
					} else {
						newLine += "\n" + prefix + additionalInfo + m.Content + suffix
					}

					if _, err = f.WriteString(newLine); err != nil {
						log.Println(color.RedString("[channelConfig.LogMessages] Failed to append file:\t%s", err))
					}
				}
			}
		}

		// Filters
		if channelConfig.Filters != nil {
			shouldAbort := false

			if channelConfig.Filters.AllowedPhrases != nil ||
				channelConfig.Filters.AllowedUsers != nil ||
				channelConfig.Filters.AllowedRoles != nil {
				shouldAbort = true
				if config.DebugOutput {
					log.Println(logPrefixDebug, color.HiMagentaString("(FILTER)"), color.YellowString("Filter will be ignoring by default..."))
				}
			}

			if channelConfig.Filters.BlockedPhrases != nil {
				for _, phrase := range *channelConfig.Filters.BlockedPhrases {
					if strings.Contains(m.Content, phrase) {
						shouldAbort = true
						if config.DebugOutput {
							log.Println(logPrefixDebug, color.HiMagentaString("(FILTER)"), color.YellowString("blockedPhrases found \"%s\" in message, planning to abort...", phrase))
						}
						break
					}
				}
			}
			if channelConfig.Filters.AllowedPhrases != nil {
				for _, phrase := range *channelConfig.Filters.AllowedPhrases {
					if strings.Contains(m.Content, phrase) {
						shouldAbort = false
						if config.DebugOutput {
							log.Println(logPrefixDebug, color.HiMagentaString("(FILTER)"), color.YellowString("allowedPhrases found \"%s\" in message, planning to process...", phrase))
						}
						break
					}
				}
			}

			if channelConfig.Filters.BlockedUsers != nil {
				if stringInSlice(m.Author.ID, *channelConfig.Filters.BlockedUsers) {
					shouldAbort = true
					if config.DebugOutput {
						log.Println(logPrefixDebug, color.HiMagentaString("(FILTER)"), color.YellowString("blockedUsers caught %s, planning to abort...", m.Author.ID))
					}
				}
			}
			if channelConfig.Filters.AllowedUsers != nil {
				if stringInSlice(m.Author.ID, *channelConfig.Filters.AllowedUsers) {
					shouldAbort = false
					if config.DebugOutput {
						log.Println(logPrefixDebug, color.HiMagentaString("(FILTER)"), color.YellowString("allowedUsers caught %s, planning to process...", m.Author.ID))
					}
				}
			}

			if channelConfig.Filters.BlockedRoles != nil {
				for _, role := range m.Member.Roles {
					if stringInSlice(role, *channelConfig.Filters.BlockedRoles) {
						shouldAbort = true
						if config.DebugOutput {
							log.Println(logPrefixDebug, color.HiMagentaString("(FILTER)"), color.YellowString("blockedRoles caught %s, planning to abort...", role))
						}
						break
					}
				}
			}
			if channelConfig.Filters.AllowedRoles != nil {
				for _, role := range m.Member.Roles {
					if stringInSlice(role, *channelConfig.Filters.AllowedRoles) {
						shouldAbort = false
						if config.DebugOutput {
							log.Println(logPrefixDebug, color.HiMagentaString("(FILTER)"), color.YellowString("allowedRoles caught %s, planning to allow...", role))
						}
						break
					}
				}
			}

			// Abort
			if shouldAbort {
				if config.DebugOutput {
					log.Println(logPrefixDebug, color.HiMagentaString("(FILTER)"), color.HiYellowString("Filter decided to ignore message..."))
				}
				return -1
			} else {
				if config.DebugOutput {
					log.Println(logPrefixDebug, color.HiMagentaString("(FILTER)"), color.HiYellowString("Filter approved message..."))
				}
			}

		}

		// Skipping
		canSkip := config.AllowSkipping
		if channelConfig.OverwriteAllowSkipping != nil {
			canSkip = *channelConfig.OverwriteAllowSkipping
		}
		if canSkip {
			for _, cmd := range skipCommands {
				if m.Content == cmd {
					log.Println(color.HiYellowString("Message handling skipped due to use of skip command."))
					return -1
				}
			}
		}

		// Process Files
		var downloadCount int64
		files := getFileLinks(m)
		for _, file := range files {
			log.Println(color.CyanString("> FILE: " + file.Link))

			status := startDownload(
				file.Link,
				file.Filename,
				channelConfig.Destination,
				m,
				file.Time,
				history,
				false,
			)
			if status.Status == downloadSuccess {
				downloadCount++
			}
		}
		return downloadCount
	}

	return -1
}

//#endregion
