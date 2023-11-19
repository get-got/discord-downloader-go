package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
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

//#region Getters

func getChannel(channelID string) (*discordgo.Channel, error) {
	channel, err := bot.Channel(channelID)
	if err != nil {
		channel, err = bot.State.Channel(channelID)
	}
	return channel, err
}

func getChannelErr(channelID string) error {
	_, errr := getChannel(channelID)
	return errr
}

func getServer(guildID string) (*discordgo.Guild, error) {
	guild, err := bot.Guild(guildID)
	if err != nil {
		guild, err = bot.State.Guild(guildID)
	}
	return guild, err
}

func getServerErr(guildID string) error {
	_, errr := getServer(guildID)
	return errr
}

//#endregion

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
	if usr.Discriminator == "0" {
		return "@" + usr.Username
	}
	return fmt.Sprintf("\"%s\"#%s", usr.Username, usr.Discriminator)
}

//#endregion

//#region Time

const (
	discordEpoch = 1420070400000
)

//TODO: Clean these two

func discordTimestampToSnowflake(format string, timestamp string) string {
	var snowflake string = ""
	var err error
	parsed, err := time.ParseInLocation(format, timestamp, time.Local)
	if err == nil {
		snowflake = fmt.Sprint(((parsed.UnixNano() / int64(time.Millisecond)) - discordEpoch) << 22)
	} else {
		log.Println(lg("Main", "", color.HiRedString,
			"Failed to convert timestamp to discord snowflake... Format: '%s', Timestamp: '%s' - Error:\t%s",
			format, timestamp, err))
	}
	return snowflake
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

	if sourceConfig == emptySourceConfig {
		return config.FilenameFormat
	}

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
			{"{{message}}", clearPathIllegalChars(download.Message.Content)},
			{"{{downloadTime}}", timeSinceShort(download.StartTime)},
			{"{{downloadTimeLong}}", timeSince(download.StartTime)},
			{"{{url}}", clearPathIllegalChars(download.InputURL)},
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

func dataKeys_DiscordMessage(input string, m *discordgo.Message) string {
	ret := input
	if strings.Contains(ret, "{{") && strings.Contains(ret, "}}") {
		// Basic message data
		keys := [][]string{
			{"{{year}}",
				fmt.Sprint(m.Timestamp.Year())},
			{"{{monthNum}}",
				fmt.Sprintf("%02d", m.Timestamp.Month())},
			{"{{dayOfMonth}}",
				fmt.Sprintf("%02d", m.Timestamp.Day())},
			{"{{hour}}",
				fmt.Sprintf("%02d", m.Timestamp.Hour())},
			{"{{minute}}",
				fmt.Sprintf("%02d", m.Timestamp.Minute())},
			{"{{second}}",
				fmt.Sprintf("%02d", m.Timestamp.Second())},
			{"{{timestamp}}", discordSnowflakeToTimestamp(m.ID, "2006-01-02 15-04-05")},
			{"{{timestampYYYYMMDD}}", discordSnowflakeToTimestamp(m.ID, "2006-01-02")},
			{"{{timestampHHMMSS}}", discordSnowflakeToTimestamp(m.ID, "15-04-05")},
			{"{{messageID}}", m.ID},
			{"{{channelID}}", m.ChannelID},
			{"{{serverID}}", m.GuildID},
		}
		// Author data if present
		if m.Author != nil {
			keys = append(keys, [][]string{
				{"{{userID}}", m.Author.ID},
				{"{{username}}", m.Author.Username},
				{"{{userDisc}}", m.Author.Discriminator},
			}...)
		}
		// Lookup server
		if srv, err := bot.Guild(m.GuildID); err == nil {
			keys = append(keys, [][]string{
				{"{{serverName}}", srv.Name},
			}...)
		}
		// Lookup channel
		var ch *discordgo.Channel = nil
		ch, err := bot.State.Channel(m.ChannelID)
		if err != nil {
			ch, _ = bot.Channel(m.ChannelID)
		}
		if ch != nil {
			keys = append(keys, [][]string{
				{"{{channelName}}", ch.Name},
				{"{{channelTopic}}", ch.Topic},
			}...)
			// Lookup parent channel
			if ch.ParentID != "" {
				if cat, err := bot.State.Channel(ch.ParentID); err == nil {
					if cat.Type == discordgo.ChannelTypeGuildCategory {
						keys = append(keys, [][]string{
							{"{{categoryID}}", cat.ID},
							{"{{categoryName}}", cat.Name},
						}...)
					} else if cat.Type == discordgo.ChannelTypeGuildText ||
						cat.Type == discordgo.ChannelTypeGuildForum ||
						cat.Type == discordgo.ChannelTypeGuildNews {
						keys = append(keys, [][]string{
							{"{{threadID}}", ch.ID},
							{"{{threadName}}", ch.Name},
							{"{{threadTopic}}", ch.Topic},
							{"{{forumID}}", cat.ID},
							{"{{forumName}}", cat.Name},
						}...)
					}
				}
			}
		}
		for _, key := range keys {
			if strings.Contains(ret, key[0]) {
				ret = strings.ReplaceAll(ret, key[0], key[1])
			}
		}
	}
	return ret
}

func dataKeys_DownloadStatus(input string, status downloadStatusStruct, download downloadRequestStruct) string {
	ret := input
	if strings.Contains(ret, "{{") && strings.Contains(ret, "}}") {
		// Basic message data
		keys := [][]string{
			{"{{downloadStatus}}", getDownloadStatusShort(status.Status)},
			{"{{downloadStatusLong}}", getDownloadStatus(status.Status)},
			{"{{downloadFilename}}", download.Filename},
			{"{{downloadExt}}", download.Extension},
			{"{{downloadPath}}", download.Path},
		}
		for _, key := range keys {
			if strings.Contains(ret, key[0]) {
				ret = strings.ReplaceAll(ret, key[0], key[1])
			}
		}
	}
	return ret
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

//#region Send Status Message

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
				log.Println(lg("Debug", "Bot Status", color.YellowString, "Sending log for %s to admin channel: %s",
					strings.ToUpper(label), getChannelLabel(adminChannel.ChannelID, nil)))
			}
			if hasPerms(adminChannel.ChannelID, discordgo.PermissionEmbedLinks) && !selfbot {
				bot.ChannelMessageSendEmbed(adminChannel.ChannelID,
					buildEmbed(adminChannel.ChannelID, emoji+" Log — Status", message))
			} else if hasPerms(adminChannel.ChannelID, discordgo.PermissionSendMessages) {
				bot.ChannelMessageSend(adminChannel.ChannelID, message)
			} else {
				log.Println(lg("Debug", "Bot Status", color.HiRedString, "Perms checks failed for sending %s status log to %s",
					strings.ToUpper(label), adminChannel.ChannelID))
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

//#region Download Emojis & Stickers

func downloadDiscordEmojis() {

	dataKeysEmoji := func(emoji discordgo.Emoji, serverID string) string {
		ret := config.EmojisFilenameFormat
		keys := [][]string{
			{"{{ID}}", emoji.ID},
			{"{{name}}", emoji.Name},
		}
		for _, key := range keys {
			if strings.Contains(ret, key[0]) {
				ret = strings.ReplaceAll(ret, key[0], key[1])
			}
		}
		return ret
	}

	if config.EmojisServers != nil {
		// Handle destination
		destination := "emojis"
		if config.EmojisDestination != nil {
			destination = *config.EmojisDestination
		}
		if err = os.MkdirAll(destination, 0755); err != nil {
			log.Println(lg("Discord", "Emojis", color.HiRedString, "Error while creating destination folder \"%s\": %s", destination, err))
		}
		// Start
		log.Println(lg("Discord", "Emojis", color.MagentaString, "Starting emoji downloads..."))
		for _, serverID := range *config.EmojisServers {
			emojis, err := bot.GuildEmojis(serverID)
			if err != nil {
				log.Println(lg("Discord", "Emojis", color.HiRedString, "Error fetching emojis from %s... %s", serverID, err))
			} else {
				guildName := "UNKNOWN"
				guild, err := bot.Guild(serverID)
				if err == nil {
					guildName = guild.Name
				}
				subfolder := destination + string(os.PathSeparator) + clearPathIllegalChars(guildName)
				if err = os.MkdirAll(subfolder, 0755); err != nil {
					log.Println(lg("Discord", "Emojis", color.HiRedString, "Error while creating subfolder \"%s\": %s", subfolder, err))
				}

				countDownloaded := 0
				countSkipped := 0
				countFailed := 0
				for _, emoji := range emojis {
					url := "https://cdn.discordapp.com/emojis/" + emoji.ID

					status, _ := downloadRequestStruct{
						InputURL:   url,
						Filename:   dataKeysEmoji(*emoji, serverID),
						Path:       subfolder,
						Message:    nil,
						FileTime:   time.Now(),
						HistoryCmd: false,
						EmojiCmd:   true,
						StartTime:  time.Now(),
					}.handleDownload()

					if status.Status == downloadSuccess {
						countDownloaded++
					} else if status.Status == downloadSkippedDuplicate {
						countSkipped++
					} else {
						countFailed++
						log.Println(lg("Discord", "Emojis", color.HiRedString,
							"Failed to download emoji \"%s\": \t[%d - %s] %v",
							url, status.Status, getDownloadStatus(status.Status), status.Error))
					}
				}

				// Log
				destinationOut := destination
				abs, err := filepath.Abs(destination)
				if err == nil {
					destinationOut = abs
				}
				log.Println(lg("Discord", "Emojis", color.HiMagentaString,
					fmt.Sprintf("%d emojis downloaded, %d skipped, %d failed - Destination: %s",
						countDownloaded, countSkipped, countFailed, destinationOut,
					)))
			}
		}
	}

}

func downloadDiscordStickers() {

	dataKeysSticker := func(sticker discordgo.Sticker) string {
		ret := config.StickersFilenameFormat
		keys := [][]string{
			{"{{ID}}", sticker.ID},
			{"{{name}}", sticker.Name},
		}
		for _, key := range keys {
			if strings.Contains(ret, key[0]) {
				ret = strings.ReplaceAll(ret, key[0], key[1])
			}
		}
		return ret
	}

	if config.StickersServers != nil {
		// Handle destination
		destination := "stickers"
		if config.StickersDestination != nil {
			destination = *config.StickersDestination
		}
		if err = os.MkdirAll(destination, 0755); err != nil {
			log.Println(lg("Discord", "Stickers", color.HiRedString, "Error while creating destination folder \"%s\": %s", destination, err))
		}
		log.Println(lg("Discord", "Stickers", color.MagentaString, "Starting sticker downloads..."))
		for _, serverID := range *config.StickersServers {
			guildName := "UNKNOWN"
			guild, err := bot.Guild(serverID)
			if err != nil {
				log.Println(lg("Discord", "Stickers", color.HiRedString, "Error fetching server %s... %s", serverID, err))
			} else {
				guildName = guild.Name
				subfolder := destination + string(os.PathSeparator) + clearPathIllegalChars(guildName)
				if err = os.MkdirAll(subfolder, 0755); err != nil {
					log.Println(lg("Discord", "Emojis", color.HiRedString, "Error while creating subfolder \"%s\": %s", subfolder, err))
				}

				countDownloaded := 0
				countSkipped := 0
				countFailed := 0
				for _, sticker := range guild.Stickers {
					url := "https://media.discordapp.net/stickers/" + sticker.ID

					status, _ := downloadRequestStruct{
						InputURL:   url,
						Filename:   dataKeysSticker(*sticker),
						Path:       subfolder,
						Message:    nil,
						FileTime:   time.Now(),
						HistoryCmd: false,
						EmojiCmd:   true,
						StartTime:  time.Now(),
					}.handleDownload()

					if status.Status == downloadSuccess {
						countDownloaded++
					} else if status.Status == downloadSkippedDuplicate {
						countSkipped++
					} else {
						countFailed++
						log.Println(lg("Discord", "Stickers", color.HiRedString,
							"Failed to download sticker \"%s\": \t[%d - %s] %v",
							url, status.Status, getDownloadStatus(status.Status), status.Error))
					}
				}

				// Log
				destinationOut := destination
				abs, err := filepath.Abs(destination)
				if err == nil {
					destinationOut = abs
				}
				log.Println(lg("Discord", "Stickers", color.HiMagentaString,
					fmt.Sprintf("%d stickers downloaded, %d skipped, %d failed - Destination: %s",
						countDownloaded, countSkipped, countFailed, destinationOut,
					)))
			}
		}
	}

}

//#endregion

//#region BOT LOGIN SEQUENCE

func botLoadDiscord() {
	var err error

	// Discord Login
	connectBot := func() {
		// Connect Bot
		bot.LogLevel = -1 // to ignore dumb wsapi error
		err = bot.Open()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "web socket already opened") {
			log.Println(lg("Discord", "", color.HiRedString, "Discord login failed:\t%s", err))
			properExit()
		}

		bot.LogLevel = config.DiscordLogLevel // reset
		bot.ShouldReconnectOnError = true
		dur, err := time.ParseDuration(fmt.Sprint(config.DiscordTimeout) + "s")
		if err != nil {
			dur, _ = time.ParseDuration("180s")
		}
		bot.Client.Timeout = dur

		bot.StateEnabled = true
		bot.State.MaxMessageCount = 100000
		bot.State.TrackChannels = true
		bot.State.TrackThreads = true
		bot.State.TrackMembers = true
		bot.State.TrackThreadMembers = true

		botUser, err = bot.User("@me")
		if err != nil {
			botUser = bot.State.User
		}
	}

	discord_login_count := 0
do_discord_login:
	discord_login_count++
	if discord_login_count > 1 {
		time.Sleep(3 * time.Second)
	}

	if config.Credentials.Token != "" && config.Credentials.Token != placeholderToken {
		// Login via Token (Bot or User)
		log.Println(lg("Discord", "", color.GreenString, "Connecting to Discord via Token..."))
		// attempt login without Bot prefix
		bot, err = discordgo.New(config.Credentials.Token)
		connectBot()
		if botUser.Bot { // is bot application, reconnect properly
			//log.Println(lg("Discord", "", color.GreenString, "Reconnecting as bot..."))
			bot, err = discordgo.New("Bot " + config.Credentials.Token)
		}

	} else if (config.Credentials.Email != "" && config.Credentials.Email != placeholderEmail) &&
		(config.Credentials.Password != "" && config.Credentials.Password != placeholderPassword) {
		// Login via Email+Password (User Only obviously)
		log.Println(lg("Discord", "", color.GreenString, "Connecting via Login..."))
		bot, err = discordgo.New(config.Credentials.Email, config.Credentials.Password)
	} else {
		if discord_login_count > 5 {
			log.Println(lg("Discord", "", color.HiRedString, "No valid credentials for Discord..."))
			properExit()
		} else {
			goto do_discord_login
		}
	}
	if err != nil {
		if discord_login_count > 5 {
			log.Println(lg("Discord", "", color.HiRedString, "Error logging in: %s", err))
			properExit()
		} else {
			goto do_discord_login
		}
	}

	connectBot()

	// Fetch Bot's User Info
	botUser, err = bot.User("@me")
	if err != nil {
		botUser = bot.State.User
		if botUser == nil {
			if discord_login_count > 5 {
				log.Println(lg("Discord", "", color.HiRedString, "Error obtaining user details: %s", err))
				properExit()
			} else {
				goto do_discord_login
			}
		}
	} else if botUser == nil {
		if discord_login_count > 5 {
			log.Println(lg("Discord", "", color.HiRedString, "No error encountered obtaining user details, but it's empty..."))
			properExit()
		} else {
			goto do_discord_login
		}
	} else {
		botReady = true
		log.Println(lg("Discord", "", color.HiGreenString, "Logged into %s", getUserIdentifier(*botUser)))
		if botUser.Bot {
			log.Println(lg("Discord", "Info", color.HiMagentaString, "GENUINE DISCORD BOT APPLICATION"))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ This is the safest way to use this bot."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ INTENTS: Make sure you have all 3 intents enabled for this bot in the Discord Developer Portal."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ PRESENCE: Details don't work. Only activity and status."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~ VISIBILITY: You can only see servers you have added the bot to, which requires you to be an admin or have an admin invite the bot."))
		} else {
			log.Println(lg("Discord", "Info", color.HiYellowString, "!!! USER ACCOUNT / SELF-BOT !!!"))
			log.Println(lg("Discord", "Info", color.HiMagentaString, "~ WARNING: Discord does NOT ALLOW automated user accounts (aka Self-Bots)."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~~~ By using this bot application with a user account, you potentially risk account termination."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~~~ See the GitHub page for link to Discord's official statement."))
			log.Println(lg("Discord", "Info", color.MagentaString, "~~~ IF YOU WISH TO AVOID THIS, USE A BOT APPLICATION IF POSSIBLE."))
			log.Println(lg("Discord", "Info", color.HiMagentaString, "~ DISCORD API BUGS MAY OCCUR - KNOWN ISSUES:"))
			log.Println(lg("Discord", "Info", color.MagentaString, "~~~ Can't see active threads, only archived threads."))
			log.Println(lg("Discord", "Info", color.HiMagentaString, "~ VISIBILITY: You can download from any channels/servers this account has access to."))
		}
	}
	if bot.State.User != nil { // is selfbot
		selfbot = bot.State.User.Email != ""
	}

	// Event Handlers
	botCommands = handleCommands()
	bot.AddHandler(messageCreate)
	bot.AddHandler(messageUpdate)

	// Start Presence
	timeLastUpdated = time.Now()
	go updateDiscordPresence()

	//(SV) Source Validation
	var invalidAdminChannels []string
	var invalidServers []string
	var invalidCategories []string
	var invalidChannels []string
	var missingPermsAdminChannels [][]string
	var missingPermsCategories [][]string
	var missingPermsChannels [][]string
	log.Println(lg("Discord", "Validation", color.GreenString, "Validating your configured Discord sources..."))

	validateSource := func(checkFunc func(string) error, target string, label string, invalidStack *[]string) bool {
		if err := checkFunc(target); err != nil {
			*invalidStack = append(*invalidStack, target)
			log.Println(lg("Discord", "Validation", color.HiRedString,
				"Bot cannot access %s %s...\t%s", label, target, err))
			return false
		}
		return true
	}
	checkChannelPerm := func(perm int64, permName string, target string, label string, invalidStack *[][]string) {
		if perms, err := bot.State.UserChannelPermissions(botUser.ID, target); err == nil {
			if perms&perm == 0 { // lacks permission
				*invalidStack = append(*invalidStack, []string{target, permName})
				log.Println(lg("Discord", "Validation", color.HiRedString,
					"%s %s / %s - Lacks <%s>...", strings.ToUpper(label), target, getChannelLabel(target, nil), permName))
			}
		} else if config.Debug {
			log.Println(lg("Discord", "Validation", color.HiRedString,
				"Encountered error checking Discord permission <%s> in %s %s / %s...\t%s",
				permName, label, target, getChannelLabel(target, nil), err))
		}
	}

	//(SV) Check Admin Channels
	if config.AdminChannels != nil {
		for _, adminChannel := range config.AdminChannels {
			if adminChannel.ChannelIDs != nil {
				for _, subchannel := range *adminChannel.ChannelIDs {
					if validateSource(getChannelErr, subchannel, "admin subchannel", &invalidAdminChannels) {
						checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
							subchannel, "admin subchannel", &missingPermsChannels)
						checkChannelPerm(discordgo.PermissionSendMessages, "PermissionSendMessages",
							subchannel, "admin subchannel", &missingPermsChannels)
						checkChannelPerm(discordgo.PermissionEmbedLinks, "PermissionEmbedLinks",
							subchannel, "admin subchannel", &missingPermsChannels)
					}
				}
			} else {
				if validateSource(getChannelErr, adminChannel.ChannelID, "admin channel", &invalidAdminChannels) {
					checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
						adminChannel.ChannelID, "admin channel", &missingPermsChannels)
					checkChannelPerm(discordgo.PermissionSendMessages, "PermissionSendMessages",
						adminChannel.ChannelID, "admin channel", &missingPermsChannels)
					checkChannelPerm(discordgo.PermissionEmbedLinks, "PermissionEmbedLinks",
						adminChannel.ChannelID, "admin channel", &missingPermsChannels)
				}
			}
		}
	}
	//(SV) Check "servers" config.Servers
	for _, server := range config.Servers {
		if server.ServerIDs != nil {
			for _, subserver := range *server.ServerIDs {
				if validateSource(getServerErr, subserver, "subserver", &invalidServers) {
					// tbd?
				}
			}
		} else {
			if validateSource(getServerErr, server.ServerID, "server", &invalidServers) {
				// tbd?
			}
		}
	}
	//(SV) Check "categories" config.Categories
	for _, category := range config.Categories {
		if category.CategoryIDs != nil {
			for _, subcategory := range *category.CategoryIDs {
				if validateSource(getChannelErr, subcategory, "subcategory", &invalidCategories) {
					checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
						subcategory, "subcategory", &missingPermsChannels)
				}
			}

		} else {
			if validateSource(getChannelErr, category.CategoryID, "category", &invalidCategories) {
				checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
					category.CategoryID, "category", &missingPermsChannels)
			}
		}
	}
	//(SV) Check "channels" config.Channels
	for _, channel := range config.Channels {
		if channel.ChannelIDs != nil {
			for _, subchannel := range *channel.ChannelIDs {
				if validateSource(getChannelErr, subchannel, "subchannel", &invalidChannels) {
					checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
						subchannel, "subchannel", &missingPermsChannels)
					checkChannelPerm(discordgo.PermissionReadMessageHistory, "PermissionReadMessageHistory",
						subchannel, "subchannel", &missingPermsChannels)
					checkChannelPerm(discordgo.PermissionEmbedLinks, "PermissionEmbedLinks",
						subchannel, "subchannel", &missingPermsChannels)
					if channel.ReactWhenDownloaded != nil {
						if *channel.ReactWhenDownloaded {
							checkChannelPerm(discordgo.PermissionAddReactions, "PermissionAddReactions",
								subchannel, "subchannel", &missingPermsChannels)
						}
					}
				}
			}

		} else {
			if validateSource(getChannelErr, channel.ChannelID, "channel", &invalidChannels) {
				checkChannelPerm(discordgo.PermissionViewChannel, "PermissionViewChannel",
					channel.ChannelID, "channel", &missingPermsChannels)
				checkChannelPerm(discordgo.PermissionReadMessageHistory, "PermissionReadMessageHistory",
					channel.ChannelID, "channel", &missingPermsChannels)
				checkChannelPerm(discordgo.PermissionEmbedLinks, "PermissionEmbedLinks",
					channel.ChannelID, "channel", &missingPermsChannels)
				if channel.ReactWhenDownloaded != nil {
					if *channel.ReactWhenDownloaded {
						checkChannelPerm(discordgo.PermissionAddReactions, "PermissionAddReactions",
							channel.ChannelID, "channel", &missingPermsChannels)
					}
				}
			}
		}
	}
	//(SV) NOTE: No validation for users because no way to do that by just user ID from what I've seen.

	//(SV) Output Invalid Sources
	invalidSources := len(invalidAdminChannels) + len(invalidChannels) + len(invalidCategories) + len(invalidServers)
	if invalidSources > 0 {
		log.Println(lg("Discord", "Validation", color.HiRedString,
			"Found %d invalid sources in configuration...", invalidSources))
		logMsg := fmt.Sprintf("Validation found %d invalid sources...\n", invalidSources)
		if len(invalidAdminChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Admin Channels: (%d)** - %s",
				len(invalidAdminChannels), strings.Join(invalidAdminChannels, ", "))
		}
		if len(invalidServers) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Servers: (%d)** - %s",
				len(invalidServers), strings.Join(invalidServers, ", "))
		}
		if len(invalidCategories) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Categories: (%d)** - %s",
				len(invalidCategories), strings.Join(invalidCategories, ", "))
		}
		if len(invalidChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Channels: (%d)** - %s",
				len(invalidChannels), strings.Join(invalidChannels, ", "))
		}
		sendErrorMessage(logMsg)
	} else {
		log.Println(lg("Discord", "Validation", color.HiGreenString,
			"No ID issues detected! Bot can see all configured Discord sources, but that doesn't check Discord permissions..."))
	}
	//(SV) Output Discord Permission Issues
	missingPermsSources := len(missingPermsAdminChannels) + len(missingPermsCategories) + len(missingPermsChannels)
	if missingPermsSources > 0 {
		log.Println(lg("Discord", "Validation", color.HiRedString,
			"Found %d sources with insufficient Discord permissions...", missingPermsSources))
		// removing this part for now due to multidimensional array change
		/*logMsg := fmt.Sprintf("Validation found %d sources with insufficient Discord permissions...\n", missingPermsSources)
		if len(missingPermsAdminChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Admin Channels: (%d)** - %s",
				len(missingPermsAdminChannels), strings.Join(missingPermsAdminChannels, ", "))
		}
		if len(missingPermsCategories) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Categories: (%d)** - %s",
				len(missingPermsCategories), strings.Join(missingPermsCategories, ", "))
		}
		if len(missingPermsAdminChannels) > 0 {
			logMsg += fmt.Sprintf("\n**- Download Channels: (%d)** - %s",
				len(missingPermsAdminChannels), strings.Join(missingPermsAdminChannels, ", "))
		}
		sendErrorMessage(logMsg)*/
	} else {
		log.Println(lg("Discord", "Validation", color.HiGreenString,
			"No permission issues detected! Bot seems to have all required Discord permissions."))
	}

	mainWg.Done()
}

//#endregion
