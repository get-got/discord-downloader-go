package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
)

func presenceKeyReplacement(input string) string {
	countInt := int64(dbDownloadCount()) + *config.InflateCount
	timeNow := time.Now()
	keys := [][]string{
		{"{{dgVersion}}", discordgo.VERSION},
		{"{{ddgVersion}}", PROJECT_VERSION},
		{"{{apiVersion}}", discordgo.APIVersion},
		{"{{count}}", formatNumber(countInt)},
		{"{{countShort}}", formatNumberShort(countInt)},
		{"{{numGuilds}}", fmt.Sprint(len(bot.State.Guilds))},
		{"{{numChannels}}", fmt.Sprint(len(config.Channels))},
		{"{{numAdminChannels}}", fmt.Sprint(len(config.AdminChannels))},
		{"{{numAdmins}}", fmt.Sprint(len(config.Admins))},
		{"{{timeSavedShort}}", timeLastUpdated.Format("3:04pm")},
		{"{{timeSavedLong}}", timeLastUpdated.Format("3:04:05pm MST - January 1, 2006")},
		{"{{timeSavedShort24}}", timeLastUpdated.Format("15:04")},
		{"{{timeSavedLong24}}", timeLastUpdated.Format("15:04:05 MST - 1 January, 2006")},
		{"{{timeNowShort}}", timeNow.Format("3:04pm")},
		{"{{timeNowLong}}", timeNow.Format("3:04:05pm MST - January 1, 2006")},
		{"{{timeNowShort24}}", timeNow.Format("15:04")},
		{"{{timeNowLong24}}", timeNow.Format("15:04:05 MST - 1 January, 2006")},
		{"{{uptime}}", durafmt.ParseShort(time.Since(startTime)).String()},
	}
	//TODO: Case-insensitive key replacement.
	if strings.Contains(input, "{{") && strings.Contains(input, "}}") {
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
		timeLong := timeLastUpdated.Format("3:04:05pm MST - January 1, 2006")

		// Defaults
		status_presence := fmt.Sprintf("%s - %s files", timeShort, countShort)
		status_details := timeLong
		status_state := fmt.Sprintf("%s files total", count)

		// Overwrite Presence
		if config.PresenceOverwrite != nil {
			status_presence = *config.PresenceOverwrite
			if status_presence != "" {
				status_presence = presenceKeyReplacement(status_presence)
			}
		}
		// Overwrite Details
		if config.PresenceOverwriteDetails != nil {
			status_details = *config.PresenceOverwriteDetails
			if status_details != "" {
				status_details = presenceKeyReplacement(status_details)
			}
		}
		// Overwrite State
		if config.PresenceOverwriteState != nil {
			status_state = *config.PresenceOverwriteState
			if status_state != "" {
				status_state = presenceKeyReplacement(status_state)
			}
		}

		// Update
		bot.UpdateStatusComplex(discordgo.UpdateStatusData{
			Game: &discordgo.Game{
				Name:    status_presence,
				Type:    config.PresenceType,
				Details: status_details, // Only visible if real user
				State:   status_state,   // Only visible if real user
			},
			Status: config.PresenceStatus,
		})
	}
}

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
			IconURL: PROJECT_ICON,
			Text:    fmt.Sprintf("%s v%s — discordgo v%s", PROJECT_NAME, PROJECT_VERSION, discordgo.VERSION),
		},
	}
}

// Shortcut function for quickly replying a styled embed with Title & Description
func replyEmbed(m *discordgo.Message, title string, description string) (*discordgo.Message, error) {
	return bot.ChannelMessageSendComplex(m.ChannelID,
		&discordgo.MessageSend{
			Content: m.Author.Mention(),
			Embed:   buildEmbed(m.ChannelID, title, description),
		},
	)
}

// Checks if message author is a specified bot admin.
func isBotAdmin(m *discordgo.Message) bool {
	return m.Author.ID == user.ID || stringInSlice(m.Author.ID, config.Admins)
}

// Checks if message author is a specified bot admin OR is server admin OR has message management perms in channel
func isLocalAdmin(m *discordgo.Message) bool {
	guild, _ := bot.State.Guild(m.GuildID)
	localPerms, _ := bot.State.UserChannelPermissions(m.Author.ID, m.ChannelID)

	botSelf := m.Author.ID == user.ID
	botAdmin := stringInSlice(m.Author.ID, config.Admins)
	guildOwner := m.Author.ID == guild.OwnerID
	guildAdmin := localPerms&discordgo.PermissionAdministrator > 0
	localManageMessages := localPerms&discordgo.PermissionManageMessages > 0

	return botSelf || botAdmin || guildOwner || guildAdmin || localManageMessages
}

func getUserIdentifier(usr discordgo.User) string {
	return fmt.Sprintf("\"%s\"#%s", usr.Username, usr.Discriminator)
}

func getGuildName(guildID string) string {
	sourceGuildName := "Server Name Unknown"
	sourceGuild, _ := bot.State.Guild(guildID)
	if sourceGuild != nil && sourceGuild.Name != "" {
		sourceGuildName = sourceGuild.Name
	}
	return sourceGuildName
}

func getChannelName(channelID string) string {
	sourceChannelName := "Channel Name Unknown"
	sourceChannel, _ := bot.State.Channel(channelID)
	if sourceChannel != nil && sourceChannel.Name != "" {
		sourceChannelName = sourceChannel.Name
	}
	return sourceChannelName
}

// For command case-insensitivity
func messageToLower(message *discordgo.Message) *discordgo.Message {
	newMessage := *message
	newMessage.Content = strings.ToLower(newMessage.Content)
	return &newMessage
}
