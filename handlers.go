package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

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
	shouldBail := false //TODO: this is messy, overlapped purpose with shouldAbort used for filters down below in this func.
	shouldBailReason := ""
	// Ignore own messages unless told not to
	if m.Author.ID == botUser.ID && !config.ScanOwnMessages {
		shouldBail = true
		shouldBailReason = "config.ScanOwnMessages"
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
	if sourceConfig := getSource(m); sourceConfig != emptySourceConfig {
		// Ignore bots if told to do so
		if m.Author.Bot && *sourceConfig.IgnoreBots {
			shouldBail = true
			shouldBailReason = "config.IgnoreBots"
		}
		// Ignore if told so by config
		if (!history && !*sourceConfig.Enabled) || (edited && !*sourceConfig.ScanEdits) {
			shouldBail = true
			shouldBailReason = "config.ScanEdits"
		}

		// Bail due to basic config rules
		if shouldBail {
			if config.Debug {
				log.Println(lg("Debug", "Message", color.YellowString,
					"%s Ignoring message due to %s...", color.HiMagentaString("(CONFIG)"), shouldBailReason))
			}
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
				encounteredErrors := false
				savePath := sourceConfig.LogMessages.Destination + string(os.PathSeparator)

				// Subfolder Division - Format Subfolders
				if sourceConfig.LogMessages.Subfolders != nil {
					subfolders := []string{}
					for _, subfolder := range *sourceConfig.LogMessages.Subfolders {
						newSubfolder := dataKeys_DiscordMessage(subfolder, m)

						// Scrub subfolder
						newSubfolder = clearSourceLogField(newSubfolder, *sourceConfig.LogMessages)

						// Do Fallback if a line contains an unparsed key (if fallback exists).
						if strings.Contains(newSubfolder, "{{") && strings.Contains(newSubfolder, "}}") &&
							sourceConfig.LogMessages.SubfoldersFallback != nil {
							subfolders = []string{}
							for _, subfolder2 := range *sourceConfig.LogMessages.SubfoldersFallback {
								newSubfolder2 := dataKeys_DiscordMessage(subfolder2, m)

								// Scrub subfolder
								newSubfolder2 = clearSourceLogField(newSubfolder2, *sourceConfig.LogMessages)

								subfolders = append(subfolders, newSubfolder2)
							}
							break
						} else {
							subfolders = append(subfolders, newSubfolder)
						}
					}

					// Subfolder Dividion - Handle Formatted Subfolders
					subpath := ""
					for _, subfolder := range subfolders {
						subpath = subpath + subfolder + string(os.PathSeparator)
						// Create folder
						if err := os.MkdirAll(filepath.Clean(savePath+subpath), 0755); err != nil {
							log.Println(lg("LogMessages", "", color.HiRedString,
								"Error while creating subfolder \"%s\": %s", savePath+subpath, err))
							encounteredErrors = true
						}
					}
					// Format Path
					savePath = filepath.Clean(savePath + string(os.PathSeparator) + subpath) // overwrite with new destination path
				}

				if !encounteredErrors {
					if _, err := os.Stat(savePath); err != nil {
						log.Println(lg("LogMessages", "", color.HiRedString,
							"Save path %s is invalid... %s", savePath, err))
					} else {
						// Format filename
						filename := m.ChannelID + ".txt"
						if sourceConfig.LogMessages.FilenameFormat != nil {
							if *sourceConfig.LogMessages.FilenameFormat != "" {
								filename = dataKeys_DiscordMessage(*sourceConfig.LogMessages.FilenameFormat, m)
								// if extension presumed missing
								if !strings.Contains(filename, ".") {
									filename += ".txt"
								}
							}
						}

						// Scrub filename
						filename = clearSourceLogField(filename, *sourceConfig.LogMessages)

						// Build path
						logPath := filepath.Clean(savePath + string(os.PathSeparator) + filename)

						// Prepend
						prefix := ""
						if sourceConfig.LogMessages.LinePrefix != nil {
							prefix = *sourceConfig.LogMessages.LinePrefix
						}
						prefix = dataKeys_DiscordMessage(prefix, m)

						// Append
						suffix := ""
						if sourceConfig.LogMessages.LineSuffix != nil {
							suffix = *sourceConfig.LogMessages.LineSuffix
						}
						suffix = dataKeys_DiscordMessage(suffix, m)

						// New Line
						var newLine string
						msgContent := m.Content
						if contentFmt, err := m.ContentWithMoreMentionsReplaced(bot); err == nil {
							msgContent = contentFmt
						}
						lineContent := msgContent
						if sourceConfig.LogMessages.LineContent != nil {
							lineContent = *sourceConfig.LogMessages.LineContent
						}
						keys := [][]string{
							{"{{message}}", msgContent},
						}
						for _, key := range keys {
							if strings.Contains(lineContent, key[0]) {
								lineContent = strings.ReplaceAll(lineContent, key[0], key[1])
							}
						}
						newLine += "\n" + prefix + lineContent + suffix

						// Read
						currentLog := ""
						if logfile, err := os.ReadFile(logPath); err == nil {
							currentLog = string(logfile)
						}
						canLog := true
						// Filter Duplicates
						if sourceConfig.LogMessages.FilterDuplicates != nil {
							if *sourceConfig.LogMessages.FilterDuplicates {
								if strings.Contains(currentLog, newLine) {
									canLog = false
								}
							}
						}

						if canLog {
							// Writer
							f, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
							if err != nil {
								log.Println(lg("LogMessages", "", color.RedString, "[sourceConfig.LogMessages] Failed to open log file:\t%s", err))
								f.Close()
							}
							defer f.Close()

							if _, err = f.WriteString(newLine); err != nil {
								log.Println(lg("Message", "", color.RedString, "[sourceConfig.LogMessages] Failed to append file:\t%s", err))
							}
						}
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

		// Process Collected Links
		var downloadedItems []downloadedItem
		files := getLinksByMessage(m)
		for _, file := range files {
			// Blank link?
			if file.Link == "" {
				continue
			}
			if (*sourceConfig.IgnoreEmojis && strings.HasPrefix(file.Link, "https://cdn.discordapp.com/emojis/")) ||
				(*sourceConfig.IgnoreStickers && strings.HasPrefix(file.Link, "https://media.discordapp.net/stickers/")) {
				continue
			}
			// Filter Checks
			shouldAbort := false
			if sourceConfig.Filters.BlockedLinkContent != nil {
				for _, phrase := range *sourceConfig.Filters.BlockedLinkContent {
					if strings.Contains(file.Link, phrase) && phrase != "" {
						shouldAbort = true
						if config.Debug {
							log.Println(lg("Debug", "Message", color.YellowString,
								"%s blockedLinkContent found \"%s\" in link, planning to abort...",
								color.HiMagentaString("(FILTER)"), phrase))
						}
						break
					}
				}
			}
			if sourceConfig.Filters.AllowedLinkContent != nil {
				for _, phrase := range *sourceConfig.Filters.AllowedLinkContent {
					if strings.Contains(file.Link, phrase) && phrase != "" {
						shouldAbort = false
						if config.Debug {
							log.Println(lg("Debug", "Message", color.YellowString,
								"%s allowedLinkContent found \"%s\" in link, planning to process...",
								color.HiMagentaString("(FILTER)"), phrase))
						}
						break
					}
				}
			}
			if shouldAbort {
				if config.Debug {
					log.Println(lg("Debug", "Message", color.YellowString,
						"%s Filter decided to ignore link...",
						color.HiMagentaString("(FILTER)")))
				}
				continue
			}
			// Output
			if config.Debug && (!history || config.MessageOutputHistory) {
				log.Println(lg("Debug", "Message", color.HiCyanString, "FOUND FILE: "+file.Link+fmt.Sprintf(" \t<%s>", m.ID)))
			}
			// Handle Download
			status, filesize := downloadRequestStruct{
				InputURL:     file.Link,
				Filename:     file.Filename,
				Path:         sourceConfig.Destination,
				Message:      m,
				Channel:      c,
				FileTime:     file.Time,
				HistoryCmd:   history,
				EmojiCmd:     false,
				StartTime:    time.Now(),
				AttachmentID: file.AttachmentID,
			}.handleDownload()
			// Await Status
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
