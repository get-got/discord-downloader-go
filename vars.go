package main

import (
	"os"
)

const (
	projectName    = "discord-downloader-go"
	projectLabel   = "Discord Downloader GO"
	projectVersion = "2.0.1" // follows Semantic Versioning, (http://semver.org/)
	projectIcon    = "https://cdn.discordapp.com/icons/780985109608005703/9dc25f1b91e6d92664590254e0797fad.webp?size=256"

	projectRepo          = "get-got/discord-downloader-go"
	projectRepoURL       = "https://github.com/" + projectRepo
	projectReleaseURL    = projectRepoURL + "/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/" + projectRepo + "/releases/latest"

	databasePath = "database"

	cachePath          = "cache"
	historyCachePath   = cachePath + string(os.PathSeparator) + "history"
	duploCatalogPath   = cachePath + string(os.PathSeparator) + ".duplo"
	instagramCachePath = cachePath + string(os.PathSeparator) + "instagram.json"
	constantsPath      = cachePath + string(os.PathSeparator) + "constants.json"

	defaultReact = "✅"

	limitMsg       = 2000
	limitEmbedDesc = 4096
)

var (
	configFileBase = "settings"
	configFile     string
	configFileC    bool
)

const (
	fmtBotSendPerm = "Bot does not have permission to send messages in %s"
)
