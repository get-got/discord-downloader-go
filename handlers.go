package main

import (
	"fmt"
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

//#region Events

var lastMessageID string

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if lastMessageID != m.ID {
		handleMessage(m.Message, nil, false, false)
	}
	lastMessageID = m.ID
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if lastMessageID != m.ID {
		if m.EditedTimestamp != nil {
			handleMessage(m.Message, nil, true, false)
		}
	}
	lastMessageID = m.ID
}

func handleMessage(m *discordgo.Message, c *discordgo.Channel, edited bool, history bool) (int64, int64) {
	// Ignore own messages unless told not to
	if m.Author.ID == botUser.ID && !config.ScanOwnMessages {
		return -1, 0
	}

	if !history && !edited {
		timeLastMessage = time.Now()
	}

	// Admin Channel
	if isAdminChannelRegistered(m.ChannelID) {
		m = fixMessage(m)

		// Log
		sendLabel := fmt.Sprintf("%s in \"%s\"#%s",
			getUserIdentifier(*m.Author),
			getGuildName(m.GuildID), getChannelName(m.ChannelID, nil),
		)
		content := m.Content
		if len(m.Attachments) > 0 {
			content = content + fmt.Sprintf(" (%d attachments)", len(m.Attachments))
		}
		if edited {
			log.Println(lg("Message", "ADMIN CHANNEL", color.CyanString, "Edited [%s]: %s", sendLabel, content))
		} else {
			log.Println(lg("Message", "ADMIN CHANNEL", color.CyanString, "[%s]: %s", sendLabel, content))
		}
	}

	// Registered Channel
	if channelConfig := getSource(m, c); channelConfig != emptyConfig {
		// Ignore bots if told to do so
		if m.Author.Bot && *channelConfig.IgnoreBots {
			return -1, 0
		}
		// Ignore if told so by config
		if (!history && !*channelConfig.Enabled) || (edited && !*channelConfig.ScanEdits) {
			return -1, 0
		}

		m = fixMessage(m)

		// Log
		if config.MessageOutput {
			sendLabel := fmt.Sprintf("%s in \"%s\"#%s",
				getUserIdentifier(*m.Author),
				getGuildName(m.GuildID), getChannelName(m.ChannelID, nil),
			)
			content := m.Content
			if len(m.Attachments) > 0 {
				content += fmt.Sprintf(" \t[%d attachments]", len(m.Attachments))
			}
			content += fmt.Sprintf(" \t<%s>", m.ID)

			if !history || config.MessageOutputHistory {
				addOut := ""
				if history && config.MessageOutputHistory && !m.Timestamp.IsZero() {
					addOut = fmt.Sprintf(" @ %s", m.Timestamp.String()[:19])
				}
				if edited {
					log.Println(lg("Message", "", color.CyanString, "Edited [%s%s]: %s", sendLabel, addOut, content))
				} else {
					log.Println(lg("Message", "", color.CyanString, "[%s%s]: %s", sendLabel, addOut, content))
				}
			}
		}

		// Log Messages to File
		if channelConfig.LogMessages != nil {
			if channelConfig.LogMessages.Destination != "" {
				logPath := channelConfig.LogMessages.Destination
				if *channelConfig.LogMessages.DestinationIsFolder {
					if !strings.HasSuffix(logPath, string(os.PathSeparator)) {
						logPath += string(os.PathSeparator)
					}
					err := os.MkdirAll(logPath, 0755)
					if err == nil {
						logPath += "Log_Messages"
						if *channelConfig.LogMessages.DivideLogsByServer {
							if m.GuildID == "" {
								ch, err := bot.State.Channel(m.ChannelID)
								if err != nil && c != nil {
									ch = c
								}
								if ch != nil {
									if ch.Type == discordgo.ChannelTypeDM {
										logPath += " DM"
									} else if ch.Type == discordgo.ChannelTypeGroupDM {
										logPath += " GroupDM"
									} else if ch.Type == discordgo.ChannelTypeGuildText {
										logPath += " Text"
									} else if ch.Type == discordgo.ChannelTypeGuildCategory {
										logPath += " Category"
									} else if ch.Type == discordgo.ChannelTypeGuildForum {
										logPath += " Forum"
									} else if ch.Type == discordgo.ChannelTypeGuildPrivateThread || ch.Type == discordgo.ChannelTypeGuildPublicThread {
										logPath += " Thread"
									} else {
										logPath += " Unknown"
									}

									if ch.Name != "" {
										logPath += " - " + clearPath(ch.Name) + " -"
									} else if ch.Topic != "" {
										logPath += " - " + clearPath(ch.Topic) + " -"
									}
								}
							} else {
								logPath += " SID_" + m.GuildID
							}
						}
						if *channelConfig.LogMessages.DivideLogsByChannel {
							logPath += " CID_" + m.ChannelID
						}
						if *channelConfig.LogMessages.DivideLogsByUser {
							logPath += " UID_" + m.Author.ID
						}
					}
					logPath += ".txt"
				}
				// Read
				currentLog, err := os.ReadFile(logPath)
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
						log.Println(lg("Message", "", color.RedString, "[channelConfig.LogMessages] Failed to open log file:\t%s", err))
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
						if *channelConfig.LogMessages.UserData {
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
						log.Println(lg("Message", "", color.RedString, "[channelConfig.LogMessages] Failed to append file:\t%s", err))
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
				if config.Debug {
					log.Println(lg("Debug", "Message", color.YellowString,
						"%s Filter will be ignoring by default...",
						color.HiMagentaString("(FILTER)")))
				}
			}

			if channelConfig.Filters.BlockedPhrases != nil {
				for _, phrase := range *channelConfig.Filters.BlockedPhrases {
					if strings.Contains(m.Content, phrase) && phrase != "" {
						shouldAbort = true
						if config.Debug {
							log.Println(lg("Debug", "Message", color.YellowString,
								"%s blockedPhrases found \"%s\" in message, planning to abort...",
								color.HiMagentaString("(FILTER)"), phrase))
						}
						break
					}
				}
			}
			if channelConfig.Filters.AllowedPhrases != nil {
				for _, phrase := range *channelConfig.Filters.AllowedPhrases {
					if strings.Contains(m.Content, phrase) && phrase != "" {
						shouldAbort = false
						if config.Debug {
							log.Println(lg("Debug", "Message", color.YellowString,
								"%s allowedPhrases found \"%s\" in message, planning to process...",
								color.HiMagentaString("(FILTER)"), phrase))
						}
						break
					}
				}
			}

			if channelConfig.Filters.BlockedUsers != nil {
				if stringInSlice(m.Author.ID, *channelConfig.Filters.BlockedUsers) {
					shouldAbort = true
					if config.Debug {
						log.Println(lg("Debug", "Message", color.YellowString,
							"%s blockedUsers caught %s, planning to abort...",
							color.HiMagentaString("(FILTER)"), m.Author.ID))
					}
				}
			}
			if channelConfig.Filters.AllowedUsers != nil {
				if stringInSlice(m.Author.ID, *channelConfig.Filters.AllowedUsers) {
					shouldAbort = false
					if config.Debug {
						log.Println(lg("Debug", "Message", color.YellowString,
							"%s allowedUsers caught %s, planning to process...",
							color.HiMagentaString("(FILTER)"), m.Author.ID))
					}
				}
			}

			if channelConfig.Filters.BlockedRoles != nil {
				member := m.Member
				if member == nil {
					member, _ = bot.GuildMember(m.GuildID, m.Author.ID)
				}
				if member != nil {
					for _, role := range member.Roles {
						if stringInSlice(role, *channelConfig.Filters.BlockedRoles) {
							shouldAbort = true
							if config.Debug {
								log.Println(lg("Debug", "Message", color.YellowString,
									"%s blockedRoles caught %s, planning to abort...",
									color.HiMagentaString("(FILTER)"), role))
							}
							break
						}
					}
				}
			}
			if channelConfig.Filters.AllowedRoles != nil {
				member := m.Member
				if member == nil {
					member, _ = bot.GuildMember(m.GuildID, m.Author.ID)
				}
				if member != nil {
					for _, role := range member.Roles {
						if stringInSlice(role, *channelConfig.Filters.AllowedRoles) {
							shouldAbort = false
							if config.Debug {
								log.Println(lg("Debug", "Message", color.YellowString,
									"%s allowedRoles caught %s, planning to allow...",
									color.HiMagentaString("(FILTER)"), role))
							}
							break
						}
					}
				}
			}

			// Abort
			if shouldAbort {
				if config.Debug {
					log.Println(lg("Debug", "Message", color.YellowString,
						"%s Filter decided to ignore message...",
						color.HiMagentaString("(FILTER)")))
				}
				return -1, 0
			}
		}

		// Delays
		delay := 0
		if history {
			if channelConfig.DelayHandlingHistory != nil {
				delay = *channelConfig.DelayHandlingHistory
			}
		} else {
			if channelConfig.DelayHandling != nil {
				delay = *channelConfig.DelayHandling
			}
		}
		if delay > 0 {
			if config.Debug {
				log.Println(lg("Debug", "Message", color.YellowString, "Delaying for %d milliseconds...", delay))
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		// Process Files
		var downloadCount int64 = 0
		var totalfilesize int64 = 0
		files := getFileLinks(m)
		for _, file := range files {
			if file.Link == "" {
				continue
			}
			if config.Debug && (!history || config.MessageOutputHistory) {
				log.Println(lg("Debug", "Message", color.HiCyanString, "FOUND FILE: "+file.Link+fmt.Sprintf(" \t<%s>", m.ID)))
			}
			status, filesize := downloadRequestStruct{
				InputURL:   file.Link,
				Filename:   file.Filename,
				Path:       channelConfig.Destination,
				Message:    m,
				FileTime:   file.Time,
				HistoryCmd: history,
				EmojiCmd:   false,
				StartTime:  time.Now(),
			}.handleDownload()
			if status.Status == downloadSuccess {
				downloadCount++
				totalfilesize += filesize
			}
		}
		return downloadCount, totalfilesize
	}

	return -1, 0
}

//#endregion
