package main

import (
	"os"

	"github.com/fatih/color"
)

const (
	projectName    = "discord-downloader-go"
	projectLabel   = "Discord Downloader"
	projectVersion = "1.5.3"
	projectIcon    = "https://discordguide.github.io/assets/Gopher.png"

	projectRepo          = "get-got/discord-downloader-go"
	projectRepoURL       = "https://github.com/" + projectRepo
	projectReleaseURL    = projectRepoURL + "/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/" + projectRepo + "/releases/latest"

	configPath       = "settings.json"
	databasePath     = "database"
	cachePath        = "cache"
	historyCachePath = cachePath + string(os.PathSeparator) + "history"
	imgStorePath     = cachePath + string(os.PathSeparator) + "imgStore"

	defaultReact = "✅"
)

// Log prefixes aren't to be used for constant messages where context is obvious.
var (
	logPrefixSetup   = color.HiGreenString("[Setup]")
	logPrefixDebug   = color.HiYellowString("[Debug]")
	logPrefixHelper  = color.HiMagentaString("[Help]")
	logPrefixInfo    = color.CyanString("[Info]")
	logPrefixHistory = color.HiCyanString("[History]")

	logPrefixFileSkip = color.GreenString(">>> SKIPPING FILE:")
)

const (
	fmtBotSendPerm = "Bot does not have permission to send messages in %s"
)
