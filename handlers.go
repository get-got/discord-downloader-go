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

func handleMessage(m *discordgo.Message, c *discordgo.Channel, edited bool, history bool) []downloadedItem {
	// Ignore own messages unless told not to
	if m.Author.ID == botUser.ID && !config.ScanOwnMessages {
		return nil
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
			getServerLabel(m.GuildID), getChannelLabel(m.ChannelID, nil),
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
	if sourceConfig := getSource(m, c); sourceConfig != emptyConfig {
		// Ignore bots if told to do so
		if m.Author.Bot && *sourceConfig.IgnoreBots {
			return nil
		}
		// Ignore if told so by config
		if (!history && !*sourceConfig.Enabled) || (edited && !*sourceConfig.ScanEdits) {
			return nil
		}

		m = fixMessage(m)

		// Log
		if config.MessageOutput {
			sendLabel := fmt.Sprintf("%s in \"%s\"#%s",
				getUserIdentifier(*m.Author),
				getServerLabel(m.GuildID), getChannelLabel(m.ChannelID, nil),
			)
			content := m.Content
			if len(m.Attachments) > 0 {
				content += fmt.Sprintf(" \t[%d attachments]", len(m.Attachments))
			}
			content += fmt.Sprintf(" \t<%s>", m.ID)

			if !history || config.MessageOutputHistory {
				addOut := ""
				if history && config.MessageOutputHistory && !m.Timestamp.IsZero() {
					addOut = fmt.Sprintf(" @ %s", discordSnowflakeToTimestamp(m.ID, "2006-01-02 15-04-05"))
				}
				if edited {
					log.Println(lg("Message", "", color.CyanString, "Edited [%s%s]: %s", sendLabel, addOut, content))
				} else {
					log.Println(lg("Message", "", color.CyanString, "[%s%s]: %s", sendLabel, addOut, content))
				}
			}
		}

		// Log Messages to File
		if sourceConfig.LogMessages != nil {
			if sourceConfig.LogMessages.Destination != "" {
				logPath := sourceConfig.LogMessages.Destination
				if *sourceConfig.LogMessages.DestinationIsFolder {
					if !strings.HasSuffix(logPath, string(os.PathSeparator)) {
						logPath += string(os.PathSeparator)
					}
					err := os.MkdirAll(logPath, 0755)
					if err == nil {
						logPath += "Log_Messages"
						if *sourceConfig.LogMessages.DivideLogsByServer {
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
						if *sourceConfig.LogMessages.DivideLogsByChannel {
							logPath += " CID_" + m.ChannelID
						}
						if *sourceConfig.LogMessages.DivideLogsByUser {
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
				if sourceConfig.LogMessages.FilterDuplicates != nil {
					if *sourceConfig.LogMessages.FilterDuplicates {
						if strings.Contains(currentLogS, fmt.Sprintf("[%s/%s/%s]", m.GuildID, m.ChannelID, m.ID)) {
							canLog = false
						}
					}
				}

				if canLog {
					// Writer
					f, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
					if err != nil {
						log.Println(lg("Message", "", color.RedString, "[sourceConfig.LogMessages] Failed to open log file:\t%s", err))
						f.Close()
					}
					defer f.Close()

					var newLine string
					// Prepend
					prefix := ""
					if sourceConfig.LogMessages.Prefix != nil {
						prefix = *sourceConfig.LogMessages.Prefix
					}
					// More Data
					additionalInfo := ""
					if sourceConfig.LogMessages.UserData != nil {
						if *sourceConfig.LogMessages.UserData {
							additionalInfo = fmt.Sprintf("[%s/%s/%s] \"%s\"#%s (%s) @ %s: ", m.GuildID, m.ChannelID, m.ID,
								m.Author.Username, m.Author.Discriminator, m.Author.ID,
								discordSnowflakeToTimestamp(m.ID, "2006-01-02 15-04-05"))
						}
					}
					if len(m.Attachments) > 0 {
						additionalInfo += fmt.Sprintf("<%d ATTACHMENTS> ", len(m.Attachments))
					}
					// Append
					suffix := ""
					if sourceConfig.LogMessages.Suffix != nil {
						suffix = *sourceConfig.LogMessages.Suffix
					}
					// New Line
					contentFmt, err := m.ContentWithMoreMentionsReplaced(bot)
					if err == nil {
						newLine += "\n" + prefix + additionalInfo + contentFmt + suffix
					} else {
						newLine += "\n" + prefix + additionalInfo + m.Content + suffix
					}

					if _, err = f.WriteString(newLine); err != nil {
						log.Println(lg("Message", "", color.RedString, "[sourceConfig.LogMessages] Failed to append file:\t%s", err))
					}
				}
			}
		}

		// Filters
		if sourceConfig.Filters != nil {
			shouldAbort := false

			if sourceConfig.Filters.AllowedPhrases != nil ||
				sourceConfig.Filters.AllowedUsers != nil ||
				sourceConfig.Filters.AllowedRoles != nil {
				shouldAbort = true
				if config.Debug {
					log.Println(lg("Debug", "Message", color.YellowString,
						"%s Filter will be ignoring by default...",
						color.HiMagentaString("(FILTER)")))
				}
			}

			if sourceConfig.Filters.BlockedPhrases != nil {
				for _, phrase := range *sourceConfig.Filters.BlockedPhrases {
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
			if sourceConfig.Filters.AllowedPhrases != nil {
				for _, phrase := range *sourceConfig.Filters.AllowedPhrases {
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

			if sourceConfig.Filters.BlockedUsers != nil {
				if stringInSlice(m.Author.ID, *sourceConfig.Filters.BlockedUsers) {
					shouldAbort = true
					if config.Debug {
						log.Println(lg("Debug", "Message", color.YellowString,
							"%s blockedUsers caught %s, planning to abort...",
							color.HiMagentaString("(FILTER)"), m.Author.ID))
					}
				}
			}
			if sourceConfig.Filters.AllowedUsers != nil {
				if stringInSlice(m.Author.ID, *sourceConfig.Filters.AllowedUsers) {
					shouldAbort = false
					if config.Debug {
						log.Println(lg("Debug", "Message", color.YellowString,
							"%s allowedUsers caught %s, planning to process...",
							color.HiMagentaString("(FILTER)"), m.Author.ID))
					}
				}
			}

			if sourceConfig.Filters.BlockedRoles != nil {
				member := m.Member
				if member == nil {
					member, _ = bot.GuildMember(m.GuildID, m.Author.ID)
				}
				if member != nil {
					for _, role := range member.Roles {
						if stringInSlice(role, *sourceConfig.Filters.BlockedRoles) {
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
			if sourceConfig.Filters.AllowedRoles != nil {
				member := m.Member
				if member == nil {
					member, _ = bot.GuildMember(m.GuildID, m.Author.ID)
				}
				if member != nil {
					for _, role := range member.Roles {
						if stringInSlice(role, *sourceConfig.Filters.AllowedRoles) {
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
				return nil
			}
		}

		// Delays
		delay := 0
		if history {
			if sourceConfig.DelayHandlingHistory != nil {
				delay = *sourceConfig.DelayHandlingHistory
			}
		} else {
			if sourceConfig.DelayHandling != nil {
				delay = *sourceConfig.DelayHandling
			}
		}
		if delay > 0 {
			if config.Debug {
				log.Println(lg("Debug", "Message", color.YellowString, "Delaying for %d milliseconds...", delay))
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		// Process Files
		var downloadedItems []downloadedItem
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
				Path:       sourceConfig.Destination,
				Message:    m,
				FileTime:   file.Time,
				HistoryCmd: history,
				EmojiCmd:   false,
				StartTime:  time.Now(),
			}.handleDownload()
			if status.Status == downloadSuccess {
				domain, _ := getDomain(file.Link)
				downloadedItems = append(downloadedItems, downloadedItem{
					URL:      file.Link,
					Domain:   domain,
					Filesize: filesize,
				})
			}
		}
		return downloadedItems
	}

	return nil
}

//#endregion
