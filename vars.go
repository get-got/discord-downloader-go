package main

import (
	"os"

	"github.com/fatih/color"
)

const (
	projectName    = "discord-downloader-go"
	projectLabel   = "Discord Downloader"
	projectVersion = "1.4.2-a"
	projectIcon    = "https://discordguide.github.io/assets/Gopher.png"

	projectRepoURL       = "https://github.com/get-got/discord-downloader-go"
	projectReleaseURL    = "https://github.com/get-got/discord-downloader-go/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/get-got/discord-downloader-go/releases/latest"

	configPath       = "settings.json"
	databasePath     = "database"
	historyCachePath = databasePath + string(os.PathSeparator) + ".history"
	imgStorePath     = databasePath + string(os.PathSeparator) + "imgStore"

	imgurClientID   = "08af502a9e70d65"
	sneakyUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36"

	defaultReact = "✅"
)

/* Logging Colors:
- HiCyan:		Main Init, Command Use, Handled Event Action
- Cyan:			Main Additional Info, Command Additional Info, Handled Event Additional Info
- HiGreen:		Discord Login Success, Save Success,
- Green:		Discord Login Background, Save Additional Info,
- HiYellow:		Debug, Settings,
- Yellow:		Debug Additional Info, Settings Additional Info,
- HiMagenta:	API Connection Success, Helper
- Magenta:		API Connection Additional Info, Helper Additional Info
- HiRed:		Errors, Exit,
- Red:			Error Additional Info, Exit Additional Info,
*/

// Log prefixes aren't to be used for constant messages where context is obvious.
var (
	logPrefixDebug        = color.HiYellowString("[Debug]")
	logPrefixDebugExtra   = color.YellowString("[Debug]")
	logPrefixHelper       = color.HiMagentaString("[Info]")
	logPrefixHelperExtra  = color.MagentaString("[Info]")
	logPrefixHistory      = color.HiCyanString("[History]")
	logPrefixHistoryExtra = color.CyanString("[History]")

	logPrefixFileSkip = color.GreenString(">>> SKIPPING FILE:")
)

// Multiple use messages to save space and make cleaner.
//TODO: Implement this for more?
const (
	cmderrLackingLocalAdminPerms = "You do not have permission to use this command.\n" +
		"\nTo use this command you must:" +
		"\n• Be set as a bot administrator (in the settings)" +
		"\n• Own this Discord Server" +
		"\n• Have Server Administrator Permissions"
	cmderrLackingBotAdminPerms = "You do not have permission to use this command. Your User ID must be set as a bot administrator in the settings file."
	cmderrChannelNotRegistered = "Specified channel is not registered in the bot settings."
	cmderrHistoryCancelled     = "History cataloging was cancelled."
)

const (
	fmtBotSendPerm = "Bot does not have permission to send messages in %s"
)
