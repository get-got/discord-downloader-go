package main

import "github.com/fatih/color"

const (
	PROJECT_NAME    = "discord-downloader-go"
	PROJECT_LABEL   = "Discord Download Bot"
	PROJECT_VERSION = "1.2.0"
	PROJECT_ICON    = "https://discordguide.github.io/assets/Gopher.png"

	PROJECT_URL             = "https://github.com/get-got/discord-downloader-go"
	PROJECT_RELEASE_URL     = "https://github.com/get-got/discord-downloader-go/releases/latest"
	PROJECT_RELEASE_API_URL = "https://api.github.com/repos/get-got/discord-downloader-go/releases/latest"

	LOC_CONFIG_FILE  = "settings.json"
	LOC_DATABASE_DIR = "database"

	BASE_URL_DISCORD_EMOJI = "https://cdn.discordapp.com/emojis/"

	IMGUR_CLIENT_ID = "08af502a9e70d65"
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
	CMDERR_LACKING_LOCAL_ADMIN_PERMS = "You do not have permission to use this command.\n" +
		"\nTo use this command you must:" +
		"\n• Be a specified bot administrator (in settings)" +
		"\n• Be Server Owner" +
		"\n• Have Server Administrator Permissions"
	CMDERR_LACKING_BOT_ADMIN_PERMS = "You do not have permission to use this command. You must be a specified bot administrator."
	CMDERR_CHANNEL_NOT_REGISTERED  = "Specified channel is not registered in the bot settings."
	CMDRESP_HISTORY_CANCELLED      = "History cataloging was cancelled."
)
