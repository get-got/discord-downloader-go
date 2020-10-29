package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/bwmarrin/discordgo"
)

func userDisplay(usr discordgo.User) string {
	return fmt.Sprintf("\"%s\"#%s", usr.Username, usr.Discriminator)
}

func adminCheck(m *discordgo.Message) bool {
	return m.Author.ID == user.ID || stringInSlice(m.Author.ID, config.Admins)
}

func adminCheckLocal(m *discordgo.Message) bool {
	guild, _ := bot.State.Guild(m.GuildID)
	localPerms, _ := bot.State.UserChannelPermissions(m.Author.ID, m.ChannelID)

	botSelf := m.Author.ID == user.ID
	botAdmin := stringInSlice(m.Author.ID, config.Admins)
	guildOwner := m.Author.ID == guild.OwnerID
	guildAdmin := localPerms&discordgo.PermissionAdministrator > 0
	localManageMessages := localPerms&discordgo.PermissionManageMessages > 0

	return botSelf || botAdmin || guildOwner || guildAdmin || localManageMessages
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

func isDiscordEmoji(link string) bool {
	// always match discord emoji URLs, eg https://cdn.discordapp.com/emojis/340989430460317707.png
	if strings.HasPrefix(link, BASE_URL_DISCORD_EMOJI) {
		return true
	}
	return false
}

func embedColor(channelID string) int {
	// Defined Color
	if isChannelRegistered(channelID) {
		channelConfig := getChannelConfig(channelID)
		if channelConfig.OverwriteEmbedColor != nil {
			if *channelConfig.OverwriteEmbedColor != "" {
				if *channelConfig.OverwriteEmbedColor == "role" || *channelConfig.OverwriteEmbedColor == "user" {
					botColor := bot.State.UserColor(user.ID, channelID)
					if botColor != 0 {
						return botColor
					}
					goto color_random
				}
				if *channelConfig.OverwriteEmbedColor == "random" || *channelConfig.OverwriteEmbedColor == "rand" {
					goto color_random
				}

				var colorString string = *channelConfig.OverwriteEmbedColor

				// Input is Hex
				colorString = strings.ReplaceAll(colorString, "#", "")
				if convertedHex, err := strconv.ParseUint(colorString, 16, 64); err == nil {
					return int(convertedHex)
				}

				// Input is Int
				if convertedInt, err := strconv.Atoi(colorString); err == nil {
					return convertedInt
				}
			}
		}
	}
	// Defined Color
	if config.EmbedColor != nil {
		if *config.EmbedColor != "" {

			if *config.EmbedColor == "role" || *config.EmbedColor == "user" {
				botColor := bot.State.UserColor(user.ID, channelID)
				if botColor != 0 {
					return botColor
				}
				goto color_random
			}
			if *config.EmbedColor == "random" || *config.EmbedColor == "rand" {
				goto color_random
			}

			var colorString string = *config.EmbedColor

			// Input is Hex
			colorString = strings.ReplaceAll(colorString, "#", "")
			if convertedHex, err := strconv.ParseUint(colorString, 16, 64); err == nil {
				return int(convertedHex)
			}

			// Input is Int
			if convertedInt, err := strconv.Atoi(colorString); err == nil {
				return convertedInt
			}
		}
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

func buildEmbed(channelID string, title string, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       embedColor(channelID),
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: PROJECT_ICON,
			Text:    fmt.Sprintf("%s v%s — discordgo v%s", PROJECT_NAME, PROJECT_VERSION, discordgo.VERSION),
		},
	}
}

func sendEmbed(m *discordgo.Message, title string, description string) *discordgo.Message {
	message, _ := bot.ChannelMessageSendComplex(m.ChannelID,
		&discordgo.MessageSend{
			Content: m.Author.Mention(),
			Embed:   buildEmbed(m.ChannelID, title, description),
		},
	)
	return message
}
