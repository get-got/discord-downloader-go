package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/aidarkhanov/nanoid/v2"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/teris-io/shortid"
)

const (
	fmtBotSendPerm = "Bot does not have permission to send messages in %s"
)

//#region Labels

func getServerLabel(serverID string) (displayLabel string) {
	displayLabel = "Discord"
	sourceGuild, err := bot.State.Guild(serverID)
	if err != nil {
		sourceGuild, _ = bot.Guild(serverID)
	}
	if sourceGuild != nil {
		if sourceGuild.Name != "" {
			displayLabel = sourceGuild.Name
		}
	}
	return displayLabel
}

func getCategoryLabel(channelID string) (displayLabel string) {
	displayLabel = "Category"
	sourceChannel, _ := bot.State.Channel(channelID)
	if sourceChannel != nil {
		sourceParent, _ := bot.State.Channel(sourceChannel.ParentID)
		if sourceParent != nil {
			if sourceChannel.Name != "" {
				displayLabel = sourceParent.Name
			}
		}
	}
	return displayLabel
}

func getChannelLabel(channelID string, channelData *discordgo.Channel) (displayLabel string) {
	displayLabel = channelID
	sourceChannel, err := bot.State.Channel(channelID)
	if err != nil {
		sourceChannel, _ = bot.Channel(channelID)
	}
	if channelData != nil {
		sourceChannel = channelData
	}
	if sourceChannel != nil {
		if sourceChannel.Name != "" {
			displayLabel = sourceChannel.Name
		} else if sourceChannel.Topic != "" {
			displayLabel = sourceChannel.Topic
		} else {
			switch sourceChannel.Type {
			case discordgo.ChannelTypeDM:
				displayLabel = "DM"
			case discordgo.ChannelTypeGroupDM:
				displayLabel = "Group-DM"
			}
		}
	}
	return displayLabel
}

func getUserIdentifier(usr discordgo.User) string {
	return fmt.Sprintf("\"%s\"#%s", usr.Username, usr.Discriminator)
}

//#endregion

//#region Time

const (
	discordEpoch = 1420070400000
)

//TODO: Clean these two

func discordTimestampToSnowflake(format string, timestamp string) string {
	if t, err := time.ParseInLocation(format, timestamp, time.Local); err == nil {
		return fmt.Sprint(((t.UnixNano() / int64(time.Millisecond)) - discordEpoch) << 22)
	}
	log.Println(lg("Main", "", color.HiRedString,
		"Failed to convert timestamp to discord snowflake... Format: '%s', Timestamp: '%s' - Error:\t%s",
		format, timestamp, err))
	return ""
}

func discordSnowflakeToTimestamp(snowflake string, format string) string {
	i, err := strconv.ParseInt(snowflake, 10, 64)
	if err != nil {
		return ""
	}
	t := time.Unix(0, ((i>>22)+discordEpoch)*1000000)
	return t.Local().Format(format)
}

//#endregion

//#region Messages

// For command case-insensitivity
func messageToLower(message *discordgo.Message) *discordgo.Message {
	newMessage := *message
	newMessage.Content = strings.ToLower(newMessage.Content)
	return &newMessage
}

func fixMessage(m *discordgo.Message) *discordgo.Message {
	// If message content is empty (likely due to userbot/selfbot)
	ubIssue := "Message is corrupted due to endpoint restriction"
	if m.Content == "" && len(m.Attachments) == 0 && len(m.Embeds) == 0 {
		// Get message history
		mCache, err := bot.ChannelMessages(m.ChannelID, 20, "", "", "")
		if err == nil {
			if len(mCache) > 0 {
				for _, mCached := range mCache {
					if mCached.ID == m.ID {
						// Fix original message having empty Guild ID
						serverID := m.GuildID
						// Replace message
						m = mCached
						// ^^
						if m.GuildID == "" && serverID != "" {
							m.GuildID = serverID
						}
						// Parse commands
						botCommands.FindAndExecute(bot, strings.ToLower(config.CommandPrefix), bot.State.User.ID, messageToLower(m))

						break
					}
				}
			} else if config.Debug {
				log.Println(lg("Debug", "fixMessage",
					color.RedString, "%s, and an attempt to get channel messages found nothing...",
					ubIssue))
			}
		} else if config.Debug {
			log.Println(lg("Debug", "fixMessage",
				color.HiRedString, "%s, and an attempt to get channel messages encountered an error:\t%s", ubIssue, err))
		}
	}
	if m.Content == "" && len(m.Attachments) == 0 && len(m.Embeds) == 0 {
		if config.Debug && selfbot {
			log.Println(lg("Debug", "fixMessage",
				color.YellowString, "%s, and attempts to fix seem to have failed...", ubIssue))
		}
	}
	return m
}

//#endregion

func channelDisplay(channelID string) (sourceName string, sourceChannelName string) {
	sourceChannelName = channelID
	sourceName = "UNKNOWN"
	sourceChannel, _ := bot.State.Channel(channelID)
	if sourceChannel != nil {
		// Channel Naming
		if sourceChannel.Name != "" {
			sourceChannelName = "#" + sourceChannel.Name // #example
		}
		switch sourceChannel.Type {
		case discordgo.ChannelTypeGuildText:
			// Server Naming
			if sourceChannel.GuildID != "" {
				sourceGuild, _ := bot.State.Guild(sourceChannel.GuildID)
				if sourceGuild != nil && sourceGuild.Name != "" {
					sourceName = sourceGuild.Name
				}
			}
			// Category Naming
			if sourceChannel.ParentID != "" {
				sourceParent, _ := bot.State.Channel(sourceChannel.ParentID)
				if sourceParent != nil {
					if sourceParent.Name != "" {
						sourceChannelName = sourceParent.Name + " - " + sourceChannelName
					}
				}
			}
		case discordgo.ChannelTypeDM:
			sourceName = "Direct Messages"
		case discordgo.ChannelTypeGroupDM:
			sourceName = "Group Messages"
		}
	}
	return sourceName, sourceChannelName
}

//#region Presence

func dataKeys(input string) string {
	//TODO: Case-insensitive key replacement. -- If no streamlined way to do it, convert to lower to find substring location but replace normally
	if strings.Contains(input, "{{") && strings.Contains(input, "}}") {
		countInt := int64(dbDownloadCount()) + *config.InflateDownloadCount
		timeNow := time.Now()
		keys := [][]string{
			{"{{dgVersion}}",
				discordgo.VERSION},
			{"{{ddgVersion}}",
				projectVersion},
			{"{{apiVersion}}",
				discordgo.APIVersion},
			{"{{countNoCommas}}",
				fmt.Sprint(countInt)},
			{"{{count}}",
				formatNumber(countInt)},
			{"{{countShort}}",
				formatNumberShort(countInt)},
			{"{{numServers}}",
				fmt.Sprint(len(bot.State.Guilds))},
			{"{{numBoundChannels}}",
				fmt.Sprint(getBoundChannelsCount())},
			{"{{numBoundCategories}}",
				fmt.Sprint(getBoundCategoriesCount())},
			{"{{numBoundServers}}",
				fmt.Sprint(getBoundServersCount())},
			{"{{numBoundUsers}}",
				fmt.Sprint(getBoundUsersCount())},
			{"{{numAdminChannels}}",
				fmt.Sprint(len(config.AdminChannels))},
			{"{{numAdmins}}",
				fmt.Sprint(len(config.Admins))},
			//TODO: redo time stuff
			{"{{timeSavedShort}}",
				timeLastUpdated.Format("3:04pm")},
			{"{{timeSavedShortTZ}}",
				timeLastUpdated.Format("3:04pm MST")},
			{"{{timeSavedMid}}",
				timeLastUpdated.Format("3:04pm MST 1/2/2006")},
			{"{{timeSavedLong}}",
				timeLastUpdated.Format("3:04:05pm MST - January 2, 2006")},
			{"{{timeSavedShort24}}",
				timeLastUpdated.Format("15:04")},
			{"{{timeSavedShortTZ24}}",
				timeLastUpdated.Format("15:04 MST")},
			{"{{timeSavedMid24}}",
				timeLastUpdated.Format("15:04 MST 2/1/2006")},
			{"{{timeSavedLong24}}",
				timeLastUpdated.Format("15:04:05 MST - 2 January, 2006")},
			{"{{timeNowShort}}",
				timeNow.Format("3:04pm")},
			{"{{timeNowShortTZ}}",
				timeNow.Format("3:04pm MST")},
			{"{{timeNowMid}}",
				timeNow.Format("3:04pm MST 1/2/2006")},
			{"{{timeNowLong}}",
				timeNow.Format("3:04:05pm MST - January 2, 2006")},
			{"{{timeNowShort24}}",
				timeNow.Format("15:04")},
			{"{{timeNowShortTZ24}}",
				timeNow.Format("15:04 MST")},
			{"{{timeNowMid24}}",
				timeNow.Format("15:04 MST 2/1/2006")},
			{"{{timeNowLong24}}",
				timeNow.Format("15:04:05 MST - 2 January, 2006")},
			{"{{uptime}}",
				timeSinceShort(startTime)},
		}
		for _, key := range keys {
			if strings.Contains(input, key[0]) {
				input = strings.ReplaceAll(input, key[0], key[1])
			}
		}
	}
	return input
}

func dataKeysChannel(input string, srcchannel string) string {
	ret := input
	if strings.Contains(ret, "{{") && strings.Contains(ret, "}}") {
		if channel, err := bot.State.Channel(srcchannel); err == nil {
			keys := [][]string{
				{"{{channelID}}", channel.ID},
				{"{{serverID}}", channel.GuildID},
				{"{{channelName}}", channel.Name},
			}
			for _, key := range keys {
				if strings.Contains(ret, key[0]) {
					ret = strings.ReplaceAll(ret, key[0], key[1])
				}
			}
		}
	}
	return dataKeys(ret)
}

func dataKeysDownload(sourceConfig configurationSource, download downloadRequestStruct) string {
	//TODO: same as dataKeys

	ret := config.FilenameFormat
	if sourceConfig.FilenameFormat != nil {
		if *sourceConfig.FilenameFormat != "" {
			ret = *sourceConfig.FilenameFormat
		}
	}

	if strings.Contains(ret, "{{") && strings.Contains(ret, "}}") {

		// Format Filename Date
		filenameDateFormat := config.FilenameDateFormat
		if sourceConfig.FilenameDateFormat != nil {
			if *sourceConfig.FilenameDateFormat != "" {
				filenameDateFormat = *sourceConfig.FilenameDateFormat
			}
		}
		messageTime := download.Message.Timestamp

		shortID, err := shortid.Generate()
		if err != nil && config.Debug {
			log.Println(lg("Debug", "dataKeysDownload", color.HiCyanString, "Error when generating a shortID %s", err))
		}

		nanoID, err := nanoid.New()
		if err != nil && config.Debug {
			log.Println(lg("Debug", "dataKeysDownload", color.HiCyanString, "Error when creating a nanoID %s", err))
		}

		userID := ""
		username := ""
		if download.Message.Author != nil {
			userID = download.Message.Author.ID
			username = download.Message.Author.Username
		}

		channelName := download.Message.ChannelID
		categoryID := download.Message.ChannelID
		categoryName := download.Message.ChannelID
		guildName := download.Message.GuildID
		if chinfo, err := bot.State.Channel(download.Message.ChannelID); err == nil {
			channelName = chinfo.Name
			categoryID = chinfo.ParentID
			if catinfo, err := bot.State.Channel(categoryID); err == nil {
				categoryName = catinfo.Name
			}
		}
		if guildinfo, err := bot.State.Guild(download.Message.GuildID); err == nil {
			guildName = guildinfo.Name
		}

		domain := "unknown"
		if parsedURL, err := url.Parse(download.InputURL); err == nil {
			domain = parsedURL.Hostname()
		}

		fileinfo, err := os.Stat(download.Path + download.Filename)
		filesize := "unknown"
		if err == nil {
			filesize = humanize.Bytes(uint64(fileinfo.Size()))
		}

		keys := [][]string{
			{"{{date}}", messageTime.Format(filenameDateFormat)},
			{"{{file}}", download.Filename},
			{"{{fileType}}", download.Extension},
			{"{{fileSize}}", filesize},
			{"{{messageID}}", download.Message.ID},
			{"{{userID}}", userID},
			{"{{username}}", username},
			{"{{channelID}}", download.Message.ChannelID},
			{"{{channelName}}", channelName},
			{"{{categoryID}}", categoryID},
			{"{{categoryName}}", categoryName},
			{"{{serverID}}", download.Message.GuildID},
			{"{{serverName}}", guildName},
			{"{{message}}", clearPath(download.Message.Content)},
			{"{{downloadTime}}", timeSinceShort(download.StartTime)},
			{"{{downloadTimeLong}}", timeSince(download.StartTime)},
			{"{{url}}", clearPath(download.InputURL)},
			{"{{domain}}", domain},
			{"{{nanoID}}", nanoID},
			{"{{shortID}}", shortID},
		}
		for _, key := range keys {
			if strings.Contains(ret, key[0]) {
				ret = strings.ReplaceAll(ret, key[0], key[1])
			}
		}
	}
	return dataKeys(ret)
}

func updateDiscordPresence() {
	if config.PresenceEnabled {
		// Vars
		countInt := int64(dbDownloadCount()) + *config.InflateDownloadCount
		count := formatNumber(countInt)
		countShort := formatNumberShort(countInt)
		timeShort := timeLastUpdated.Format("3:04pm")
		timeLong := timeLastUpdated.Format("3:04:05pm MST - January 2, 2006")

		// Defaults
		status := fmt.Sprintf("%s - %s files", timeShort, countShort)
		statusDetails := timeLong
		statusState := fmt.Sprintf("%s files total", count)

		// Overwrite Presence
		if config.PresenceLabel != nil {
			status = *config.PresenceLabel
			if status != "" {
				status = dataKeys(status)
			}
		}
		// Overwrite Details
		if config.PresenceDetails != nil {
			statusDetails = *config.PresenceDetails
			if statusDetails != "" {
				statusDetails = dataKeys(statusDetails)
			}
		}
		// Overwrite State
		if config.PresenceState != nil {
			statusState = *config.PresenceState
			if statusState != "" {
				statusState = dataKeys(statusState)
			}
		}

		// Update
		bot.UpdateStatusComplex(discordgo.UpdateStatusData{
			Game: &discordgo.Game{
				Name:    status,
				Type:    config.PresenceType,
				Details: statusDetails, // Only visible if real user
				State:   statusState,
			},
			Status: config.PresenceStatus,
		})
	} else if config.PresenceStatus != string(discordgo.StatusOnline) {
		bot.UpdateStatusComplex(discordgo.UpdateStatusData{
			Status: config.PresenceStatus,
		})
	}
}

//#endregion

//#region Embeds

func getEmbedColor(channelID string) int {
	var err error
	var color *string
	var channelInfo *discordgo.Channel

	// Assign Defined Color
	if config.EmbedColor != nil {
		if *config.EmbedColor != "" {
			color = config.EmbedColor
		}
	}
	// Overwrite with Defined Color for Channel
	/*var msg *discordgo.Message
	msg.ChannelID = channelID
	if channelRegistered(msg) {
		sourceConfig := getSource(channelID)
		if sourceConfig.OverwriteEmbedColor != nil {
			if *sourceConfig.OverwriteEmbedColor != "" {
				color = sourceConfig.OverwriteEmbedColor
			}
		}
	}*/

	// Use Defined Color
	if color != nil {
		// Defined as Role, fetch role color
		if *color == "role" || *color == "user" {
			botColor := bot.State.UserColor(botUser.ID, channelID)
			if botColor != 0 {
				return botColor
			}
			goto color_random
		}
		// Defined as Random, jump below (not preferred method but seems to work flawlessly)
		if *color == "random" || *color == "rand" {
			goto color_random
		}

		var colorString string = *color

		// Input is Hex
		colorString = strings.ReplaceAll(colorString, "#", "")
		if convertedHex, err := strconv.ParseUint(colorString, 16, 64); err == nil {
			return int(convertedHex)
		}

		// Input is Int
		if convertedInt, err := strconv.Atoi(colorString); err == nil {
			return convertedInt
		}

		// Definition is invalid since hasn't returned, so defaults to below...
	}

	// User color
	channelInfo, err = bot.State.Channel(channelID)
	if err == nil {
		if channelInfo.Type != discordgo.ChannelTypeDM && channelInfo.Type != discordgo.ChannelTypeGroupDM {
			if bot.State.UserColor(botUser.ID, channelID) != 0 {
				return bot.State.UserColor(botUser.ID, channelID)
			}
		}
	}

	// Random color
color_random:
	var randomColor string = randomcolor.GetRandomColorInHex()
	if convertedRandom, err := strconv.ParseUint(strings.ReplaceAll(randomColor, "#", ""), 16, 64); err == nil {
		return int(convertedRandom)
	}

	return 16777215 // white
}

// Shortcut function for quickly constructing a styled embed with Title & Description
func buildEmbed(channelID string, title string, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       getEmbedColor(channelID),
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: projectIcon,
			Text:    fmt.Sprintf("%s v%s", projectName, projectVersion),
		},
	}
}

// Shortcut function for quickly replying a styled embed with Title & Description
func replyEmbed(m *discordgo.Message, title string, description string) (*discordgo.Message, error) {
	if m != nil {
		if hasPerms(m.ChannelID, discordgo.PermissionSendMessages) {
			mention := m.Author.Mention()
			if !config.CommandTagging { // Erase mention if tagging disabled
				mention = ""
			}
			if selfbot {
				if mention != "" { // Add space if mentioning
					mention += " "
				}
				return bot.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s**%s**\n\n%s", mention, title, description))
			} else {
				return bot.ChannelMessageSendComplex(m.ChannelID,
					&discordgo.MessageSend{
						Content: mention,
						Embed:   buildEmbed(m.ChannelID, title, description),
					},
				)
			}
		}
		log.Println(lg("Discord", "replyEmbed", color.HiRedString, fmtBotSendPerm, m.ChannelID))
	}
	return nil, nil
}

//#endregion

//#region Send Status

type sendStatusType int

const (
	sendStatusStartup sendStatusType = iota
	sendStatusReconnect
	sendStatusExit
	sendStatusSettings
)

func sendStatusLabel(status sendStatusType) string {
	switch status {
	case sendStatusStartup:
		return "has launched"
	case sendStatusReconnect:
		return "has reconnected"
	case sendStatusExit:
		return "is exiting"
	case sendStatusSettings:
		return "updated settings"
	}
	return "is confused"
}

func sendStatusMessage(status sendStatusType) {
	for _, adminChannel := range config.AdminChannels {
		if *adminChannel.LogStatus {
			var message string
			var label string
			var emoji string

			//TODO: CLEAN
			if status == sendStatusStartup || status == sendStatusReconnect {
				label = "startup"
				emoji = "🟩"
				if status == sendStatusReconnect {
					emoji = "🟧"
				}
				message += fmt.Sprintf("%s %s and connected to %d server%s...\n", projectLabel, sendStatusLabel(status), len(bot.State.Guilds), pluralS(len(bot.State.Guilds)))
				message += fmt.Sprintf("\n• Uptime is %s", uptime())
				message += fmt.Sprintf("\n• %s total downloads", formatNumber(int64(dbDownloadCount())))
				message += fmt.Sprintf("\n• Bound to %d channel%s, %d categories, %d server%s, %d user%s",
					getBoundChannelsCount(), pluralS(getBoundChannelsCount()),
					getBoundCategoriesCount(),
					getBoundServersCount(), pluralS(getBoundServersCount()),
					getBoundUsersCount(), pluralS(getBoundUsersCount()),
				)
				if config.All != nil {
					message += "\n• **ALL MODE ENABLED -** Bot will use all available channels"
				}
				allChannels := getAllRegisteredChannels()
				message += fmt.Sprintf("\n• ***Listening to %s channel%s...***\n", formatNumber(int64(len(allChannels))), pluralS(len(allChannels)))
				message += fmt.Sprintf("\n_%s_", versions(true))
			} else if status == sendStatusExit {
				label = "exit"
				emoji = "🟥"
				message += fmt.Sprintf("%s %s...\n", projectLabel, sendStatusLabel(status))
				message += fmt.Sprintf("\n• Uptime was %s", uptime())
				message += fmt.Sprintf("\n• %s total downloads", formatNumber(int64(dbDownloadCount())))
				message += fmt.Sprintf("\n• Bound to %d channel%s, %d categories, %d server%s, %d user%s",
					getBoundChannelsCount(), pluralS(getBoundChannelsCount()),
					getBoundCategoriesCount(),
					getBoundServersCount(), pluralS(getBoundServersCount()),
					getBoundUsersCount(), pluralS(getBoundUsersCount()),
				)
			} else if status == sendStatusSettings {
				label = "settings"
				emoji = "🟨"
				message += fmt.Sprintf("%s %s...\n", projectLabel, sendStatusLabel(status))
				message += fmt.Sprintf("\n• Bound to %d channel%s, %d categories, %d server%s, %d user%s",
					getBoundChannelsCount(), pluralS(getBoundChannelsCount()),
					getBoundCategoriesCount(),
					getBoundServersCount(), pluralS(getBoundServersCount()),
					getBoundUsersCount(), pluralS(getBoundUsersCount()),
				)
			}
			// Send
			if config.Debug {
				log.Println(lg("Debug", "Status", color.HiCyanString, "Sending log for %s to admin channel %s",
					label, adminChannel.ChannelID))
			}
			if hasPerms(adminChannel.ChannelID, discordgo.PermissionEmbedLinks) && !selfbot {
				bot.ChannelMessageSendEmbed(adminChannel.ChannelID,
					buildEmbed(adminChannel.ChannelID, emoji+" Log — Status", message))
			} else if hasPerms(adminChannel.ChannelID, discordgo.PermissionSendMessages) {
				bot.ChannelMessageSend(adminChannel.ChannelID, message)
			} else {
				log.Println(lg("Debug", "Status", color.HiRedString, "Perms checks failed for sending status log to %s",
					adminChannel.ChannelID))
			}
		}
	}
}

func sendErrorMessage(err string) {
	for _, adminChannel := range config.AdminChannels {
		if *adminChannel.LogErrors {
			// Send
			if hasPerms(adminChannel.ChannelID, discordgo.PermissionEmbedLinks) && !selfbot { // not confident this is the right permission
				if config.Debug {
					log.Println(lg("Debug", "sendErrorMessage", color.HiCyanString, "Sending embed log for error to %s",
						adminChannel.ChannelID))
				}
				bot.ChannelMessageSendEmbed(adminChannel.ChannelID, buildEmbed(adminChannel.ChannelID, "Log — Error", err))
			} else if hasPerms(adminChannel.ChannelID, discordgo.PermissionSendMessages) {
				if config.Debug {
					log.Println(lg("Debug", "sendErrorMessage", color.HiCyanString, "Sending embed log for error to %s",
						adminChannel.ChannelID))
				}
				bot.ChannelMessageSend(adminChannel.ChannelID, err)
			} else {
				log.Println(lg("Debug", "sendErrorMessage", color.HiRedString, "Perms checks failed for sending error log to %s",
					adminChannel.ChannelID))
			}
		}
	}
}

//#endregion

//#region Permissions

// Checks if message author is a specified bot admin OR is server admin OR has message management perms in channel
/*func isLocalAdmin(m *discordgo.Message) bool {
	if m == nil {
		if config.Debug {
			log.Println(lg("Debug", "isLocalAdmin", color.YellowString, "check failed due to empty message"))
		}
		return true
	}
	sourceChannel, err := bot.State.Channel(m.ChannelID)
	if err != nil || sourceChannel == nil {
		if config.Debug {
			log.Println(lg("Debug", "isLocalAdmin", color.YellowString,
				"check failed due to an error or received empty channel info for message:\t%s", err))
		}
		return true
	} else if sourceChannel.Name == "" || sourceChannel.GuildID == "" {
		if config.Debug {
			log.Println(lg("Debug", "isLocalAdmin", color.YellowString,
				"check failed due to incomplete channel info"))
		}
		return true
	}

	guild, _ := bot.State.Guild(m.GuildID)
	localPerms, err := bot.State.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		if config.Debug {
			log.Println(lg("Debug", "isLocalAdmin", color.YellowString,
				"check failed due to error when checking permissions:\t%s", err))
		}
		return true
	}

	botSelf := m.Author.ID == botUser.ID
	botAdmin := stringInSlice(m.Author.ID, config.Admins)
	guildOwner := m.Author.ID == guild.OwnerID
	guildAdmin := localPerms&discordgo.PermissionAdministrator > 0
	localManageMessages := localPerms&discordgo.PermissionManageMessages > 0

	return botSelf || botAdmin || guildOwner || guildAdmin || localManageMessages
}*/

func hasPerms(channelID string, permission int64) bool {
	if selfbot {
		return true
	}

	sourceChannel, err := bot.State.Channel(channelID)
	if sourceChannel != nil && err == nil {
		switch sourceChannel.Type {
		case discordgo.ChannelTypeDM:
			return true
		case discordgo.ChannelTypeGroupDM:
			return true
		case discordgo.ChannelTypeGuildText:
			perms, err := bot.UserChannelPermissions(botUser.ID, channelID)
			if err == nil {
				return perms&permission == permission
			}
			log.Println(lg("Debug", "hasPerms", color.HiRedString,
				"Failed to check permissions (%d) for %s:\t%s", permission, channelID, err))
		}
	}
	return true
}

//#endregion
