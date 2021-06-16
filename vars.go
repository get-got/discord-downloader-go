package main

import (
	"os"

	"github.com/fatih/color"
)

const (
	projectName    = "discord-downloader-go"
	projectLabel   = "Discord Downloader"
	projectVersion = "1.6.1-dev"
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
