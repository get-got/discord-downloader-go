package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/muhammadmuzzammil1998/jsonc"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

var (
	configFileBase string = "settings"
	configFile     string
	configFileC    bool
	configFileYaml bool
	config         configuration = defaultConfiguration()
)

//#region Config, Credentials

var (
	placeholderToken    string = "REPLACE_WITH_YOUR_TOKEN_OR_DELETE_LINE"
	placeholderEmail    string = "REPLACE_WITH_YOUR_EMAIL_OR_DELETE_LINE"
	placeholderPassword string = "REPLACE_WITH_YOUR_PASSWORD_OR_DELETE_LINE"
)

type configurationCredentials struct {
	// Login
	Token    string `json:"token" yaml:"token"`       // required for bot token (this or login)
	Email    string `json:"email" yaml:"email"`       // required for login (this or token)
	Password string `json:"password" yaml:"password"` // required for login (this or token)
	// APIs
	TwitterAuthEnabled       *bool   `json:"twitterEnabled" yaml:"twitterEnabled"`
	TwitterUsername          string  `json:"twitterUsername" yaml:"twitterUsername"`
	TwitterPassword          string  `json:"twitterPassword" yaml:"twitterPassword"`
	TwitterProxy             string  `json:"twitterProxy,omitempty" yaml:"twitterProxy,omitempty"`
	InstagramAuthEnabled     *bool   `json:"instagramEnabled" yaml:"instagramEnabled"`
	InstagramUsername        string  `json:"instagramUsername" yaml:"instagramUsername"`
	InstagramPassword        string  `json:"instagramPassword" yaml:"instagramPassword"`
	InstagramTOTP            *string `json:"instagramTOTP,omitempty" yaml:"instagramTOTP,omitempty"`
	InstagramProxy           string  `json:"instagramProxy,omitempty" yaml:"instagramProxy,omitempty"`
	InstagramProxyInsecure   *bool   `json:"instagramProxyInsecure,omitempty" yaml:"instagramProxyInsecure,omitempty"`
	InstagramProxyForceHTTP2 *bool   `json:"instagramProxyForceHTTP2,omitempty" yaml:"instagramProxyForceHTTP2,omitempty"`
	FlickrApiKey             string  `json:"flickrApiKey" yaml:"flickrApiKey"`
}

//#endregion

//#region Config, Main

// defConfig_ = Config Default
// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	defConfig_AuthTwitter   bool = true
	defConfig_AuthInstagram bool = true

	defConfig_Debug                bool   = false
	defConfig_CommandPrefix        string = "ddg "
	defConfig_ScanOwnMessages      bool   = false
	defConfig_GithubUpdateChecking bool   = true
	// Appearance
	defConfig_PresenceEnabled      bool               = true
	defConfig_PresenceStatus       string             = string(discordgo.StatusIdle)
	defConfig_PresenceType         discordgo.GameType = discordgo.GameTypeGame
	defConfig_ReactWhenDownloaded  bool               = false
	defConfig_InflateDownloadCount int64              = 0

	// These are only defaults to "fix" when loading settings for when people put stupid values
	defConfig_ProcessLimit int = 32

	defConfig_DiscordTimeout   int = 180
	defConfig_DownloadTimeout  int = 60
	defConfig_DownloadRetryMax int = 2

	defConfig_HistoryManagerRate  int = 5
	defConfig_CheckupRate         int = 30
	defConfig_ConnectionCheckRate int = 5
	defConfig_PresenceRefreshRate int = 3

	defConfig_FilenameDateFormat string = "2006-01-02_15-04-05"
	defConfig_FilenameFormat     string = "{{date}} {{file}}"

	defConfig_HistoryMaxJobs int = 3
)

func defaultConfiguration() configuration {
	return configuration{

		// Logins
		Credentials: configurationCredentials{
			Token:                placeholderToken,
			Email:                placeholderEmail,
			Password:             placeholderPassword,
			TwitterAuthEnabled:   &defConfig_AuthTwitter,
			InstagramAuthEnabled: &defConfig_AuthInstagram,
		},

		// Owner Settings
		Admins:        []string{},
		AdminChannels: []configurationAdminChannel{},

		// Program Settings
		LogIndent:             true,
		ProcessLimit:          defConfig_ProcessLimit,
		Debug:                 defConfig_Debug,
		BackupDatabaseOnStart: false,
		WatchSettings:         true,
		LogSettings:           true,
		MessageOutput:         true,
		MessageOutputHistory:  false,

		DiscordLogLevel:      discordgo.LogError,
		DiscordTimeout:       defConfig_DiscordTimeout,
		DownloadTimeout:      defConfig_DownloadTimeout,
		DownloadRetryMax:     defConfig_DownloadRetryMax,
		ExitOnBadConnection:  false,
		GithubUpdateChecking: defConfig_GithubUpdateChecking,

		CommandPrefix:        defConfig_CommandPrefix,
		CommandTagging:       true,
		ScanOwnMessages:      defConfig_ScanOwnMessages,
		AllowGeneralCommands: true,
		InflateDownloadCount: &defConfig_InflateDownloadCount,
		EuropeanNumbers:      false,

		HistoryManagerRate:  defConfig_HistoryManagerRate,
		CheckupRate:         defConfig_CheckupRate,
		ConnectionCheck:     true,
		ConnectionCheckRate: defConfig_ConnectionCheckRate,
		PresenceRefreshRate: defConfig_PresenceRefreshRate,

		// Emojis & Stickers
		EmojisFilenameFormat:   "{{ID}} {{name}}",
		StickersFilenameFormat: "{{ID}} {{name}}",

		// Source Setup Defaults
		Save:          true,
		AllowCommands: true,
		ScanEdits:     true,
		IgnoreBots:    true,

		SendErrorMessages: false,
		SendFileToChannel: "",
		SendFileDirectly:  true,
		SendFileCaption:   "",

		// Appearance
		PresenceEnabled:            defConfig_PresenceEnabled,
		PresenceStatus:             defConfig_PresenceStatus,
		PresenceType:               defConfig_PresenceType,
		ReactWhenDownloaded:        defConfig_ReactWhenDownloaded,
		ReactWhenDownloadedHistory: false,
		HistoryTyping:              true,

		// History
		HistoryMaxJobs:        defConfig_HistoryMaxJobs,
		AutoHistory:           false,
		AutoHistoryBefore:     "",
		AutoHistorySince:      "",
		SendHistoryStatus:     true,
		SendAutoHistoryStatus: false,
		OutputHistoryStatus:   true,
		OutputHistoryErrors:   true,
		HistoryRequestCount:   100,
		HistoryRequestDelay:   0,

		// Rules for Saving
		Subfolders:             []string{"{{fileType}}"},
		SubfoldersFallback:     nil,
		FilenameDateFormat:     defConfig_FilenameDateFormat,
		FilenameFormat:         defConfig_FilenameFormat,
		FilepathNormalizeText:  false,
		FilepathStripSymbols:   false,
		SaveImages:             true,
		SaveVideos:             true,
		SaveAudioFiles:         true,
		SaveTextFiles:          false,
		SaveOtherFiles:         false,
		SavePossibleDuplicates: false,
		DelayHandling:          0,
		DelayHandlingHistory:   0,
		Filters: &configurationSourceFilters{
			BlockedExtensions: &[]string{
				".htm",
				".html",
				".php",
				".exe",
				".dll",
				".bin",
				".cmd",
				".sh",
				".py",
				".jar",
			},
		},
		Duplo:          false,
		DuploThreshold: 0,
	}
}

type configuration struct {
	Constants map[string]string `json:"_constants" yaml:"_constants"`

	// Logins
	Credentials configurationCredentials `json:"credentials" yaml:"credentials"`

	// Owner Settings
	Admins        []string                    `json:"admins" yaml:"admins"`
	AdminChannels []configurationAdminChannel `json:"adminChannels" yaml:"adminChannels"`

	// Path Overwrites
	OverwriteCachePath           string `json:"overwriteCachePath,omitempty" yaml:"overwriteCachePath,omitempty"`
	OverwriteHistoryPath         string `json:"overwriteHistoryPath,omitempty" yaml:"overwriteHistoryPath,omitempty"`
	OverwriteDuploPath           string `json:"overwriteDuploPath,omitempty" yaml:"overwriteDuploPath,omitempty"`
	OverwriteTwitterPath         string `json:"overwriteTwitterPath,omitempty" yaml:"overwriteTwitterPath,omitempty"`
	OverwriteInstagramPath       string `json:"overwriteInstagramPath,omitempty" yaml:"overwriteInstagramPath,omitempty"`
	OverwriteConstantsPath       string `json:"overwriteConstantsPath,omitempty" yaml:"overwriteConstantsPath,omitempty"`
	OverwriteDatabasePath        string `json:"overwriteDatabasePath,omitempty" yaml:"overwriteDatabasePath,omitempty"`
	OverwriteDatabaseBackupsPath string `json:"overwriteDatabaseBackupsPath,omitempty" yaml:"overwriteDatabaseBackupsPath,omitempty"`

	// Logging
	Verbose              bool   `json:"verbose" yaml:"verbose"`
	Debug                bool   `json:"debug" yaml:"debug"`
	DebugExtra           bool   `json:"debugExtra" yaml:"debugExtra"`
	LogSettings          bool   `json:"settingsOutput" yaml:"settingsOutput"`
	LogOutput            string `json:"logOutput,omitempty" yaml:"logOutput,omitempty"`
	LogIndent            bool   `json:"logIndent" yaml:"logIndent"`
	MessageOutput        bool   `json:"messageOutput" yaml:"messageOutput"`
	MessageOutputHistory bool   `json:"messageOutputHistory" yaml:"messageOutputHistory"`
	DiscordLogLevel      int    `json:"discordLogLevel" yaml:"discordLogLevel"`

	// Program Settings
	ProcessLimit          int    `json:"processLimit" yaml:"processLimit"`
	GithubUpdateChecking  bool   `json:"githubUpdateChecking" yaml:"githubUpdateChecking"`
	ExitOnBadConnection   bool   `json:"exitOnBadConnection" yaml:"exitOnBadConnection"`
	WatchSettings         bool   `json:"watchSettings" yaml:"watchSettings"`
	BackupDatabaseOnStart bool   `json:"backupDatabaseOnStart" yaml:"backupDatabaseOnStart"`
	CheckupRate           int    `json:"checkupRate,omitempty" yaml:"checkupRate,omitempty"`
	ConnectionCheck       bool   `json:"connectionCheck,omitempty" yaml:"connectionCheck,omitempty"`
	ConnectionCheckRate   int    `json:"connectionCheckRate,omitempty" yaml:"connectionCheckRate,omitempty"`
	HistoryManagerRate    int    `json:"historyManagerRate,omitempty" yaml:"historyManagerRate,omitempty"`
	InflateDownloadCount  *int64 `json:"inflateDownloadCount,omitempty" yaml:"inflateDownloadCount,omitempty"`
	EuropeanNumbers       bool   `json:"europeanNumbers,omitempty" yaml:"europeanNumbers,omitempty"`

	// Discord
	ScanEdits            bool   `json:"scanEdits" yaml:"scanEdits"`
	IgnoreBots           bool   `json:"ignoreBots" yaml:"ignoreBots"`
	ScanOwnMessages      bool   `json:"scanOwnMessages" yaml:"scanOwnMessages"`
	AllowCommands        bool   `json:"allowCommands" yaml:"allowCommands"`
	AllowGeneralCommands bool   `json:"allowGeneralCommands" yaml:"allowGeneralCommands"`
	CommandPrefix        string `json:"commandPrefix" yaml:"commandPrefix"`
	CommandTagging       bool   `json:"commandTagging" yaml:"commandTagging"`
	DiscordTimeout       int    `json:"discordTimeout" yaml:"discordTimeout"`
	DownloadTimeout      int    `json:"downloadTimeout" yaml:"downloadTimeout"`
	DownloadRetryMax     int    `json:"downloadRetryMax" yaml:"downloadRetryMax"`
	SendErrorMessages    bool   `json:"sendErrorMessages" yaml:"sendErrorMessages"`

	// Discord Emojis & Stickers
	EmojisServers          *[]string `json:"emojisServers" yaml:"emojisServers"`
	EmojisFilenameFormat   string    `json:"emojisFilenameFormat" yaml:"emojisFilenameFormat"`
	EmojisDestination      *string   `json:"emojisDestination" yaml:"emojisDestination"`
	StickersServers        *[]string `json:"stickersServers" yaml:"stickersServers"`
	StickersFilenameFormat string    `json:"stickersFilenameFormat" yaml:"stickersFilenameFormat"`
	StickersDestination    *string   `json:"stickersDestination" yaml:"stickersDestination"`

	// File Forwarding to Discord Channel
	SendFileToChannel  string   `json:"sendFileToChannel" yaml:"sendFileToChannel"`
	SendFileToChannels []string `json:"sendFileToChannels,omitempty" yaml:"sendFileToChannels,omitempty"`
	SendFileDirectly   bool     `json:"sendFileDirectly,omitempty" yaml:"sendFileDirectly,omitempty"`
	SendFileCaption    string   `json:"sendFileCaption,omitempty" yaml:"sendFileCaption,omitempty"`

	// Discord Presence
	PresenceEnabled     bool               `json:"presenceEnabled" yaml:"presenceEnabled"`
	PresenceStatus      string             `json:"presenceStatus" yaml:"presenceStatus"`
	PresenceType        discordgo.GameType `json:"presenceType" yaml:"presenceType"`
	PresenceLabel       *string            `json:"presenceLabel" yaml:"presenceLabel"`
	PresenceDetails     *string            `json:"presenceDetails" yaml:"presenceDetails"`
	PresenceState       *string            `json:"presenceState" yaml:"presenceState"`
	PresenceRefreshRate int                `json:"presenceRefreshRate,omitempty" yaml:"presenceRefreshRate,omitempty"`

	// Discord Appearance
	ReactWhenDownloaded        bool    `json:"reactWhenDownloaded" yaml:"reactWhenDownloaded"`
	ReactWhenDownloadedEmoji   *string `json:"reactWhenDownloadedEmoji" yaml:"reactWhenDownloadedEmoji"`
	ReactWhenDownloadedHistory bool    `json:"reactWhenDownloadedHistory" yaml:"reactWhenDownloadedHistory"`
	OverwriteDefaultReaction   *string `json:"overwriteDefaultReaction,omitempty" yaml:"overwriteDefaultReaction,omitempty"`
	HistoryTyping              bool    `json:"historyTyping,omitempty" yaml:"historyTyping,omitempty"`
	EmbedColor                 *string `json:"embedColor,omitempty" yaml:"embedColor,omitempty"`

	// History
	HistoryMaxJobs        int    `json:"historyMaxJobs" yaml:"historyMaxJobs"`
	AutoHistory           bool   `json:"autoHistory" yaml:"autoHistory"`
	AutoHistoryBefore     string `json:"autoHistoryBefore" yaml:"autoHistoryBefore"`
	AutoHistorySince      string `json:"autoHistorySince" yaml:"autoHistorySince"`
	SendAutoHistoryStatus bool   `json:"sendAutoHistoryStatus" yaml:"sendAutoHistoryStatus"`
	SendHistoryStatus     bool   `json:"sendHistoryStatus" yaml:"sendHistoryStatus"`
	OutputHistoryStatus   bool   `json:"outputHistoryStatus" yaml:"outputHistoryStatus"`
	OutputHistoryErrors   bool   `json:"outputHistoryErrors" yaml:"outputHistoryErrors"`
	HistoryRequestCount   int    `json:"historyRequestCount" yaml:"historyRequestCount"`
	HistoryRequestDelay   int    `json:"historyRequestDelay" yaml:"historyRequestDelay"`

	// Rules for Saving
	Save                   bool                        `json:"save" yaml:"save"`
	Subfolders             []string                    `json:"subfolders" yaml:"subfolders"`
	SubfoldersFallback     []string                    `json:"subfoldersFallback,omitempty" yaml:"subfoldersFallback,omitempty"`
	FilenameDateFormat     string                      `json:"filenameDateFormat" yaml:"filenameDateFormat"`
	FilenameFormat         string                      `json:"filenameFormat" yaml:"filenameFormat"`
	FilepathNormalizeText  bool                        `json:"filepathNormalizeText,omitempty" yaml:"filepathNormalizeText,omitempty"`
	FilepathStripSymbols   bool                        `json:"filepathStripSymbols,omitempty" yaml:"filepathStripSymbols,omitempty"`
	SaveImages             bool                        `json:"saveImages" yaml:"saveImages"`
	SaveVideos             bool                        `json:"saveVideos" yaml:"saveVideos"`
	SaveAudioFiles         bool                        `json:"saveAudioFiles" yaml:"saveAudioFiles"`
	SaveTextFiles          bool                        `json:"saveTextFiles" yaml:"saveTextFiles"`
	SaveOtherFiles         bool                        `json:"saveOtherFiles" yaml:"saveOtherFiles"`
	SavePossibleDuplicates bool                        `json:"savePossibleDuplicates" yaml:"savePossibleDuplicates"`
	DelayHandling          int                         `json:"delayHandling,omitempty" yaml:"delayHandling,omitempty"`
	DelayHandlingHistory   int                         `json:"delayHandlingHistory,omitempty" yaml:"delayHandlingHistory,omitempty"`
	Filters                *configurationSourceFilters `json:"filters" yaml:"filters"`
	Duplo                  bool                        `json:"duplo,omitempty" yaml:"duplo,omitempty"`
	DuploThreshold         float64                     `json:"duploThreshold,omitempty" yaml:"duploThreshold,omitempty"`

	// Misc Rules
	LogLinks    *configurationSourceLog `json:"logLinks,omitempty" yaml:"logLinks,omitempty"`
	LogMessages *configurationSourceLog `json:"logMessages,omitempty" yaml:"logMessages,omitempty"`

	// Sources
	All                    *configurationSource  `json:"all,omitempty" yaml:"all,omitempty"`
	AllBlacklistUsers      *[]string             `json:"allBlacklistUsers,omitempty" yaml:"allBlacklistUsers,omitempty"`
	AllBlacklistServers    *[]string             `json:"allBlacklistServers,omitempty" yaml:"allBlacklistServers,omitempty"`
	AllBlacklistCategories *[]string             `json:"allBlacklistCategories,omitempty" yaml:"allBlacklistCategories,omitempty"`
	AllBlacklistChannels   *[]string             `json:"allBlacklistChannels,omitempty" yaml:"allBlacklistChannels,omitempty"`
	Users                  []configurationSource `json:"users,omitempty" yaml:"users,omitempty"`
	Servers                []configurationSource `json:"servers,omitempty" yaml:"servers,omitempty"`
	Categories             []configurationSource `json:"categories,omitempty" yaml:"categories,omitempty"`
	Channels               []configurationSource `json:"channels,omitempty" yaml:"channels,omitempty"`
}

//#endregion

//#region Config, Sources

// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	defSource_Enabled bool = true
)

type configurationSource struct {
	// ~
	UserID            string    `json:"user,omitempty" yaml:"user,omitempty"`
	UserIDs           *[]string `json:"users,omitempty" yaml:"users,omitempty"`
	ServerID          string    `json:"server,omitempty" yaml:"server,omitempty"`
	ServerIDs         *[]string `json:"servers,omitempty" yaml:"servers,omitempty"`
	ServerBlacklist   *[]string `json:"serverBlacklist,omitempty" yaml:"serverBlacklist,omitempty"`
	CategoryID        string    `json:"category,omitempty" yaml:"category,omitempty"`
	CategoryIDs       *[]string `json:"categories,omitempty" yaml:"categories,omitempty"`
	CategoryBlacklist *[]string `json:"categoryBlacklist,omitempty" yaml:"categoryBlacklist,omitempty"`
	ChannelID         string    `json:"channel,omitempty" yaml:"channel,omitempty"`
	ChannelIDs        *[]string `json:"channels,omitempty" yaml:"channels,omitempty"`
	Destination       string    `json:"destination" yaml:"destination"`
	Alias             *string   `json:"alias,omitempty" yaml:"alias,omitempty"`
	Aliases           *[]string `json:"aliases,omitempty" yaml:"aliases,omitempty"`

	// Setup
	Enabled       *bool   `json:"enabled" yaml:"enabled"`
	Save          *bool   `json:"save" yaml:"save"`
	AllowCommands *bool   `json:"allowCommands" yaml:"allowCommands"`
	ScanEdits     *bool   `json:"scanEdits" yaml:"scanEdits"`
	IgnoreBots    *bool   `json:"ignoreBots" yaml:"ignoreBots"`
	CommandPrefix *string `json:"commandPrefix" yaml:"commandPrefix"`

	SendErrorMessages  *bool     `json:"sendErrorMessages" yaml:"sendErrorMessages"`
	SendFileToChannel  *string   `json:"sendFileToChannel" yaml:"sendFileToChannel"`
	SendFileToChannels *[]string `json:"sendFileToChannels,omitempty" yaml:"sendFileToChannels,omitempty"`
	SendFileDirectly   *bool     `json:"sendFileDirectly,omitempty" yaml:"sendFileDirectly,omitempty"`
	SendFileCaption    *string   `json:"sendFileCaption,omitempty" yaml:"sendFileCaption,omitempty"`

	// Appearance
	PresenceEnabled            *bool     `json:"presenceEnabled" yaml:"presenceEnabled"`
	ReactWhenDownloaded        *bool     `json:"reactWhenDownloaded" yaml:"reactWhenDownloaded"`
	ReactWhenDownloadedEmoji   *string   `json:"reactWhenDownloadedEmoji" yaml:"reactWhenDownloadedEmoji"`
	ReactWhenDownloadedHistory *bool     `json:"reactWhenDownloadedHistory" yaml:"reactWhenDownloadedHistory"`
	BlacklistReactEmojis       *[]string `json:"blacklistReactEmojis,omitempty" yaml:"blacklistReactEmojis,omitempty"`
	HistoryTyping              *bool     `json:"historyTyping,omitempty" yaml:"historyTyping,omitempty"`
	EmbedColor                 *string   `json:"embedColor,omitempty" yaml:"embedColor,omitempty"`

	// History
	AutoHistory           *bool   `json:"autoHistory" yaml:"autoHistory"`
	AutoHistoryBefore     *string `json:"autoHistoryBefore" yaml:"autoHistoryBefore"`
	AutoHistorySince      *string `json:"autoHistorySince" yaml:"autoHistorySince"`
	SendAutoHistoryStatus *bool   `json:"sendAutoHistoryStatus" yaml:"sendAutoHistoryStatus"`
	SendHistoryStatus     *bool   `json:"sendHistoryStatus" yaml:"sendHistoryStatus"`
	OutputHistoryStatus   *bool   `json:"outputHistoryStatus" yaml:"outputHistoryStatus"`
	OutputHistoryErrors   *bool   `json:"outputHistoryErrors" yaml:"outputHistoryErrors"`

	// Rules for Saving
	Subfolders             *[]string                   `json:"subfolders,omitempty" yaml:"subfolders,omitempty"`
	SubfoldersFallback     *[]string                   `json:"subfoldersFallback,omitempty" yaml:"subfoldersFallback,omitempty"`
	FilenameDateFormat     *string                     `json:"filenameDateFormat" yaml:"filenameDateFormat"`
	FilenameFormat         *string                     `json:"filenameFormat" yaml:"filenameFormat"`
	FilepathNormalizeText  *bool                       `json:"filepathNormalizeText,omitempty" yaml:"filepathNormalizeText,omitempty"`
	FilepathStripSymbols   *bool                       `json:"filepathStripSymbols,omitempty" yaml:"filepathStripSymbols,omitempty"`
	SaveImages             *bool                       `json:"saveImages" yaml:"saveImages"`
	SaveVideos             *bool                       `json:"saveVideos" yaml:"saveVideos"`
	SaveAudioFiles         *bool                       `json:"saveAudioFiles" yaml:"saveAudioFiles"`
	SaveTextFiles          *bool                       `json:"saveTextFiles" yaml:"saveTextFiles"`
	SaveOtherFiles         *bool                       `json:"saveOtherFiles" yaml:"saveOtherFiles"`
	SavePossibleDuplicates *bool                       `json:"savePossibleDuplicates" yaml:"savePossibleDuplicates"`
	DelayHandling          *int                        `json:"delayHandling,omitempty" yaml:"delayHandling,omitempty"`
	DelayHandlingHistory   *int                        `json:"delayHandlingHistory,omitempty" yaml:"delayHandlingHistory,omitempty"`
	Filters                *configurationSourceFilters `json:"filters" yaml:"filters"`
	Duplo                  *bool                       `json:"duplo,omitempty" yaml:"duplo,omitempty"`
	DuploThreshold         *float64                    `json:"duploThreshold,omitempty" yaml:"duploThreshold,omitempty"`

	// Misc Rules
	LogLinks    *configurationSourceLog `json:"logLinks,omitempty" yaml:"logLinks,omitempty"`
	LogMessages *configurationSourceLog `json:"logMessages,omitempty" yaml:"logMessages,omitempty"`
}

type configurationSourceFilters struct {
	BlockedPhrases *[]string `json:"blockedPhrases,omitempty" yaml:"blockedPhrases,omitempty"`
	AllowedPhrases *[]string `json:"allowedPhrases,omitempty" yaml:"allowedPhrases,omitempty"`

	BlockedUsers *[]string `json:"blockedUsers,omitempty" yaml:"blockedUsers,omitempty"`
	AllowedUsers *[]string `json:"allowedUsers,omitempty" yaml:"allowedUsers,omitempty"`

	BlockedRoles *[]string `json:"blockedRoles,omitempty" yaml:"blockedRoles,omitempty"`
	AllowedRoles *[]string `json:"allowedRoles,omitempty" yaml:"allowedRoles,omitempty"`

	BlockedLinkContent *[]string `json:"blockedLinkContent,omitempty" yaml:"blockedLinkContent,omitempty"`
	AllowedLinkContent *[]string `json:"allowedLinkContent,omitempty" yaml:"allowedLinkContent,omitempty"`

	BlockedDomains *[]string `json:"blockedDomains,omitempty" yaml:"blockedDomains,omitempty"`
	AllowedDomains *[]string `json:"allowedDomains,omitempty" yaml:"allowedDomains,omitempty"`

	BlockedExtensions *[]string `json:"blockedExtensions,omitempty" yaml:"blockedExtensions,omitempty"`
	AllowedExtensions *[]string `json:"allowedExtensions,omitempty" yaml:"allowedExtensions,omitempty"`

	BlockedFilenames *[]string `json:"blockedFilenames,omitempty" yaml:"blockedFilenames,omitempty"`
	AllowedFilenames *[]string `json:"allowedFilenames,omitempty" yaml:"allowedFilenames,omitempty"`

	BlockedReactions *[]string `json:"blockedReactions,omitempty" yaml:"blockedReactions,omitempty"`
	AllowedReactions *[]string `json:"allowedReactions,omitempty" yaml:"allowedReactions,omitempty"`
}

var (
	defSourceLog_Subfolders            []string = []string{"{{year}}-{{monthNum}}-{{dayOfMonth}}"}
	defSourceLog_SubfoldersFallback    []string = nil
	defSourceLog_FilenameFormat        string   = "{{serverName}} - {{channelName}}.txt"
	defSourceLog_FilepathNormalizeText bool     = false
	defSourceLog_FilepathStripSymbols  bool     = false
	defSourceLog_LinePrefix            string   = "[{{serverName}} / {{channelName}}] \"{{username}}\" @ {{timestamp}}: "
	defSourceLog_LogDownloads          bool     = true
	defSourceLog_LogFailures           bool     = true
	defSourceLogMsg_LineContent        string   = "{{message}}"
	defSourceLogLink_LineContent       string   = "{{link}}"
)

type configurationSourceLog struct {
	Destination           string    `json:"destination" yaml:"destination"`
	Subfolders            *[]string `json:"subfolders,omitempty" yaml:"subfolders,omitempty"`
	SubfoldersFallback    *[]string `json:"subfoldersFallback,omitempty" yaml:"subfoldersFallback,omitempty"`
	FilenameFormat        *string   `json:"filenameFormat,omitempty" yaml:"filenameFormat,omitempty"`
	FilepathNormalizeText *bool     `json:"filepathNormalizeText,omitempty" yaml:"filepathNormalizeText,omitempty"`
	FilepathStripSymbols  *bool     `json:"filepathStripSymbols,omitempty" yaml:"filepathStripSymbols,omitempty"`

	LinePrefix  *string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	LineSuffix  *string `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	LineContent *string `json:"content,omitempty" yaml:"content,omitempty"`

	FilterDuplicates *bool `json:"filterDuplicates" yaml:"filterDuplicates"`

	LogDownloads *bool `json:"logDownloads" yaml:"logDownloads"` // links only
	LogFailures  *bool `json:"logFailures" yaml:"logFailures"`   // links only
}

//#endregion

//#region Config, Admin Channels

var (
	adefConfig_LogProgram     bool = false
	adefConfig_LogStatus      bool = true
	adefConfig_LogErrors      bool = true
	adefConfig_UnlockCommands bool = false
)

type configurationAdminChannel struct {
	// Specify target command channels
	ChannelID      string    `json:"channel" yaml:"channel"`
	ChannelIDs     *[]string `json:"channels,omitempty" yaml:"channels,omitempty"`
	LogProgram     *bool     `json:"logProgram" yaml:"logProgram"`
	LogStatus      *bool     `json:"logStatus" yaml:"logStatus"`
	LogErrors      *bool     `json:"logErrors" yaml:"logErrors"`
	UnlockCommands *bool     `json:"unlockCommands" yaml:"unlockCommands"`
	CommandPrefix  *string   `json:"commandPrefix,omitempty" yaml:"commandPrefix,omitempty"`
}

//#endregion

//#region Management

func loadConfig() error {
	// Determine config type
	if _, err := os.Stat(configFileBase + ".yaml"); err == nil {
		configFile = configFileBase + ".yaml"
		configFileYaml = true
	} else if _, err := os.Stat(configFileBase + ".jsonc"); err == nil {
		configFile = configFileBase + ".jsonc"
		configFileC = true
	} else {
		configFile = configFileBase + ".json"
		configFileC = false
	}

	log.Println(lg("Settings", "loadConfig", color.YellowString, "Loading from \"%s\"...", configFile))

	// Load settings
	configContent, err := os.ReadFile(configFile)
	if err != nil {
		log.Println(lg("Settings", "loadConfig", color.HiRedString, "Failed to open file...\t%s", err))
		createConfig()
		properExit()
	} else {
		fixed := string(configContent)
		// Fix backslashes
		fixed = strings.ReplaceAll(fixed, "\\", "\\\\")
		for strings.Contains(fixed, "\\\\\\") {
			fixed = strings.ReplaceAll(fixed, "\\\\\\", "\\\\")
		}
		//TODO: Not even sure if this is realistic to do but would be nice to have line comma & trailing comma fixing

		// Parse
		newConfig := defaultConfiguration()
		if configFileYaml {
			err = yaml.Unmarshal([]byte(fixed), &newConfig)
		} else if configFileC {
			err = jsonc.Unmarshal([]byte(fixed), &newConfig)
		} else {
			err = json.Unmarshal([]byte(fixed), &newConfig)
		}
		if err != nil {
			log.Println(lg("Settings", "loadConfig", color.HiRedString, "Failed to parse settings file...\t%s", err))
			log.Println(lg("Settings", "loadConfig", color.MagentaString, "Please ensure you're following proper JSON format syntax."))
			properExit()
		}
		// Constants
		if newConfig.Constants != nil {
			for key, value := range newConfig.Constants {
				if strings.Contains(fixed, key) {
					fixed = strings.ReplaceAll(fixed, key, value)
				}
			}
			// Re-parse
			newConfig = defaultConfiguration()
			if configFileYaml {
				err = yaml.Unmarshal([]byte(fixed), &newConfig)
			} else if configFileC {
				err = jsonc.Unmarshal([]byte(fixed), &newConfig)
			} else {
				err = json.Unmarshal([]byte(fixed), &newConfig)
			}
			if err != nil {
				log.Println(lg("Settings", "loadConfig", color.HiRedString,
					"Failed to re-parse settings file after replacing constants...\t%s", err))
				log.Println(lg("Settings", "loadConfig", color.MagentaString, "Please ensure you're following proper JSON format syntax."))
				properExit()
			}
			newConfig.Constants = nil
		}
		config = newConfig

		// Source Defaults
		for i := 0; i < len(config.Channels); i++ {
			sourceDefault(&config.Channels[i])
		}
		for i := 0; i < len(config.Categories); i++ {
			sourceDefault(&config.Categories[i])
		}
		for i := 0; i < len(config.Servers); i++ {
			sourceDefault(&config.Servers[i])
		}
		for i := 0; i < len(config.Users); i++ {
			sourceDefault(&config.Users[i])
		}
		if config.All != nil {
			sourceDefault(config.All)
		}
		// Admin Channel Defaults
		for i := 0; i < len(config.AdminChannels); i++ {
			adminChannelDefault(&config.AdminChannels[i])
		}

		// Checks & Fixes
		if config.ProcessLimit < 1 {
			config.ProcessLimit = defConfig_ProcessLimit
		}
		if config.DiscordTimeout < 10 {
			config.DiscordTimeout = defConfig_DiscordTimeout
		}
		if config.DownloadTimeout < 10 {
			config.DownloadTimeout = defConfig_DownloadTimeout
		}
		if config.DownloadRetryMax < 1 {
			config.DownloadRetryMax = defConfig_DownloadRetryMax
		}
		if config.CheckupRate < 1 {
			config.CheckupRate = defConfig_CheckupRate
		}
		if config.ConnectionCheckRate < 1 {
			config.ConnectionCheckRate = defConfig_ConnectionCheckRate
		}
		if config.PresenceRefreshRate < 1 {
			config.PresenceRefreshRate = defConfig_PresenceRefreshRate
		}
		if config.FilenameDateFormat == "" {
			config.FilenameDateFormat = defConfig_FilenameDateFormat
		}
		if config.FilenameFormat == "" {
			config.FilenameFormat = defConfig_FilenameFormat
		}
		if config.HistoryMaxJobs < 1 {
			config.HistoryMaxJobs = defConfig_HistoryMaxJobs
		}

		// Log to File
		if config.LogOutput != "" {
			f, err := os.OpenFile(config.LogOutput, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Println(lg("Settings", "loadConfig", color.HiRedString, "Failed to open LogOutput file...\t%s", err))
			} else {
				log.SetOutput(io.MultiWriter(color.Output, f))
			}
		}

		// Misc Rules
		if config.LogLinks != nil {
			if config.LogLinks.Subfolders == nil {
				config.LogLinks.Subfolders = &defSourceLog_Subfolders
			}
			if config.LogLinks.SubfoldersFallback == nil {
				config.LogLinks.SubfoldersFallback = &defSourceLog_SubfoldersFallback
			}
			if config.LogLinks.FilenameFormat == nil {
				config.LogLinks.FilenameFormat = &defSourceLog_FilenameFormat
			}
			if config.LogLinks.FilepathNormalizeText == nil {
				config.LogLinks.FilepathNormalizeText = &defSourceLog_FilepathNormalizeText
			}
			if config.LogLinks.FilepathStripSymbols == nil {
				config.LogLinks.FilepathStripSymbols = &defSourceLog_FilepathStripSymbols
			}
			if config.LogLinks.LinePrefix == nil {
				config.LogLinks.LinePrefix = &defSourceLog_LinePrefix
			}
			// vv unique vv
			if config.LogLinks.LineContent == nil {
				config.LogLinks.LineContent = &defSourceLogLink_LineContent
			}
			if config.LogLinks.LogDownloads == nil {
				config.LogLinks.LogDownloads = &defSourceLog_LogDownloads
			}
			if config.LogLinks.LogFailures == nil {
				config.LogLinks.LogFailures = &defSourceLog_LogFailures
			}
		}
		if config.LogMessages != nil {
			if config.LogMessages.Subfolders == nil {
				config.LogMessages.Subfolders = &defSourceLog_Subfolders
			}
			if config.LogMessages.SubfoldersFallback == nil {
				config.LogMessages.SubfoldersFallback = &defSourceLog_SubfoldersFallback
			}
			if config.LogMessages.FilenameFormat == nil {
				config.LogMessages.FilenameFormat = &defSourceLog_FilenameFormat
			}
			if config.LogMessages.FilepathNormalizeText == nil {
				config.LogMessages.FilepathNormalizeText = &defSourceLog_FilepathNormalizeText
			}
			if config.LogMessages.FilepathStripSymbols == nil {
				config.LogMessages.FilepathStripSymbols = &defSourceLog_FilepathStripSymbols
			}
			if config.LogMessages.LinePrefix == nil {
				config.LogMessages.LinePrefix = &defSourceLog_LinePrefix
			}
			// vv unique vv
			if config.LogMessages.LineContent == nil {
				config.LogMessages.LineContent = &defSourceLogMsg_LineContent
			}
		}

		// Overwrite Paths
		if config.OverwriteCachePath != "" {
			pathCache = config.OverwriteCachePath
		}
		if config.OverwriteHistoryPath != "" {
			pathCacheHistory = config.OverwriteHistoryPath
		}
		if config.OverwriteDuploPath != "" {
			pathCacheDuplo = config.OverwriteDuploPath
		}
		if config.OverwriteTwitterPath != "" {
			pathCacheTwitter = config.OverwriteTwitterPath
		}
		if config.OverwriteInstagramPath != "" {
			pathCacheInstagram = config.OverwriteInstagramPath
		}
		if config.OverwriteConstantsPath != "" {
			pathConstants = config.OverwriteConstantsPath
		}
		if config.OverwriteDatabasePath != "" {
			pathDatabaseBase = config.OverwriteDatabasePath
		}
		if config.OverwriteDatabaseBackupsPath != "" {
			pathDatabaseBackups = config.OverwriteDatabaseBackupsPath
		}

		// Overwrite Default Reaction
		if config.OverwriteDefaultReaction != nil {
			defaultReact = *config.OverwriteDefaultReaction
		}

		// Settings Output
		if config.LogSettings {
			dupeConfig := config
			if dupeConfig.Credentials.Token != "" && dupeConfig.Credentials.Token != placeholderToken {
				dupeConfig.Credentials.Token = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.Email != "" && dupeConfig.Credentials.Email != placeholderEmail {
				dupeConfig.Credentials.Email = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.Password != "" && dupeConfig.Credentials.Password != placeholderPassword {
				dupeConfig.Credentials.Password = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.TwitterUsername != "" {
				dupeConfig.Credentials.TwitterUsername = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.TwitterPassword != "" {
				dupeConfig.Credentials.TwitterPassword = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.InstagramUsername != "" {
				dupeConfig.Credentials.InstagramUsername = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.InstagramPassword != "" {
				dupeConfig.Credentials.InstagramPassword = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.FlickrApiKey != "" {
				dupeConfig.Credentials.FlickrApiKey = "STRIPPED_FOR_OUTPUT"
			}
			s, err := json.MarshalIndent(dupeConfig, "", "\t")
			if err != nil {
				log.Println(lg("Debug", "loadConfig", color.HiRedString, "Failed to output...\t%s", err))
			} else {
				log.Println(lg("Debug", "loadConfig", color.HiYellowString, "Parsed into JSON:\n\n%s", color.YellowString(string(s))))
			}
		}

		// Credentials Check
		if (config.Credentials.Token == "" || config.Credentials.Token == placeholderToken) &&
			(config.Credentials.Email == "" || config.Credentials.Email == placeholderEmail) &&
			(config.Credentials.Password == "" || config.Credentials.Password == placeholderPassword) {
			log.Println(lg("Settings", "loadConfig", color.HiRedString, "No valid discord login found..."))
			log.Println(lg("Settings", "loadConfig", color.HiYellowString, "Please save your credentials & info into \"%s\" then restart...", configFile))
			log.Println(lg("Settings", "loadConfig", color.MagentaString, "If your credentials are already properly saved, please ensure you're following proper JSON format syntax..."))
			log.Println(lg("Settings", "loadConfig", color.MagentaString, "You DO NOT NEED token *AND* email/password, just one OR the other."))
			properExit()
		}

		allString := ""
		if config.All != nil {
			allString = ", ALL GROUP ENABLED"
		}
		log.Println(lg("Settings", "", color.HiYellowString,
			"Finished loading ... bound to %d channel%s, %d categories, %d server%s, %d user%s%s",
			getBoundChannelsCount(), pluralS(getBoundChannelsCount()),
			getBoundCategoriesCount(),
			getBoundServersCount(), pluralS(getBoundServersCount()),
			getBoundUsersCount(), pluralS(getBoundUsersCount()), allString,
		))

		// SETTINGS TO BE APPLIED IMMEDIATELY

		if config.ProcessLimit > 0 {
			runtime.GOMAXPROCS(config.ProcessLimit)
		}

		// Make cache path
		os.MkdirAll(pathCache, 0755)

		// Convert and cache
		if !configFileYaml {
			d, err := yaml.Marshal(&config)
			if err == nil {
				os.WriteFile(pathCacheSettingsYAML, d, 0644)
			}
		} else {
			d, err := json.Marshal(&config)
			if err == nil {
				os.WriteFile(pathCacheSettingsJSON, d, 0644)
			}
		}
	}

	return nil
}

func createConfig() {
	log.Println(lg("Settings", "create", color.YellowString, "Creating new settings file..."))

	defaultConfig := defaultConfiguration()
	defaultConfig.Credentials.Token = placeholderToken
	defaultConfig.Credentials.Email = placeholderEmail
	defaultConfig.Credentials.Password = placeholderPassword
	defaultConfig.Admins = []string{"REPLACE_WITH_YOUR_DISCORD_USER_ID"}

	enteredBaseChannel := "REPLACE_WITH_DISCORD_CHANNEL_ID_TO_DOWNLOAD_FROM"
	enteredBaseDestination := "REPLACE_WITH_FOLDER_LOCATION_TO_DOWNLOAD_TO"

	// Import old config
	if _, err := os.Stat("config.ini"); err == nil {
		log.Println(lg("Settings", "create", color.HiGreenString,
			"Detected config.ini from Seklfreak's discord-image-downloader-go, importing..."))
		cfg, err := ini.Load("config.ini")
		if err != nil {
			log.Println(lg("Settings", "create", color.HiRedString,
				"Unable to read your old config file:\t%s", err))
			cfg = ini.Empty()
		} else { // Import old ini
			importKey := func(section string, key string, outVar interface{}, outType string) bool {
				if cfg.Section(section).HasKey(key) {
					if outType == "string" {
						outVar = cfg.Section(section).Key(key).String()
					} else if outType == "int" {
						outVar = cfg.Section(section).Key(key).MustInt()
					} else if outType == "bool" {
						outVar = cfg.Section(section).Key(key).MustBool()
					}
					log.Println(lg("Settings", "create", color.GreenString, "IMPORTED %s - %s:\t\t\t%s", section, key, outVar))
					return true
				}
				return false
			}

			// Auth
			if !importKey("auth", "token", &defaultConfig.Credentials.Token, "string") {
				defaultConfig.Credentials.Token = ""
			}
			if !importKey("auth", "email", &defaultConfig.Credentials.Email, "string") {
				defaultConfig.Credentials.Email = ""
			}
			if !importKey("auth", "password", &defaultConfig.Credentials.Password, "string") {
				defaultConfig.Credentials.Password = ""
			}
			importKey("flickr", "api key", &defaultConfig.Credentials.FlickrApiKey, "string")
			// General
			importKey("general", "max download retries", &defaultConfig.DownloadRetryMax, "int")
			importKey("general", "download timeout", &defaultConfig.DownloadTimeout, "int")

			// Status
			importKey("status", "status enabled", &defaultConfig.PresenceEnabled, "bool")
			importKey("status", "status type", &defaultConfig.PresenceStatus, "string")
			importKey("status", "status label", &defaultConfig.PresenceType, "int")

			// Channels
			InteractiveChannelWhitelist := cfg.Section("interactive channels").KeysHash()
			for key := range InteractiveChannelWhitelist {
				newChannel := configurationAdminChannel{
					ChannelID: key,
				}
				log.Println(lg("Settings", "create", color.GreenString, "IMPORTED Admin Channel:\t\t%s", key))
				defaultConfig.AdminChannels = append(defaultConfig.AdminChannels, newChannel)
			}
			ChannelWhitelist := cfg.Section("channels").KeysHash()
			for key, value := range ChannelWhitelist {
				newChannel := configurationSource{
					ChannelID:   key,
					Destination: value,
				}
				log.Println(lg("Settings", "create", color.GreenString, "IMPORTED Channel:\t\t\t%s to \"%s\"", key, value))
				defaultConfig.Channels = append(defaultConfig.Channels, newChannel)
			}
		}
		log.Println(lg("Settings", "create", color.HiGreenString,
			"Finished importing config.ini from Seklfreak's discord-image-downloader-go!"))
	} else {
		var baseChannel configurationSource
		sourceDefault(&baseChannel)
		baseChannel.ChannelID = enteredBaseChannel
		baseChannel.Destination = enteredBaseDestination

		defaultConfig.Channels = append(defaultConfig.Channels, baseChannel)

		baseAdminChannel := configurationAdminChannel{
			ChannelID: "REPLACE_WITH_DISCORD_CHANNEL_ID_FOR_ADMIN_COMMANDS",
		}
		defaultConfig.AdminChannels = append(defaultConfig.AdminChannels, baseAdminChannel)
		adminChannelDefault(&defaultConfig.AdminChannels[0])

		//TODO: Improve, this is very crude, I just wanted *something* for this.
		log.Print(lg("Settings", "create", color.HiCyanString, "Would you like to enter settings info now? [Y/N]: "))
		reader := bufio.NewReader(os.Stdin)
		inputCredsYN, _ := reader.ReadString('\n')
		inputCredsYN = strings.ReplaceAll(inputCredsYN, "\n", "")
		inputCredsYN = strings.ReplaceAll(inputCredsYN, "\r", "")
		if strings.Contains(strings.ToLower(inputCredsYN), "y") {
		EnterCreds:
			log.Print(color.HiCyanString("Token or Login? [\"token\"/\"login\"]: "))
			inputCreds, _ := reader.ReadString('\n')
			inputCreds = strings.ReplaceAll(inputCreds, "\n", "")
			inputCreds = strings.ReplaceAll(inputCreds, "\r", "")
			if strings.Contains(strings.ToLower(inputCreds), "token") {
			EnterToken:
				log.Print(color.HiCyanString("Enter token: "))
				inputToken, _ := reader.ReadString('\n')
				inputToken = strings.ReplaceAll(inputToken, "\n", "")
				inputToken = strings.ReplaceAll(inputToken, "\r", "")
				if inputToken != "" {
					defaultConfig.Credentials.Token = inputToken
				} else {
					log.Println(lg("Settings", "create", color.HiRedString, "Please input token..."))
					goto EnterToken
				}
			} else if strings.Contains(strings.ToLower(inputCreds), "login") {
			EnterEmail:
				log.Print(color.HiCyanString("Enter email: "))
				inputEmail, _ := reader.ReadString('\n')
				inputEmail = strings.ReplaceAll(inputEmail, "\n", "")
				inputEmail = strings.ReplaceAll(inputEmail, "\r", "")
				if strings.Contains(inputEmail, "@") {
					defaultConfig.Credentials.Email = inputEmail
				EnterPassword:
					log.Print(color.HiCyanString("Enter password: "))
					inputPassword, _ := reader.ReadString('\n')
					inputPassword = strings.ReplaceAll(inputPassword, "\n", "")
					inputPassword = strings.ReplaceAll(inputPassword, "\r", "")
					if inputPassword != "" {
						defaultConfig.Credentials.Password = inputPassword
					} else {
						log.Println(lg("Settings", "create", color.HiRedString, "Please input password..."))
						goto EnterPassword
					}
				} else {
					log.Println(lg("Settings", "create", color.HiRedString, "Please input email..."))
					goto EnterEmail
				}
			} else {
				log.Println(lg("Settings", "create", color.HiRedString, "Please input \"token\" or \"login\"..."))
				goto EnterCreds
			}

		EnterAdmin:
			log.Print(color.HiCyanString("Input your Discord User ID: "))
			inputAdmin, _ := reader.ReadString('\n')
			inputAdmin = strings.ReplaceAll(inputAdmin, "\n", "")
			inputAdmin = strings.ReplaceAll(inputAdmin, "\r", "")
			if isNumeric(inputAdmin) {
				defaultConfig.Admins = []string{inputAdmin}
			} else {
				log.Println(lg("Settings", "create", color.HiRedString, "Please input your Discord User ID..."))
				goto EnterAdmin
			}

			//TODO: Base channel setup? Would be kind of annoying and may limit options
			//TODO: Admin channel setup?
		}
	}

	log.Println(lg("Settings", "create", color.MagentaString,
		"The default settings will be missing some options to avoid clutter."))
	log.Println(lg("Settings", "create", color.HiMagentaString,
		"There are MANY MORE SETTINGS! If you would like to maximize customization, see the GitHub README for all available settings."))

	defaultJSON, err := json.MarshalIndent(defaultConfig, "", "\t")
	if err != nil {
		log.Println(lg("Settings", "create", color.HiRedString, "Failed to format new settings...\t%s", err))
	} else {
		err := os.WriteFile(configFile, defaultJSON, 0644)
		if err != nil {
			log.Println(lg("Settings", "create", color.HiRedString, "Failed to save new settings file...\t%s", err))
		} else {
			log.Println(lg("Settings", "create", color.HiYellowString, "Created new settings file..."))
			log.Println(lg("Settings", "create", color.HiYellowString,
				"Please save your credentials & info into \"%s\" then restart...", configFile))
			log.Println(lg("Settings", "create", color.MagentaString,
				"You DO NOT NEED token *AND* email/password, just one OR the other."))
			log.Println(lg("Settings", "create", color.MagentaString,
				"THERE ARE MANY HIDDEN SETTINGS AVAILABLE, SEE THE GITHUB README github.com/"+projectRepoBase))
		}
	}
}

func sourceDefault(source *configurationSource) {
	// These have to use the default variables since literal values and consts can't be set to the pointers

	// Setup
	if source.Enabled == nil {
		source.Enabled = &defSource_Enabled
	}
	if source.Save == nil {
		source.Save = &config.Save
	}
	if source.AllowCommands == nil {
		source.AllowCommands = &config.AllowCommands
	}
	if source.SendErrorMessages == nil {
		source.SendErrorMessages = &config.SendErrorMessages
	}
	if source.ScanEdits == nil {
		source.ScanEdits = &config.ScanEdits
	}
	if source.IgnoreBots == nil {
		source.IgnoreBots = &config.IgnoreBots
	}

	if source.SendErrorMessages == nil {
		source.SendErrorMessages = &config.SendErrorMessages
	}
	if source.SendFileToChannel == nil && config.SendFileToChannel != "" {
		source.SendFileToChannel = &config.SendFileToChannel
	}
	if source.SendFileToChannels == nil && config.SendFileToChannels != nil {
		source.SendFileToChannels = &config.SendFileToChannels
	}
	if source.SendFileDirectly == nil {
		source.SendFileDirectly = &config.SendFileDirectly
	}
	if source.SendFileCaption == nil && config.SendFileCaption != "" {
		source.SendFileCaption = &config.SendFileCaption
	}

	// Appearance
	if source.PresenceEnabled == nil {
		source.PresenceEnabled = &config.PresenceEnabled
	}
	if source.ReactWhenDownloaded == nil {
		source.ReactWhenDownloaded = &config.ReactWhenDownloaded
	}
	if source.ReactWhenDownloadedEmoji == nil && config.ReactWhenDownloadedEmoji != nil {
		source.ReactWhenDownloadedEmoji = config.ReactWhenDownloadedEmoji
	}
	if source.ReactWhenDownloadedHistory == nil {
		source.ReactWhenDownloadedHistory = &config.ReactWhenDownloadedHistory
	}
	if source.BlacklistReactEmojis == nil {
		source.BlacklistReactEmojis = &[]string{}
	}
	if source.HistoryTyping == nil && config.HistoryTyping {
		source.HistoryTyping = &config.HistoryTyping
	}
	if source.EmbedColor == nil && config.EmbedColor != nil {
		source.EmbedColor = config.EmbedColor
	}

	// History
	if source.AutoHistory == nil {
		source.AutoHistory = &config.AutoHistory
	}
	if source.AutoHistoryBefore == nil {
		source.AutoHistoryBefore = &config.AutoHistoryBefore
	}
	if source.AutoHistorySince == nil {
		source.AutoHistorySince = &config.AutoHistorySince
	}
	if source.SendAutoHistoryStatus == nil {
		source.SendAutoHistoryStatus = &config.SendAutoHistoryStatus
	}
	if source.SendHistoryStatus == nil {
		source.SendHistoryStatus = &config.SendHistoryStatus
	}
	if source.OutputHistoryStatus == nil {
		source.OutputHistoryStatus = &config.OutputHistoryStatus
	}
	if source.OutputHistoryErrors == nil {
		source.OutputHistoryErrors = &config.OutputHistoryErrors
	}

	// Rules for Saving
	if source.Subfolders == nil {
		source.Subfolders = &config.Subfolders
	}
	if source.SubfoldersFallback == nil {
		source.SubfoldersFallback = &config.SubfoldersFallback
	}
	if source.FilenameDateFormat == nil {
		source.FilenameDateFormat = &config.FilenameDateFormat
	}
	if source.FilenameFormat == nil {
		source.FilenameFormat = &config.FilenameFormat
	}
	if source.FilepathNormalizeText == nil {
		source.FilepathNormalizeText = &config.FilepathNormalizeText
	}
	if source.FilepathStripSymbols == nil {
		source.FilepathStripSymbols = &config.FilepathStripSymbols
	}
	if source.SaveImages == nil {
		source.SaveImages = &config.SaveImages
	}
	if source.SaveVideos == nil {
		source.SaveVideos = &config.SaveVideos
	}
	if source.SaveAudioFiles == nil {
		source.SaveAudioFiles = &config.SaveAudioFiles
	}
	if source.SaveTextFiles == nil {
		source.SaveTextFiles = &config.SaveTextFiles
	}
	if source.SaveOtherFiles == nil {
		source.SaveOtherFiles = &config.SaveOtherFiles
	}
	if source.SavePossibleDuplicates == nil {
		source.SavePossibleDuplicates = &config.SavePossibleDuplicates
	}
	if source.DelayHandling == nil && config.DelayHandling != 0 {
		source.DelayHandling = &config.DelayHandling
	}
	if source.DelayHandlingHistory == nil && config.DelayHandlingHistory != 0 {
		source.DelayHandlingHistory = &config.DelayHandlingHistory
	}
	if source.Filters == nil {
		source.Filters = &configurationSourceFilters{}
	}
	if source.Filters.BlockedPhrases == nil && config.Filters.BlockedPhrases != nil {
		source.Filters.BlockedPhrases = config.Filters.BlockedPhrases
	}
	if source.Filters.AllowedPhrases == nil && config.Filters.AllowedPhrases != nil {
		source.Filters.AllowedPhrases = config.Filters.AllowedPhrases
	}
	if source.Filters.BlockedUsers == nil && config.Filters.BlockedUsers != nil {
		source.Filters.BlockedUsers = config.Filters.BlockedUsers
	}
	if source.Filters.AllowedUsers == nil && config.Filters.AllowedUsers != nil {
		source.Filters.AllowedUsers = config.Filters.AllowedUsers
	}
	if source.Filters.BlockedRoles == nil && config.Filters.BlockedRoles != nil {
		source.Filters.BlockedRoles = config.Filters.BlockedRoles
	}
	if source.Filters.AllowedRoles == nil && config.Filters.AllowedRoles != nil {
		source.Filters.AllowedRoles = config.Filters.AllowedRoles
	}
	if source.Filters.BlockedLinkContent == nil && config.Filters.BlockedLinkContent != nil {
		source.Filters.BlockedLinkContent = config.Filters.BlockedLinkContent
	}
	if source.Filters.AllowedLinkContent == nil && config.Filters.AllowedLinkContent != nil {
		source.Filters.AllowedLinkContent = config.Filters.AllowedLinkContent
	}
	if source.Filters.BlockedDomains == nil && config.Filters.BlockedDomains != nil {
		source.Filters.BlockedDomains = config.Filters.BlockedDomains
	}
	if source.Filters.AllowedDomains == nil && config.Filters.AllowedDomains != nil {
		source.Filters.AllowedDomains = config.Filters.AllowedDomains
	}
	if source.Filters.BlockedExtensions == nil && config.Filters.BlockedExtensions != nil {
		source.Filters.BlockedExtensions = config.Filters.BlockedExtensions
	}
	if source.Filters.AllowedExtensions == nil && config.Filters.AllowedExtensions != nil {
		source.Filters.AllowedExtensions = config.Filters.AllowedExtensions
	}
	if source.Filters.BlockedFilenames == nil && config.Filters.BlockedFilenames != nil {
		source.Filters.BlockedFilenames = config.Filters.BlockedFilenames
	}
	if source.Filters.AllowedFilenames == nil && config.Filters.AllowedFilenames != nil {
		source.Filters.AllowedFilenames = config.Filters.AllowedFilenames
	}
	if source.Filters.BlockedReactions == nil && config.Filters.BlockedReactions != nil {
		source.Filters.BlockedReactions = config.Filters.BlockedReactions
	}
	if source.Filters.AllowedReactions == nil && config.Filters.AllowedReactions != nil {
		source.Filters.AllowedReactions = config.Filters.AllowedReactions
	}
	if source.Duplo == nil && config.Duplo {
		source.Duplo = &config.Duplo
	}
	if source.DuploThreshold == nil && config.DuploThreshold != 0 {
		source.DuploThreshold = &config.DuploThreshold
	}

	// Misc Rules
	if source.LogLinks != nil {
		if source.LogLinks.Subfolders == nil {
			source.LogLinks.Subfolders = &defSourceLog_Subfolders
		}
		if source.LogLinks.SubfoldersFallback == nil {
			source.LogLinks.SubfoldersFallback = &defSourceLog_SubfoldersFallback
		}
		if source.LogLinks.FilenameFormat == nil {
			source.LogLinks.FilenameFormat = &defSourceLog_FilenameFormat
		}
		if source.LogLinks.FilepathNormalizeText == nil {
			source.LogLinks.FilepathNormalizeText = &defSourceLog_FilepathNormalizeText
		}
		if source.LogLinks.FilepathStripSymbols == nil {
			source.LogLinks.FilepathStripSymbols = &defSourceLog_FilepathStripSymbols
		}
		if source.LogLinks.LinePrefix == nil {
			source.LogLinks.LinePrefix = &defSourceLog_LinePrefix
		}
		// vv unique vv
		if source.LogLinks.LineContent == nil {
			source.LogLinks.LineContent = &defSourceLogLink_LineContent
		}
		if source.LogLinks.LogDownloads == nil {
			source.LogLinks.LogDownloads = &defSourceLog_LogDownloads
		}
		if source.LogLinks.LogFailures == nil {
			source.LogLinks.LogFailures = &defSourceLog_LogFailures
		}
	} else if config.LogLinks != nil {
		source.LogLinks = config.LogLinks
	}
	if source.LogMessages != nil {
		if source.LogMessages.Subfolders == nil {
			source.LogMessages.Subfolders = &defSourceLog_Subfolders
		}
		if source.LogMessages.SubfoldersFallback == nil {
			source.LogMessages.SubfoldersFallback = &defSourceLog_SubfoldersFallback
		}
		if source.LogMessages.FilenameFormat == nil {
			source.LogMessages.FilenameFormat = &defSourceLog_FilenameFormat
		}
		if source.LogMessages.FilepathNormalizeText == nil {
			source.LogMessages.FilepathNormalizeText = &defSourceLog_FilepathNormalizeText
		}
		if source.LogMessages.FilepathStripSymbols == nil {
			source.LogMessages.FilepathStripSymbols = &defSourceLog_FilepathStripSymbols
		}
		if source.LogMessages.LinePrefix == nil {
			source.LogMessages.LinePrefix = &defSourceLog_LinePrefix
		}
		// vv unique vv
		if source.LogMessages.LineContent == nil {
			source.LogMessages.LineContent = &defSourceLogMsg_LineContent
		}
	} else if config.LogMessages != nil {
		source.LogMessages = config.LogMessages
	}

	// LAZY CHECKS
	if source.Duplo != nil {
		if *source.Duplo {
			sourceHasDuplo = true
		}
	}
}

var sourceHasDuplo bool = false

func adminChannelDefault(channel *configurationAdminChannel) {
	if channel.LogProgram == nil {
		channel.LogProgram = &adefConfig_LogProgram
	}
	if channel.LogStatus == nil {
		channel.LogStatus = &adefConfig_LogStatus
	}
	if channel.LogErrors == nil {
		channel.LogErrors = &adefConfig_LogErrors
	}
	if channel.UnlockCommands == nil {
		channel.UnlockCommands = &adefConfig_UnlockCommands
	}
}

// Checks if message author is a specified bot admin.
func isBotAdmin(m *discordgo.Message) bool {
	// No Admins or Admin Channels
	if len(config.Admins) == 0 && len(config.AdminChannels) == 0 {
		return true
	}
	// configurationAdminChannel.UnlockCommands Bypass
	if isAdminChannelRegistered(m.ChannelID) {
		sourceConfig := getAdminChannelConfig(m.ChannelID)
		if *sourceConfig.UnlockCommands {
			return true
		}
	}

	return m.Author.ID == botUser.ID || stringInSlice(m.Author.ID, config.Admins)
}

//#endregion

//#region Functions, Admin & Source

var emptySourceConfig configurationSource = configurationSource{}

func getSource(m *discordgo.Message) configurationSource {
	chinfo, err := bot.State.Channel(m.ChannelID)
	if err != nil {
		chinfo, err = bot.Channel(m.ChannelID)
	}
	if err != nil || chinfo == nil {
		log.Println(lg("Settings", "getSource", color.HiRedString, "Failed to find channel info...\t%s", err))
	}

	// Channel
	for _, item := range config.Channels {
		// Single Channel Config
		if m.ChannelID == item.ChannelID || // Standard text channel
			chinfo.ParentID == item.ChannelID { // Nested text channel
			return item
		}
		// Multi-Channel Config
		if item.ChannelIDs != nil {
			for _, subchannel := range *item.ChannelIDs {
				if m.ChannelID == subchannel || // Standard text channel
					chinfo.ParentID == subchannel { // Nested text channel
					return item
				}
			}
		}
	}

	// Category Config
	for _, item := range config.Categories {
		if item.CategoryBlacklist != nil {
			if stringInSlice(chinfo.ID, *item.CategoryBlacklist) {
				return emptySourceConfig
			}
		}
		if item.CategoryID != "" {
			if err == nil {
				if chinfo.ParentID == item.CategoryID {
					return item
				}
			}
		}
		// Multi-Category Config
		if item.CategoryIDs != nil {
			for _, subcategory := range *item.CategoryIDs {
				if err == nil {
					if chinfo.ParentID == subcategory {
						return item
					}
				}
			}
		}
	}

	// Server
	getSourceServer := func(testID string, testSource configurationSource) configurationSource {
		guild, err := bot.State.Guild(testID)
		if err != nil {
			guild, err = bot.Guild(testID)
		}
		if err == nil {
			for _, channel := range guild.Channels {
				if m.ChannelID == channel.ID || // Standard text channel
					chinfo.ParentID == channel.ID { // Nested text channel
					// Channel Blacklisting within Server
					if testSource.ServerBlacklist != nil {
						if stringInSlice(m.ChannelID, *testSource.ServerBlacklist) {
							return emptySourceConfig
						}
						// Categories
						if channel.ParentID != "" {
							if stringInSlice(channel.ParentID, *testSource.ServerBlacklist) {
								return emptySourceConfig
							}
						}
					}
					return testSource
				}
			}
		}
		return emptySourceConfig
	}
	for _, item := range config.Servers {
		if item.ServerID != "" {
			if lookup := getSourceServer(item.ServerID, item); lookup != emptySourceConfig {
				return lookup
			}
		}
		// Multi-Server Config
		if item.ServerIDs != nil {
			for _, subserver := range *item.ServerIDs {
				if lookup := getSourceServer(subserver, item); lookup != emptySourceConfig {
					return lookup
				}
			}
		}
	}

	// User Config
	if m.Author != nil {
		for _, item := range config.Users {
			if item.UserID != "" {
				if m.Author.ID == item.UserID {
					return item
				}
			}
			// Multi-User Config
			if item.UserIDs != nil {
				for _, subuser := range *item.UserIDs {
					if m.Author.ID == subuser {
						return item
					}
				}
			}
		}
	}

	// All
	if config.All != nil {
		if config.AllBlacklistChannels != nil {
			if stringInSlice(m.ChannelID, *config.AllBlacklistChannels) ||
				stringInSlice(chinfo.ParentID, *config.AllBlacklistChannels) {
				return emptySourceConfig
			}
		}
		if config.AllBlacklistCategories != nil {
			chinf, err := bot.State.Channel(m.ChannelID)
			if err != nil {
				chinf, err = bot.Channel(m.ChannelID)
			}
			if err == nil {
				if stringInSlice(chinf.ParentID, *config.AllBlacklistCategories) || stringInSlice(m.ChannelID, *config.AllBlacklistCategories) {
					return emptySourceConfig
				}
			}
		}
		if config.AllBlacklistServers != nil {
			if stringInSlice(m.GuildID, *config.AllBlacklistServers) {
				return emptySourceConfig
			}
		}
		if config.AllBlacklistUsers != nil && m.Author != nil {
			if stringInSlice(m.Author.ID, *config.AllBlacklistUsers) {
				return emptySourceConfig
			}
		}
		return *config.All
	}

	return emptySourceConfig
}

var emptyAdminChannelConfig configurationAdminChannel = configurationAdminChannel{}

func isAdminChannelRegistered(ChannelID string) bool {
	if config.AdminChannels != nil {
		for _, item := range config.AdminChannels {
			// Single Channel Config
			if ChannelID == item.ChannelID {
				return true
			}
			// Multi-Channel Config
			if item.ChannelIDs != nil {
				if stringInSlice(ChannelID, *item.ChannelIDs) {
					return true
				}
			}
		}
	}
	return false
}

func getAdminChannelConfig(ChannelID string) configurationAdminChannel {
	if config.AdminChannels != nil {
		for _, item := range config.AdminChannels {
			// Single Channel Config
			if ChannelID == item.ChannelID {
				return item
			}
			// Multi-Channel Config
			if item.ChannelIDs != nil {
				for _, subchannel := range *item.ChannelIDs {
					if ChannelID == subchannel {
						return item
					}
				}
			}
		}
	}
	return configurationAdminChannel{}
}

func isCommandableChannel(m *discordgo.Message) bool {
	if config.AllowGeneralCommands {
		return true
	}
	if isAdminChannelRegistered(m.ChannelID) {
		return true
	} else if sourceConfig := getSource(m); sourceConfig != emptySourceConfig {
		if *sourceConfig.AllowCommands || isBotAdmin(m) || m.Author.ID == bot.State.User.ID {
			return true
		}
	}
	return false
}

func getBoundUsers() []string {
	var users []string
	for _, item := range config.Users {
		if item.UserID != "" {
			if !stringInSlice(item.UserID, users) {
				users = append(users, item.UserID)
			}
		} else if item.UserIDs != nil {
			for _, subuser := range *item.UserIDs {
				if subuser != "" {
					if !stringInSlice(subuser, users) {
						users = append(users, subuser)
					}
				}
			}
		}
	}
	return users
}

func getBoundUsersCount() int {
	return len(getBoundUsers())
}

func getBoundServers() []string {
	var servers []string
	for _, item := range config.Servers {
		if item.ServerID != "" {
			if !stringInSlice(item.ServerID, servers) {
				servers = append(servers, item.ServerID)
			}
		} else if item.ServerIDs != nil {
			for _, subserver := range *item.ServerIDs {
				if subserver != "" {
					if !stringInSlice(subserver, servers) {
						servers = append(servers, subserver)
					}
				}
			}
		}
	}
	return servers
}

func getBoundServersCount() int {
	return len(getBoundServers())
}

func getBoundChannels() []string {
	var channels []string
	for _, item := range config.Channels {
		if item.ChannelID != "" {
			if !stringInSlice(item.ChannelID, channels) {
				channels = append(channels, item.ChannelID)
			}
		} else if item.ChannelIDs != nil {
			for _, subchannel := range *item.ChannelIDs {
				if subchannel != "" {
					if !stringInSlice(subchannel, channels) {
						channels = append(channels, subchannel)
					}
				}
			}
		}
	}
	return channels
}

func getBoundChannelsCount() int {
	return len(getBoundChannels())
}

func getBoundCategories() []string {
	var categories []string
	for _, item := range config.Categories {
		if item.CategoryID != "" {
			if !stringInSlice(item.CategoryID, categories) {
				categories = append(categories, item.CategoryID)
			}
		} else if item.CategoryIDs != nil {
			for _, subcategory := range *item.CategoryIDs {
				if subcategory != "" {
					if !stringInSlice(subcategory, categories) {
						categories = append(categories, subcategory)
					}
				}
			}
		}
	}
	return categories
}

func getBoundCategoriesCount() int {
	return len(getBoundCategories())
}

type registeredChannelSource struct {
	ChannelID string
	Source    configurationSource
}

func getAllRegisteredChannels() []registeredChannelSource {
	var channels []registeredChannelSource
	if config.All != nil { // ALL MODE
		for _, guild := range bot.State.Guilds {
			if config.AllBlacklistServers != nil {
				if stringInSlice(guild.ID, *config.AllBlacklistServers) {
					continue
				}
			}
			for _, channel := range guild.Channels {
				if r := getSource(&discordgo.Message{ChannelID: channel.ID}); r == emptySourceConfig { // easier than redoing it all but way less efficient, im lazy
					continue
				} else {
					if hasPerms(channel.ID, discordgo.PermissionViewChannel) && hasPerms(channel.ID, discordgo.PermissionReadMessageHistory) {
						channels = append(channels, registeredChannelSource{channel.ID, *config.All})
					}
				}
			}
		}
	} else { // STANDARD MODE
		// Compile all channels sourced from config servers
		for _, server := range config.Servers {
			if server.ServerIDs != nil {
				for _, subserver := range *server.ServerIDs {
					guild, err := bot.State.Guild(subserver)
					if err != nil {
						guild, err = bot.Guild(subserver)
					}
					if err == nil {
						for _, channel := range guild.Channels {
							if hasPerms(channel.ID, discordgo.PermissionReadMessageHistory) {
								channels = append(channels, registeredChannelSource{channel.ID, server})
							}
						}
					}
				}
			} else if isNumeric(server.ServerID) {
				guild, err := bot.State.Guild(server.ServerID)
				if err != nil {
					guild, err = bot.Guild(server.ServerID)
				}
				if err == nil {
					for _, channel := range guild.Channels {
						if hasPerms(channel.ID, discordgo.PermissionReadMessageHistory) {
							channels = append(channels, registeredChannelSource{channel.ID, server})
						}
					}
				}
			}
		}
		// Compile all config channels under categories, no practical way to poll these other than checking state.
		for _, guild := range bot.State.Guilds {
			for _, channel := range guild.Channels {
				for _, source := range config.Categories {
					if source.CategoryIDs != nil {
						for _, category := range *source.CategoryIDs {
							if channel.ParentID == category {
								channels = append(channels, registeredChannelSource{channel.ID, source})
							}
						}
					} else if isNumeric(source.CategoryID) {
						if channel.ParentID == source.CategoryID {
							channels = append(channels, registeredChannelSource{channel.ID, source})
						}
					}
				}
			}
		}
		// Compile all config channels
		for _, channel := range config.Channels {
			if channel.ChannelIDs != nil {
				for _, subchannel := range *channel.ChannelIDs {
					channels = append(channels, registeredChannelSource{subchannel, channel})
				}
			} else if isNumeric(channel.ChannelID) {
				channels = append(channels, registeredChannelSource{channel.ChannelID, channel})
			}
		}
	}
	return channels
}

//#endregion
