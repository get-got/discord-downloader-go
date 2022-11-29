package main

import (
	"os"
)

const (
	projectName    = "discord-downloader-go"
	projectLabel   = "Discord Downloader GO"
	projectVersion = "2.0.0-rewrite-1"
	projectIcon    = "https://cdn.discordapp.com/icons/780985109608005703/9dc25f1b91e6d92664590254e0797fad.webp?size=256"

	projectRepo          = "get-got/discord-downloader-go"
	projectRepoURL       = "https://github.com/" + projectRepo
	projectReleaseURL    = projectRepoURL + "/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/" + projectRepo + "/releases/latest"

	configFileBase     = "settings"
	databasePath       = "database"
	cachePath          = "cache"
	historyCachePath   = cachePath + string(os.PathSeparator) + "history"
	historyCacheBefore = historyCachePath + string(os.PathSeparator) + "before"
	historyCacheSince  = historyCachePath + string(os.PathSeparator) + "since"
	imgStorePath       = cachePath + string(os.PathSeparator) + "imgStore"
	constantsPath      = cachePath + string(os.PathSeparator) + "constants.json"

	defaultReact = "✅"
)

var (
	configFile  string
	configFileC bool
)

const (
	fmtBotSendPerm = "Bot does not have permission to send messages in %s"
)
