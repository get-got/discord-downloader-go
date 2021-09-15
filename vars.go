package main

import (
	"os"

	"github.com/fatih/color"
)

const (
	projectName    = "discord-downloader-go"
	projectLabel   = "Discord Downloader"
	projectVersion = "1.6.3-dev"
	projectIcon    = "https://cdn.discordapp.com/icons/780985109608005703/9dc25f1b91e6d92664590254e0797fad.webp?size=256"

	projectRepo          = "get-got/discord-downloader-go"
	projectRepoURL       = "https://github.com/" + projectRepo
	projectReleaseURL    = projectRepoURL + "/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/" + projectRepo + "/releases/latest"

	configFileBase   = "settings"
	databasePath     = "database"
	cachePath        = "cache"
	historyCachePath = cachePath + string(os.PathSeparator) + "history"
	imgStorePath     = cachePath + string(os.PathSeparator) + "imgStore"
	constantsPath    = cachePath + string(os.PathSeparator) + "constants.json"

	defaultReact = "✅"

	exampleBadImgur  = "https://i.imgur.com/oBsM4iw.jpg"
	exampleBadReddit = "https://i.redd.it/inyxtnitnrs61.jpg"
	exampleBadTumblr = "https://68.media.tumblr.com/946e056f3ccdbe66035a6c813688098c/tumblr_okjnvdxQ5Y1svlqo6o1_400.gif"
	exampleBadGfycat = "https://gfycat.com/SilverDelightfulBlacklemur"
)

var (
	configFile  string
	configFileC bool

	example404s map[string]string
)

// Log prefixes aren't to be used for constant messages where context is obvious.
var (
	logPrefixSetup = color.HiGreenString("[Setup]")

	logPrefixDebug = color.HiYellowString("[Debug]")

	logPrefixHistory = color.HiGreenString("[History]")
	logPrefixInfo    = color.CyanString("[Info]")

	logPrefixDatabase    = color.BlueString("[Database]")
	logPrefixSettings    = color.GreenString("[Settings]")
	logPrefixVersion     = color.HiMagentaString("[Version]")
	logPrefixRegex       = color.HiRedString("[Regex]")
	logPrefixDiscord     = color.HiBlueString("[Discord]")
	logPrefixTwitter     = color.HiCyanString("[Twitter]")
	logPrefixGoogleDrive = color.HiGreenString("[Google Drive]")

	logPrefixFileSkip = color.GreenString(">>> SKIPPING FILE:")
)

func logPrefixDebugLabel(label string) string {
	return color.HiYellowString("[Debug: %s]", label)
}
func logPrefixErrorLabel(label string) string {
	return color.HiRedString("[Error: %s]", label)
}

const (
	fmtBotSendPerm = "Bot does not have permission to send messages in %s"
)
