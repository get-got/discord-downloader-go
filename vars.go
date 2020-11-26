package main

import "github.com/fatih/color"

const (
	projectName    = "discord-downloader-go"
	projectLabel   = "Discord Download Bot"
	projectVersion = "1.3.0"
	projectIcon    = "https://discordguide.github.io/assets/Gopher.png"

	projectRepoURL       = "https://github.com/get-got/discord-downloader-go"
	projectReleaseURL    = "https://github.com/get-got/discord-downloader-go/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/get-got/discord-downloader-go/releases/latest"

	configPath   = "settings.json"
	databasePath = "database"
	imgStorePath = databasePath + "/imgStore"

	imgurClientID = "08af502a9e70d65"
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
	logPrefixDebug       = color.HiYellowString("[DebugOutput]")
	logPrefixDebugExtra  = color.YellowString("[DebugOutput]")
	logPrefixHelper      = color.HiMagentaString("[Helper]")
	logPrefixHelperExtra = color.MagentaString("[Helper]")

	logPrefixFileSkip = color.GreenString(">>> SKIPPING FILE:")
)

// Multiple use messages to save space and make cleaner.
//TODO: Implement this for more?
const (
	cmderrLackingLocalAdminPerms = "You do not have permission to use this command.\n" +
		"\nTo use this command you must:" +
		"\n• Be a specified bot administrator (in settings)" +
		"\n• Be Server Owner" +
		"\n• Have Server Administrator Permissions"
	cmderrLackingBotAdminPerms = "You do not have permission to use this command. You must be a specified bot administrator."
	cmderrChannelNotRegistered = "Specified channel is not registered in the bot settings."
	cmderrHistoryCancelled     = "History cataloging was cancelled."
)
