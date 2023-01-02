package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/aidarkhanov/nanoid/v2"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/teris-io/shortid"
)

//#region Get

//TODO: Clean below

func getChannelState(channelID string) *discordgo.Channel {
	sourceChannel, _ := bot.State.Channel(channelID)
	if sourceChannel != nil {
		return sourceChannel
	}
	return &discordgo.Channel{}
}

func getGuildState(guildID string) *discordgo.Guild {
	sourceGuild, _ := bot.State.Guild(guildID)
	if sourceGuild != nil {
		return sourceGuild
	}
	return &discordgo.Guild{}
}

func getChannelGuildID(channelID string) string {
	sourceChannel, _ := bot.State.Channel(channelID)
	if sourceChannel != nil {
		return sourceChannel.GuildID
	}
	return ""
}

func getGuildName(guildID string) string {
	sourceGuildName := "UNKNOWN"
	sourceGuild, _ := bot.State.Guild(guildID)
	if sourceGuild != nil && sourceGuild.Name != "" {
		sourceGuildName = sourceGuild.Name
	}
	return sourceGuildName
}

func getChannelCategoryName(channelID string) string {
	sourceChannelName := "unknown"
	sourceChannel, _ := bot.State.Channel(channelID)
	if sourceChannel != nil {
		sourceParent, _ := bot.State.Channel(sourceChannel.ParentID)
		if sourceParent != nil {
			if sourceChannel.Name != "" {
				sourceChannelName = sourceParent.Name
			}
		}
	}
	return sourceChannelName
}

func getChannelName(channelID string) string {
	sourceChannelName := "unknown"
	sourceChannel, _ := bot.State.Channel(channelID)
	if sourceChannel != nil {
		if sourceChannel.Name != "" {
			sourceChannelName = sourceChannel.Name
		} else {
			switch sourceChannel.Type {
			case discordgo.ChannelTypeDM:
				sourceChannelName = "dm"
			case discordgo.ChannelTypeGroupDM:
				sourceChannelName = "group-dm"
			}
		}
	}
	return sourceChannelName
}

func getSourceName(guildID string, channelID string) string {
	guildName := getGuildName(guildID)
	channelName := getChannelName(channelID)
	if channelName == "dm" || channelName == "group-dm" {
		return channelName
	}
	return fmt.Sprintf("\"%s\"#%s", guildName, channelName)
}

//#endregion

//#region Time

const (
	discordEpoch = 1420070400000
)

//TODO: Clean these two

func discordTimestampToSnowflake(format string, timestamp string) string {
	if t, err := time.Parse(format, timestamp); err == nil {
		return fmt.Sprint(((t.Local().UnixNano() / int64(time.Millisecond)) - discordEpoch) << 22)
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
						guildID := m.GuildID
						// Replace message
						m = mCached
						// ^^
						if m.GuildID == "" && guildID != "" {
							m.GuildID = guildID
						}
						// Parse commands
						botCommands.FindAndExecute(bot, strings.ToLower(config.CommandPrefix), bot.State.User.ID, messageToLower(m))

						break
					}
				}
			} else if config.DebugOutput {
				log.Println(lg("Debug", "fixMessage",
					color.RedString, "%s, and an attempt to get channel messages found nothing...",
					ubIssue))
			}
		} else if config.DebugOutput {
			log.Println(lg("Debug", "fixMessage",
				color.HiRedString, "%s, and an attempt to get channel messages encountered an error:\t%s", ubIssue, err))
		}
	}
	if m.Content == "" && len(m.Attachments) == 0 && len(m.Embeds) == 0 {
		if config.DebugOutput && selfbot {
			log.Println(lg("Debug", "fixMessage",
				color.YellowString, "%s, and attempts to fix seem to have failed...", ubIssue))
		}
	}
	return m
}

//#endregion

//#region Presence

func dataKeyReplacement(input string) string {
	//TODO: Case-insensitive key replacement. -- If no streamlined way to do it, convert to lower to find substring location but replace normally
	if strings.Contains(input, "{{") && strings.Contains(input, "}}") {
		countInt := int64(dbDownloadCount()) + *config.InflateCount
		timeNow := time.Now()
		keys := [][]string{
			{"{{dgVersion}}", discordgo.VERSION},
			{"{{ddgVersion}}", projectVersion},
			{"{{apiVersion}}", discordgo.APIVersion},
			{"{{countNoCommas}}", fmt.Sprint(countInt)},
			{"{{count}}", formatNumber(countInt)},
			{"{{countShort}}", formatNumberShort(countInt)},
			{"{{numServers}}", fmt.Sprint(len(bot.State.Guilds))},
			{"{{numBoundChannels}}", fmt.Sprint(getBoundChannelsCount())},
			{"{{numBoundCategories}}", fmt.Sprint(getBoundCategoriesCount())},
			{"{{numBoundServers}}", fmt.Sprint(getBoundServersCount())},
			{"{{numBoundUsers}}", fmt.Sprint(getBoundUsersCount())},
			{"{{numAdminChannels}}", fmt.Sprint(len(config.AdminChannels))},
			{"{{numAdmins}}", fmt.Sprint(len(config.Admins))},
			{"{{timeSavedShort}}", timeLastUpdated.Format("3:04pm")},
			{"{{timeSavedShortTZ}}", timeLastUpdated.Format("3:04pm MST")},
			{"{{timeSavedMid}}", timeLastUpdated.Format("3:04pm MST 1/2/2006")},
			{"{{timeSavedLong}}", timeLastUpdated.Format("3:04:05pm MST - January 2, 2006")},
			{"{{timeSavedShort24}}", timeLastUpdated.Format("15:04")},
			{"{{timeSavedShortTZ24}}", timeLastUpdated.Format("15:04 MST")},
			{"{{timeSavedMid24}}", timeLastUpdated.Format("15:04 MST 2/1/2006")},
			{"{{timeSavedLong24}}", timeLastUpdated.Format("15:04:05 MST - 2 January, 2006")},
			{"{{timeNowShort}}", timeNow.Format("3:04pm")},
			{"{{timeNowShortTZ}}", timeNow.Format("3:04pm MST")},
			{"{{timeNowMid}}", timeNow.Format("3:04pm MST 1/2/2006")},
			{"{{timeNowLong}}", timeNow.Format("3:04:05pm MST - January 2, 2006")},
			{"{{timeNowShort24}}", timeNow.Format("15:04")},
			{"{{timeNowShortTZ24}}", timeNow.Format("15:04 MST")},
			{"{{timeNowMid24}}", timeNow.Format("15:04 MST 2/1/2006")},
			{"{{timeNowLong24}}", timeNow.Format("15:04:05 MST - 2 January, 2006")},
			{"{{uptime}}", durafmt.ParseShort(time.Since(startTime)).String()},
		}
		for _, key := range keys {
			if strings.Contains(input, key[0]) {
				input = strings.ReplaceAll(input, key[0], key[1])
			}
		}
	}
	return input
}

func channelKeyReplacement(input string, srcchannel string) string {
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
	return dataKeyReplacement(ret)
}

func dynamicKeyReplacement(channelConfig configurationSource, download downloadRequestStruct) string {
	//TODO: same as dataKeyReplacement

	ret := config.FilenameFormat
	if channelConfig.OverwriteFilenameFormat != nil {
		if *channelConfig.OverwriteFilenameFormat != "" {
			ret = *channelConfig.OverwriteFilenameFormat
		}
	}

	if strings.Contains(ret, "{{") && strings.Contains(ret, "}}") {

		// Format Filename Date
		filenameDateFormat := config.FilenameDateFormat
		if channelConfig.OverwriteFilenameDateFormat != nil {
			if *channelConfig.OverwriteFilenameDateFormat != "" {
				filenameDateFormat = *channelConfig.OverwriteFilenameDateFormat
			}
		}
		messageTime := download.Message.Timestamp

		shortID, err := shortid.Generate()
		if err != nil && config.DebugOutput {
			log.Println(lg("Debug", "dynamicKeyReplacement", color.HiCyanString, "Error when generating a shortID %s", err))
		}

		nanoID, err := nanoid.New()
		if err != nil && config.DebugOutput {
			log.Println(lg("Debug", "dynamicKeyReplacement", color.HiCyanString, "Error when creating a nanoID %s", err))
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
			{"{{fileType}}", download.FileExtension},
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
			{"{{downloadTime}}", durafmt.ParseShort(time.Since(download.StartTime)).String()},
			{"{{downloadTimeLong}}", durafmt.Parse(time.Since(download.StartTime)).String()},
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
	return dataKeyReplacement(ret)
}

func updateDiscordPresence() {
	if config.PresenceEnabled {
		// Vars
		countInt := int64(dbDownloadCount()) + *config.InflateCount
		count := formatNumber(countInt)
		countShort := formatNumberShort(countInt)
		timeShort := timeLastUpdated.Format("3:04pm")
		timeLong := timeLastUpdated.Format("3:04:05pm MST - January 2, 2006")

		// Defaults
		status := fmt.Sprintf("%s - %s files", timeShort, countShort)
		statusDetails := timeLong
		statusState := fmt.Sprintf("%s files total", count)

		// Overwrite Presence
		if config.PresenceOverwrite != nil {
			status = *config.PresenceOverwrite
			if status != "" {
				status = dataKeyReplacement(status)
			}
		}
		// Overwrite Details
		if config.PresenceOverwriteDetails != nil {
			statusDetails = *config.PresenceOverwriteDetails
			if statusDetails != "" {
				statusDetails = dataKeyReplacement(statusDetails)
			}
		}
		// Overwrite State
		if config.PresenceOverwriteState != nil {
			statusState = *config.PresenceOverwriteState
			if statusState != "" {
				statusState = dataKeyReplacement(statusState)
			}
		}

		// Update
		bot.UpdateStatusComplex(discordgo.UpdateStatusData{
			Game: &discordgo.Game{
				Name:    status,
				Type:    config.PresenceType,
				Details: statusDetails, // Only visible if real user
				State:   statusState,   // Only visible if real user
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
		channelConfig := getSource(channelID)
		if channelConfig.OverwriteEmbedColor != nil {
			if *channelConfig.OverwriteEmbedColor != "" {
				color = channelConfig.OverwriteEmbedColor
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
			if selfbot {
				return bot.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s **%s**\n\n%s", m.Author.Mention(), title, description))
			} else {
				return bot.ChannelMessageSendComplex(m.ChannelID,
					&discordgo.MessageSend{
						Content: m.Author.Mention(),
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
)

func sendStatusLabel(status sendStatusType) string {
	switch status {
	case sendStatusStartup:
		return "has launched"
	case sendStatusReconnect:
		return "has reconnected"
	case sendStatusExit:
		return "is exiting"
	}
	return "is confused"
}

func sendStatusMessage(status sendStatusType) {
	for _, adminChannel := range config.AdminChannels {
		if *adminChannel.LogStatus {
			var message string
			var label string

			//TODO: CLEAN
			if status == sendStatusStartup || status == sendStatusReconnect {
				label = "startup"
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
				if twitterConnected {
					message += "\n• Connected to Twitter API"
				}
				if googleDriveConnected {
					message += "\n• Connected to Google Drive"
				}
				message += fmt.Sprintf("\n_%s-%s %s / discordgo v%s (modified) / Discord API v%s_",
					runtime.GOOS, runtime.GOARCH, runtime.Version(), discordgo.VERSION, discordgo.APIVersion)
			} else if status == sendStatusExit {
				label = "exit"
				message += fmt.Sprintf("%s %s...\n", projectLabel, sendStatusLabel(status))
				message += fmt.Sprintf("\n• Uptime was %s", uptime())
				message += fmt.Sprintf("\n• %s total downloads", formatNumber(int64(dbDownloadCount())))
				message += fmt.Sprintf("\n• Bound to %d channel%s, %d categories, %d server%s, %d user%s",
					getBoundChannelsCount(), pluralS(getBoundChannelsCount()),
					getBoundCategoriesCount(),
					getBoundServersCount(), pluralS(getBoundServersCount()),
					getBoundUsersCount(), pluralS(getBoundUsersCount()),
				)
			}
			// Send
			if config.DebugOutput {
				log.Println(lg("Debug", "Status", color.HiCyanString, "Sending log for %s to admin channel %s",
					label, adminChannel.ChannelID))
			}
			if hasPerms(adminChannel.ChannelID, discordgo.PermissionEmbedLinks) && !selfbot {
				bot.ChannelMessageSendEmbed(adminChannel.ChannelID,
					buildEmbed(adminChannel.ChannelID, "Log — Status", message))
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
				if config.DebugOutput {
					log.Println(lg("Debug", "sendErrorMessage", color.HiCyanString, "Sending embed log for error to %s",
						adminChannel.ChannelID))
				}
				bot.ChannelMessageSendEmbed(adminChannel.ChannelID, buildEmbed(adminChannel.ChannelID, "Log — Error", err))
			} else if hasPerms(adminChannel.ChannelID, discordgo.PermissionSendMessages) {
				if config.DebugOutput {
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

// Checks if message author is a specified bot admin.
func isBotAdmin(m *discordgo.Message) bool {
	// No Admins or Admin Channels
	if len(config.Admins) == 0 && len(config.AdminChannels) == 0 {
		return true
	}
	// configurationAdminChannel.UnlockCommands Bypass
	if isAdminChannelRegistered(m.ChannelID) {
		channelConfig := getAdminChannelConfig(m.ChannelID)
		if *channelConfig.UnlockCommands == true {
			return true
		}
	}

	return m.Author.ID == botUser.ID || stringInSlice(m.Author.ID, config.Admins)
}

// Checks if message author is a specified bot admin OR is server admin OR has message management perms in channel
func isLocalAdmin(m *discordgo.Message) bool {
	if m == nil {
		if config.DebugOutput {
			log.Println(lg("Debug", "isLocalAdmin", color.YellowString, "check failed due to empty message"))
		}
		return true
	}
	sourceChannel, err := bot.State.Channel(m.ChannelID)
	if err != nil || sourceChannel == nil {
		if config.DebugOutput {
			log.Println(lg("Debug", "isLocalAdmin", color.YellowString,
				"check failed due to an error or received empty channel info for message:\t%s", err))
		}
		return true
	} else if sourceChannel.Name == "" || sourceChannel.GuildID == "" {
		if config.DebugOutput {
			log.Println(lg("Debug", "isLocalAdmin", color.YellowString,
				"check failed due to incomplete channel info"))
		}
		return true
	}

	guild, _ := bot.State.Guild(m.GuildID)
	localPerms, err := bot.State.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		if config.DebugOutput {
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
}

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
	return false
}

//#endregion

//#region Labeling

func getUserIdentifier(usr discordgo.User) string {
	return fmt.Sprintf("\"%s\"#%s", usr.Username, usr.Discriminator)
}

//#endregion
