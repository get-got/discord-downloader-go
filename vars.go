package main

import (
	"os"

	"github.com/fatih/color"
)

const (
	projectName    = "discord-downloader-go"
	projectLabel   = "Discord Downloader"
	projectVersion = "1.6.2-dev"
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

	defaultReact = "✅"
)

var (
	configFile  string
	configFileC bool
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

const (
	fmtBotSendPerm = "Bot does not have permission to send messages in %s"
)
