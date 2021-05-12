package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

const (
	DISCORD_EPOCH = 1420070400000
)

//TODO: Clean these two

func discordTimestampToSnowflake(format string, timestamp string) string {
	t, err := time.Parse(format, timestamp)
	if err == nil {
		return fmt.Sprint(((t.Local().UnixNano() / int64(time.Millisecond)) - DISCORD_EPOCH) << 22)
	}
	log.Println(color.HiRedString("Failed to convert timestamp to discord snowflake... Format: '%s', Timestamp: '%s' - Error:\t%s",
		format, timestamp, err),
	)
	return ""
}

func discordSnowflakeToTimestamp(snowflake string, format string) string {
	i, err := strconv.ParseInt(snowflake, 10, 64)
	if err != nil {
		return ""
	}
	t := time.Unix(0, ((i>>22)+DISCORD_EPOCH)*1000000)
	return t.Local().Format(format)
}

func getAllChannels() []string {
	var channels []string
	if config.All != nil {
		for _, guild := range bot.State.Guilds {
			for _, channel := range guild.Channels {
				if hasPerms(channel.ID, discordgo.PermissionReadMessages) && hasPerms(channel.ID, discordgo.PermissionReadMessageHistory) {
					channels = append(channels, channel.ID)
				}
			}
		}
	} else {
		for _, channel := range config.Channels {
			if channel.ChannelIDs != nil {
				for _, subchannel := range *channel.ChannelIDs {
					channels = append(channels, subchannel)
				}
			} else {
				channels = append(channels, channel.ChannelID)
			}
		}
	}
	return channels
}

//#region Presence

func presenceKeyReplacement(input string) string {
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
			{"{{numGuilds}}", fmt.Sprint(len(bot.State.Guilds))},
			{"{{numChannels}}", fmt.Sprint(len(config.Channels))},
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
				status = presenceKeyReplacement(status)
			}
		}
		// Overwrite Details
		if config.PresenceOverwriteDetails != nil {
			statusDetails = *config.PresenceOverwriteDetails
			if statusDetails != "" {
				statusDetails = presenceKeyReplacement(statusDetails)
			}
		}
		// Overwrite State
		if config.PresenceOverwriteState != nil {
			statusState = *config.PresenceOverwriteState
			if statusState != "" {
				statusState = presenceKeyReplacement(statusState)
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
	var color *string
	// Assign Defined Color
	if config.EmbedColor != nil {
		if *config.EmbedColor != "" {
			color = config.EmbedColor
		}
	}
	// Overwrite with Defined Color for Channel
	if isChannelRegistered(channelID) {
		channelConfig := getChannelConfig(channelID)
		if channelConfig.OverwriteEmbedColor != nil {
			if *channelConfig.OverwriteEmbedColor != "" {
				color = channelConfig.OverwriteEmbedColor
			}
		}
	}

	// Use Defined Color
	if color != nil {
		// Defined as Role, fetch role color
		if *color == "role" || *color == "user" {
			botColor := bot.State.UserColor(user.ID, channelID)
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
	if bot.State.UserColor(user.ID, channelID) != 0 {
		return bot.State.UserColor(user.ID, channelID)
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
	if hasPerms(m.ChannelID, discordgo.PermissionSendMessages) {
		return bot.ChannelMessageSendComplex(m.ChannelID,
			&discordgo.MessageSend{
				Content: m.Author.Mention(),
				Embed:   buildEmbed(m.ChannelID, title, description),
			},
		)
	} else {
		log.Println(color.HiRedString(fmtBotSendPerm, m.ChannelID))
		return nil, nil
	}
}

//#endregion

//#region Permissions

// Checks if message author is a specified bot admin.
func isBotAdmin(m *discordgo.Message) bool {
	return m.Author.ID == user.ID || stringInSlice(m.Author.ID, config.Admins)
}

// Checks if message author is a specified bot admin OR is server admin OR has message management perms in channel
func isLocalAdmin(m *discordgo.Message) bool {
	sourceChannel, err := bot.State.Channel(m.ChannelID)
	if err != nil || sourceChannel == nil {
		return true
	} else if sourceChannel.Name == "" || sourceChannel.GuildID == "" {
		return true
	}

	guild, _ := bot.State.Guild(m.GuildID)
	localPerms, _ := bot.State.UserChannelPermissions(m.Author.ID, m.ChannelID)

	botSelf := m.Author.ID == user.ID
	botAdmin := stringInSlice(m.Author.ID, config.Admins)
	guildOwner := m.Author.ID == guild.OwnerID
	guildAdmin := localPerms&discordgo.PermissionAdministrator > 0
	localManageMessages := localPerms&discordgo.PermissionManageMessages > 0

	return botSelf || botAdmin || guildOwner || guildAdmin || localManageMessages
}

func hasPerms(channelID string, permission int) bool {
	if !config.CheckPermissions {
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
			perms, err := bot.UserChannelPermissions(user.ID, channelID)
			if err == nil {
				return perms&permission == permission
			}
			log.Println(color.HiRedString("Failed to check permissions (%d) for %s:\t%s", permission, channelID, err))
		}
	}
	return false
}

//#endregion

//#region Labeling

func getUserIdentifier(usr discordgo.User) string {
	return fmt.Sprintf("\"%s\"#%s", usr.Username, usr.Discriminator)
}

func getGuildName(guildID string) string {
	sourceGuildName := "UNKNOWN"
	sourceGuild, _ := bot.State.Guild(guildID)
	if sourceGuild != nil && sourceGuild.Name != "" {
		sourceGuildName = sourceGuild.Name
	}
	return sourceGuildName
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

// For command case-insensitivity
func messageToLower(message *discordgo.Message) *discordgo.Message {
	newMessage := *message
	newMessage.Content = strings.ToLower(newMessage.Content)
	return &newMessage
}
