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
	config = defaultConfiguration()
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
	TwitterAuthEnabled       *bool  `json:"twitterEnabled" yaml:"twitterEnabled"`
	TwitterUsername          string `json:"twitterUsername" yaml:"twitterUsername"`
	TwitterPassword          string `json:"twitterPassword" yaml:"twitterPassword"`
	TwitterProxy             string `json:"twitterProxy,omitempty" yaml:"twitterProxy,omitempty"`
	InstagramAuthEnabled     *bool  `json:"instagramEnabled" yaml:"instagramEnabled"`
	InstagramUsername        string `json:"instagramUsername" yaml:"instagramUsername"`
	InstagramPassword        string `json:"instagramPassword" yaml:"instagramPassword"`
	InstagramProxy           string `json:"instagramProxy,omitempty" yaml:"instagramProxy,omitempty"`
	InstagramProxyInsecure   *bool  `json:"instagramProxyInsecure,omitempty" yaml:"instagramProxyInsecure,omitempty"`
	InstagramProxyForceHTTP2 *bool  `json:"instagramProxyForceHTTP2,omitempty" yaml:"instagramProxyForceHTTP2,omitempty"`
	FlickrApiKey             string `json:"flickrApiKey" yaml:"flickrApiKey"`
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
		SettingsOutput:        true,
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
		ConnectionCheckRate: defConfig_ConnectionCheckRate,
		PresenceRefreshRate: defConfig_PresenceRefreshRate,

		// Source Setup Defaults
		Save:          true,
		AllowCommands: true,
		ScanEdits:     true,
		IgnoreBots:    true,

		SendErrorMessages: false,
		SendFileToChannel: "",
		SendFileDirectly:  true,
		SendFileCaption:   "",

		FilenameDateFormat: defConfig_FilenameDateFormat,
		FilenameFormat:     defConfig_FilenameFormat,

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
		HistoryRequestCount:   100,
		HistoryRequestDelay:   0,

		// Rules for Saving
		DelayHandling:          0,
		DelayHandlingHistory:   0,
		DivideByYear:           false,
		DivideByMonth:          false,
		DivideByDay:            false,
		DivideByHour:           false,
		DivideByServer:         false,
		DivideByChannel:        false,
		DivideByUser:           false,
		DivideByType:           true,
		DivideFoldersUseID:     false,
		SaveImages:             true,
		SaveVideos:             true,
		SaveAudioFiles:         true,
		SaveTextFiles:          false,
		SaveOtherFiles:         false,
		SavePossibleDuplicates: false,
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

	// Program Settings
	LogOutput             string `json:"logOutput,omitempty" yaml:"logOutput,omitempty"`
	LogIndent             bool   `json:"logIndent" yaml:"logIndent"`
	ProcessLimit          int    `json:"processLimit" yaml:"processLimit"`
	Debug                 bool   `json:"debug" yaml:"debug"`
	BackupDatabaseOnStart bool   `json:"backupDatabaseOnStart" yaml:"backupDatabaseOnStart"`
	WatchSettings         bool   `json:"watchSettings" yaml:"watchSettings"`
	SettingsOutput        bool   `json:"settingsOutput" yaml:"settingsOutput"`
	MessageOutput         bool   `json:"messageOutput" yaml:"messageOutput"`
	MessageOutputHistory  bool   `json:"messageOutputHistory" yaml:"messageOutputHistory"`

	DiscordLogLevel      int  `json:"discordLogLevel" yaml:"discordLogLevel"`
	DiscordTimeout       int  `json:"discordTimeout" yaml:"discordTimeout"`
	DownloadTimeout      int  `json:"downloadTimeout" yaml:"downloadTimeout"`
	DownloadRetryMax     int  `json:"downloadRetryMax" yaml:"downloadRetryMax"`
	ExitOnBadConnection  bool `json:"exitOnBadConnection" yaml:"exitOnBadConnection"`
	GithubUpdateChecking bool `json:"githubUpdateChecking" yaml:"githubUpdateChecking"`

	CommandPrefix        string `json:"commandPrefix" yaml:"commandPrefix"`
	CommandTagging       bool   `json:"commandTagging" yaml:"commandTagging"`
	ScanOwnMessages      bool   `json:"scanOwnMessages" yaml:"scanOwnMessages"`
	AllowGeneralCommands bool   `json:"allowGeneralCommands" yaml:"allowGeneralCommands"`
	InflateDownloadCount *int64 `json:"inflateDownloadCount,omitempty" yaml:"inflateDownloadCount,omitempty"`
	EuropeanNumbers      bool   `json:"europeanNumbers,omitempty" yaml:"europeanNumbers,omitempty"`

	HistoryManagerRate  int `json:"historyManagerRate,omitempty" yaml:"historyManagerRate,omitempty"`
	CheckupRate         int `json:"checkupRate,omitempty" yaml:"checkupRate,omitempty"`
	ConnectionCheckRate int `json:"connectionCheckRate,omitempty" yaml:"connectionCheckRate,omitempty"`
	PresenceRefreshRate int `json:"presenceRefreshRate,omitempty" yaml:"presenceRefreshRate,omitempty"`

	// Source Setup Defaults
	Save          bool `json:"save" yaml:"save"`
	AllowCommands bool `json:"allowCommands" yaml:"allowCommands"`
	ScanEdits     bool `json:"scanEdits" yaml:"scanEdits"`
	IgnoreBots    bool `json:"ignoreBots" yaml:"ignoreBots"`

	SendErrorMessages  bool     `json:"sendErrorMessages" yaml:"sendErrorMessages"`
	SendFileToChannel  string   `json:"sendFileToChannel" yaml:"sendFileToChannel"`
	SendFileToChannels []string `json:"sendFileToChannels,omitempty" yaml:"sendFileToChannels,omitempty"`
	SendFileDirectly   bool     `json:"sendFileDirectly,omitempty" yaml:"sendFileDirectly,omitempty"`
	SendFileCaption    string   `json:"sendFileCaption,omitempty" yaml:"sendFileCaption,omitempty"`

	FilenameDateFormat string `json:"filenameDateFormat" yaml:"filenameDateFormat"`
	FilenameFormat     string `json:"filenameFormat" yaml:"filenameFormat"`

	// Discord Presence
	PresenceEnabled bool               `json:"presenceEnabled" yaml:"presenceEnabled"`
	PresenceStatus  string             `json:"presenceStatus" yaml:"presenceStatus"`
	PresenceType    discordgo.GameType `json:"presenceType" yaml:"presenceType"`
	PresenceLabel   *string            `json:"presenceLabel" yaml:"presenceLabel"`
	PresenceDetails *string            `json:"presenceDetails" yaml:"presenceDetails"`
	PresenceState   *string            `json:"presenceState" yaml:"presenceState"`

	// Appearance
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
	HistoryRequestCount   int    `json:"historyRequestCount" yaml:"historyRequestCount"`
	HistoryRequestDelay   int    `json:"historyRequestDelay" yaml:"historyRequestDelay"`

	// Rules for Saving
	DelayHandling          int                         `json:"delayHandling,omitempty" yaml:"delayHandling,omitempty"`
	DelayHandlingHistory   int                         `json:"delayHandlingHistory,omitempty" yaml:"delayHandlingHistory,omitempty"`
	DivideByYear           bool                        `json:"divideByYear" yaml:"divideByYear"`
	DivideByMonth          bool                        `json:"divideByMonth" yaml:"divideByMonth"`
	DivideByDay            bool                        `json:"divideByDay" yaml:"divideByDay"`
	DivideByHour           bool                        `json:"divideByHour" yaml:"divideByHour"`
	DivideByServer         bool                        `json:"divideByServer" yaml:"divideByServer"`
	DivideByChannel        bool                        `json:"divideByChannel" yaml:"divideByChannel"`
	DivideByUser           bool                        `json:"divideByUser" yaml:"divideByUser"`
	DivideByType           bool                        `json:"divideByType" yaml:"divideByType"`
	DivideFoldersUseID     bool                        `json:"divideFoldersUseID" yaml:"divideFoldersUseID"`
	SaveImages             bool                        `json:"saveImages" yaml:"saveImages"`
	SaveVideos             bool                        `json:"saveVideos" yaml:"saveVideos"`
	SaveAudioFiles         bool                        `json:"saveAudioFiles" yaml:"saveAudioFiles"`
	SaveTextFiles          bool                        `json:"saveTextFiles" yaml:"saveTextFiles"`
	SaveOtherFiles         bool                        `json:"saveOtherFiles" yaml:"saveOtherFiles"`
	SavePossibleDuplicates bool                        `json:"savePossibleDuplicates" yaml:"savePossibleDuplicates"`
	Filters                *configurationSourceFilters `json:"filters" yaml:"filters"`
	Duplo                  bool                        `json:"duplo,omitempty" yaml:"duplo,omitempty"`
	DuploThreshold         float64                     `json:"duploThreshold,omitempty" yaml:"duploThreshold,omitempty"`

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

type constStruct struct {
	Constants map[string]string `json:"_constants" yaml:"_constants"`
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

	// Setup
	Enabled       *bool `json:"enabled" yaml:"enabled"`
	Save          *bool `json:"save" yaml:"save"`
	AllowCommands *bool `json:"allowCommands" yaml:"allowCommands"`
	ScanEdits     *bool `json:"scanEdits" yaml:"scanEdits"`
	IgnoreBots    *bool `json:"ignoreBots" yaml:"ignoreBots"`

	SendErrorMessages  *bool     `json:"sendErrorMessages" yaml:"sendErrorMessages"`
	SendFileToChannel  *string   `json:"sendFileToChannel" yaml:"sendFileToChannel"`
	SendFileToChannels *[]string `json:"sendFileToChannels,omitempty" yaml:"sendFileToChannels,omitempty"`
	SendFileDirectly   *bool     `json:"sendFileDirectly,omitempty" yaml:"sendFileDirectly,omitempty"`
	SendFileCaption    *string   `json:"sendFileCaption,omitempty" yaml:"sendFileCaption,omitempty"`

	FilenameDateFormat *string `json:"filenameDateFormat" yaml:"filenameDateFormat"`
	FilenameFormat     *string `json:"filenameFormat" yaml:"filenameFormat"`

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

	// Rules for Saving
	DelayHandling          *int                        `json:"delayHandling,omitempty" yaml:"delayHandling,omitempty"`
	DelayHandlingHistory   *int                        `json:"delayHandlingHistory,omitempty" yaml:"delayHandlingHistory,omitempty"`
	DivideByYear           *bool                       `json:"divideByYear" yaml:"divideByYear"`
	DivideByMonth          *bool                       `json:"divideByMonth" yaml:"divideByMonth"`
	DivideByDay            *bool                       `json:"divideByDay" yaml:"divideByDay"`
	DivideByHour           *bool                       `json:"divideByHour" yaml:"divideByHour"`
	DivideByServer         *bool                       `json:"divideByServer" yaml:"divideByServer"`
	DivideByChannel        *bool                       `json:"divideByChannel" yaml:"divideByChannel"`
	DivideByUser           *bool                       `json:"divideByUser" yaml:"divideByUser"`
	DivideByType           *bool                       `json:"divideByType" yaml:"divideByType"`
	DivideFoldersUseID     *bool                       `json:"divideFoldersUseID" yaml:"divideFoldersUseID"`
	SaveImages             *bool                       `json:"saveImages" yaml:"saveImages"`
	SaveVideos             *bool                       `json:"saveVideos" yaml:"saveVideos"`
	SaveAudioFiles         *bool                       `json:"saveAudioFiles" yaml:"saveAudioFiles"`
	SaveTextFiles          *bool                       `json:"saveTextFiles" yaml:"saveTextFiles"`
	SaveOtherFiles         *bool                       `json:"saveOtherFiles" yaml:"saveOtherFiles"`
	SavePossibleDuplicates *bool                       `json:"savePossibleDuplicates" yaml:"savePossibleDuplicates"`
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
	defSourceLog_DestinationIsFolder bool = false
	defSourceLog_DivideLogsByServer  bool = true
	defSourceLog_DivideLogsByChannel bool = true
	defSourceLog_DivideLogsByUser    bool = false
	defSourceLog_DivideLogsByStatus  bool = false
	defSourceLog_LogDownloads        bool = true
	defSourceLog_LogFailures         bool = true
)

type configurationSourceLog struct {
	Destination         string  `json:"destination" yaml:"destination"`
	DestinationIsFolder *bool   `json:"destinationIsFolder" yaml:"destinationIsFolder"`
	DivideLogsByServer  *bool   `json:"divideLogsByServer" yaml:"divideLogsByServer"`
	DivideLogsByChannel *bool   `json:"divideLogsByChannel" yaml:"divideLogsByChannel"`
	DivideLogsByUser    *bool   `json:"divideLogsByUser" yaml:"divideLogsByUser"`
	DivideLogsByStatus  *bool   `json:"divideLogsByStatus" yaml:"divideLogsByStatus"`
	LogDownloads        *bool   `json:"logDownloads" yaml:"logDownloads"`
	LogFailures         *bool   `json:"logFailures" yaml:"logFailures"`
	FilterDuplicates    *bool   `json:"filterDuplicates" yaml:"filterDuplicates"`
	Prefix              *string `json:"prefix" yaml:"prefix"`
	Suffix              *string `json:"suffix" yaml:"suffix"`
	UserData            *bool   `json:"userData" yaml:"userData"`
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
	ChannelID      string    `json:"channel" yaml:"channel"`
	ChannelIDs     *[]string `json:"channels,omitempty" yaml:"channels,omitempty"`
	LogProgram     *bool     `json:"logProgram" yaml:"logProgram"`
	LogStatus      *bool     `json:"logStatus" yaml:"logStatus"`
	LogErrors      *bool     `json:"logErrors" yaml:"logErrors"`
	UnlockCommands *bool     `json:"unlockCommands" yaml:"unlockCommands"`
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
		if config.SettingsOutput {
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
			allString = ", ALL ENABLED"
		}
		log.Println(lg("Settings", "", color.HiYellowString,
			"Loaded - bound to %d channel%s, %d categories, %d server%s, %d user%s%s",
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

func sourceDefault(channel *configurationSource) {
	// These have to use the default variables since literal values and consts can't be set to the pointers

	// Setup
	if channel.Enabled == nil {
		channel.Enabled = &defSource_Enabled
	}
	if channel.Save == nil {
		channel.Save = &config.Save
	}
	if channel.AllowCommands == nil {
		channel.AllowCommands = &config.AllowCommands
	}
	if channel.SendErrorMessages == nil {
		channel.SendErrorMessages = &config.SendErrorMessages
	}
	if channel.ScanEdits == nil {
		channel.ScanEdits = &config.ScanEdits
	}
	if channel.IgnoreBots == nil {
		channel.IgnoreBots = &config.IgnoreBots
	}

	if channel.SendErrorMessages == nil {
		channel.SendErrorMessages = &config.SendErrorMessages
	}
	if channel.SendFileToChannel == nil && config.SendFileToChannel != "" {
		channel.SendFileToChannel = &config.SendFileToChannel
	}
	if channel.SendFileToChannels == nil && config.SendFileToChannels != nil {
		channel.SendFileToChannels = &config.SendFileToChannels
	}
	if channel.SendFileDirectly == nil {
		channel.SendFileDirectly = &config.SendFileDirectly
	}
	if channel.SendFileCaption == nil && config.SendFileCaption != "" {
		channel.SendFileCaption = &config.SendFileCaption
	}

	if channel.FilenameDateFormat == nil {
		channel.FilenameDateFormat = &config.FilenameDateFormat
	}
	if channel.FilenameFormat == nil {
		channel.FilenameFormat = &config.FilenameFormat
	}

	// Appearance
	if channel.PresenceEnabled == nil {
		channel.PresenceEnabled = &config.PresenceEnabled
	}
	if channel.ReactWhenDownloaded == nil {
		channel.ReactWhenDownloaded = &config.ReactWhenDownloaded
	}
	if channel.ReactWhenDownloadedEmoji == nil && config.ReactWhenDownloadedEmoji != nil {
		channel.ReactWhenDownloadedEmoji = config.ReactWhenDownloadedEmoji
	}
	if channel.ReactWhenDownloadedHistory == nil {
		channel.ReactWhenDownloadedHistory = &config.ReactWhenDownloadedHistory
	}
	if channel.BlacklistReactEmojis == nil {
		channel.BlacklistReactEmojis = &[]string{}
	}
	if channel.HistoryTyping == nil && config.HistoryTyping {
		channel.HistoryTyping = &config.HistoryTyping
	}
	if channel.EmbedColor == nil && config.EmbedColor != nil {
		channel.EmbedColor = config.EmbedColor
	}

	// History
	if channel.AutoHistory == nil {
		channel.AutoHistory = &config.AutoHistory
	}
	if channel.AutoHistoryBefore == nil {
		channel.AutoHistoryBefore = &config.AutoHistoryBefore
	}
	if channel.AutoHistorySince == nil {
		channel.AutoHistorySince = &config.AutoHistorySince
	}
	if channel.SendAutoHistoryStatus == nil {
		channel.SendAutoHistoryStatus = &config.SendAutoHistoryStatus
	}
	if channel.SendHistoryStatus == nil {
		channel.SendHistoryStatus = &config.SendHistoryStatus
	}

	// Rules for Saving
	if channel.DelayHandling == nil && config.DelayHandling != 0 {
		channel.DelayHandling = &config.DelayHandling
	}
	if channel.DelayHandlingHistory == nil && config.DelayHandlingHistory != 0 {
		channel.DelayHandlingHistory = &config.DelayHandlingHistory
	}
	if channel.DivideByYear == nil {
		channel.DivideByYear = &config.DivideByYear
	}
	if channel.DivideByMonth == nil {
		channel.DivideByMonth = &config.DivideByMonth
	}
	if channel.DivideByDay == nil {
		channel.DivideByDay = &config.DivideByDay
	}
	if channel.DivideByHour == nil {
		channel.DivideByHour = &config.DivideByHour
	}
	if channel.DivideByServer == nil {
		channel.DivideByServer = &config.DivideByServer
	}
	if channel.DivideByChannel == nil {
		channel.DivideByChannel = &config.DivideByChannel
	}
	if channel.DivideByUser == nil {
		channel.DivideByUser = &config.DivideByUser
	}
	if channel.DivideByType == nil {
		channel.DivideByType = &config.DivideByType
	}
	if channel.DivideFoldersUseID == nil {
		channel.DivideFoldersUseID = &config.DivideFoldersUseID
	}
	if channel.SaveImages == nil {
		channel.SaveImages = &config.SaveImages
	}
	if channel.SaveVideos == nil {
		channel.SaveVideos = &config.SaveVideos
	}
	if channel.SaveAudioFiles == nil {
		channel.SaveAudioFiles = &config.SaveAudioFiles
	}
	if channel.SaveTextFiles == nil {
		channel.SaveTextFiles = &config.SaveTextFiles
	}
	if channel.SaveOtherFiles == nil {
		channel.SaveOtherFiles = &config.SaveOtherFiles
	}
	if channel.SavePossibleDuplicates == nil {
		channel.SavePossibleDuplicates = &config.SavePossibleDuplicates
	}
	if channel.Filters == nil {
		channel.Filters = &configurationSourceFilters{}
	}
	if channel.Filters.BlockedExtensions == nil && config.Filters.BlockedExtensions != nil {
		channel.Filters.BlockedExtensions = config.Filters.BlockedExtensions
	}
	if channel.Filters.BlockedPhrases == nil && config.Filters.BlockedPhrases != nil {
		channel.Filters.BlockedPhrases = config.Filters.BlockedPhrases
	}
	if channel.Duplo == nil && config.Duplo {
		channel.Duplo = &config.Duplo
	}
	if channel.DuploThreshold == nil && config.DuploThreshold != 0 {
		channel.DuploThreshold = &config.DuploThreshold
	}

	// Misc Rules
	if channel.LogLinks != nil {
		if channel.LogLinks.DestinationIsFolder == nil {
			channel.LogLinks.DestinationIsFolder = &defSourceLog_DestinationIsFolder
		}
		if channel.LogLinks.DivideLogsByServer == nil {
			channel.LogLinks.DivideLogsByServer = &defSourceLog_DivideLogsByServer
		}
		if channel.LogLinks.DivideLogsByChannel == nil {
			channel.LogLinks.DivideLogsByChannel = &defSourceLog_DivideLogsByChannel
		}
		if channel.LogLinks.DivideLogsByUser == nil {
			channel.LogLinks.DivideLogsByUser = &defSourceLog_DivideLogsByUser
		}
		if channel.LogLinks.DivideLogsByStatus == nil {
			channel.LogLinks.DivideLogsByStatus = &defSourceLog_DivideLogsByStatus
		}
		if channel.LogLinks.LogDownloads == nil {
			channel.LogLinks.LogDownloads = &defSourceLog_LogDownloads
		}
		if channel.LogLinks.LogFailures == nil {
			channel.LogLinks.LogFailures = &defSourceLog_LogFailures
		}
	}

	if channel.LogMessages != nil {
		if channel.LogMessages.DestinationIsFolder == nil {
			channel.LogMessages.DestinationIsFolder = &defSourceLog_DestinationIsFolder
		}
		if channel.LogMessages.DivideLogsByServer == nil {
			channel.LogMessages.DivideLogsByServer = &defSourceLog_DivideLogsByServer
		}
		if channel.LogMessages.DivideLogsByChannel == nil {
			channel.LogMessages.DivideLogsByChannel = &defSourceLog_DivideLogsByChannel
		}
		if channel.LogMessages.DivideLogsByUser == nil {
			channel.LogMessages.DivideLogsByUser = &defSourceLog_DivideLogsByUser
		}
	}

	// LAZY CHECKS
	if channel.Duplo != nil {
		if *channel.Duplo {
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

func isNestedMessage(subjectMessage *discordgo.Message, targetChannel string) bool {
	if subjectMessage.ID != "" {
		_, err := bot.State.Message(targetChannel, subjectMessage.ChannelID)
		return err == nil
	}
	return false
}

var emptyConfig configurationSource = configurationSource{}

func getSource(m *discordgo.Message, c *discordgo.Channel) configurationSource {

	subjectID := m.ChannelID

	if c != nil {
		if (c.Type == discordgo.ChannelTypeGuildPublicThread ||
			c.Type == discordgo.ChannelTypeGuildPrivateThread ||
			c.Type == discordgo.ChannelTypeGuildNewsThread) && c.ParentID != "" {
			subjectID = c.ParentID
		}
	}

	// Channel
	for _, item := range config.Channels {
		// Single Channel Config
		if subjectID == item.ChannelID || isNestedMessage(m, item.ChannelID) {
			return item
		}
		// Multi-Channel Config
		if item.ChannelIDs != nil {
			for _, subchannel := range *item.ChannelIDs {
				if subjectID == subchannel || isNestedMessage(m, subchannel) {
					return item
				}
			}
		}
	}

	// Category Config
	channel, err := bot.State.Channel(subjectID)
	if err != nil {
		channel, err = bot.Channel(subjectID)
	}
	for _, item := range config.Categories {
		if item.CategoryBlacklist != nil {
			if stringInSlice(channel.ID, *item.CategoryBlacklist) {
				return emptyConfig
			}
		}
		if item.CategoryID != "" {
			if err == nil {
				if channel.ParentID == item.CategoryID {
					return item
				}
			}
		}
		// Multi-Category Config
		if item.CategoryIDs != nil {
			for _, subcategory := range *item.CategoryIDs {
				if err == nil {
					if channel.ParentID == subcategory {
						return item
					}
				}
			}
		}
	}

	// Server
	for _, item := range config.Servers {
		if item.ServerID != "" {
			guild, err := bot.State.Guild(item.ServerID)
			if err != nil {
				guild, err = bot.Guild(item.ServerID)
			}
			if err == nil {
				for _, channel := range guild.Channels {
					if subjectID == channel.ID || isNestedMessage(m, channel.ID) {
						// Channel Blacklisting within Server
						if item.ServerBlacklist != nil {
							if stringInSlice(subjectID, *item.ServerBlacklist) {
								return emptyConfig
							}
							// Categories
							if channel.ParentID != "" {
								if stringInSlice(channel.ParentID, *item.ServerBlacklist) {
									return emptyConfig
								}
							}
						}
						return item
					}
				}
			}
		}
		// Multi-Server Config
		if item.ServerIDs != nil {
			for _, subserver := range *item.ServerIDs {
				guild, err := bot.State.Guild(subserver)
				if err != nil {
					guild, err = bot.Guild(subserver)
				}
				if err == nil {
					for _, channel := range guild.Channels {
						if subjectID == channel.ID || isNestedMessage(m, channel.ID) {
							// Channel Blacklisting within Servers
							if item.ServerBlacklist != nil {
								if stringInSlice(subjectID, *item.ServerBlacklist) {
									return emptyConfig
								}
								// Categories
								if channel.ParentID != "" {
									if stringInSlice(channel.ParentID, *item.ServerBlacklist) {
										return emptyConfig
									}
								}
							}
							return item
						}
					}
				}
			}
		}
	}

	// User Config
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

	// All
	if config.All != nil {
		if config.AllBlacklistChannels != nil {
			if stringInSlice(subjectID, *config.AllBlacklistChannels) {
				return emptyConfig
			}
		}
		if config.AllBlacklistCategories != nil {
			chinf, err := bot.State.Channel(subjectID)
			if err == nil {
				if stringInSlice(chinf.ParentID, *config.AllBlacklistCategories) || stringInSlice(subjectID, *config.AllBlacklistCategories) {
					return emptyConfig
				}
			}
		}
		if config.AllBlacklistServers != nil {
			if stringInSlice(m.GuildID, *config.AllBlacklistServers) {
				return emptyConfig
			}
		}
		if config.AllBlacklistUsers != nil && m.Author != nil {
			if stringInSlice(m.Author.ID, *config.AllBlacklistUsers) {
				return emptyConfig
			}
		}
		return *config.All
	}

	return emptyConfig
}

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
	} else if sourceConfig := getSource(m, nil); sourceConfig != emptyConfig {
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

func getAllRegisteredChannels() []string {
	var channels []string
	if config.All != nil { // ALL MODE
		for _, guild := range bot.State.Guilds {
			if config.AllBlacklistServers != nil {
				if stringInSlice(guild.ID, *config.AllBlacklistServers) {
					continue
				}
			}
			for _, channel := range guild.Channels {
				if r := getSource(&discordgo.Message{ChannelID: channel.ID}, nil); r == emptyConfig { // easier than redoing it all but way less efficient, im lazy
					continue
				} else {
					if hasPerms(channel.ID, discordgo.PermissionViewChannel) && hasPerms(channel.ID, discordgo.PermissionReadMessageHistory) {
						channels = append(channels, channel.ID)
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
					if err == nil {
						for _, channel := range guild.Channels {
							if hasPerms(channel.ID, discordgo.PermissionReadMessageHistory) {
								channels = append(channels, channel.ID)
							}
						}
					}
				}
			} else if isNumeric(server.ServerID) {
				guild, err := bot.State.Guild(server.ServerID)
				if err == nil {
					for _, channel := range guild.Channels {
						if hasPerms(channel.ID, discordgo.PermissionReadMessageHistory) {
							channels = append(channels, channel.ID)
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
								channels = append(channels, channel.ID)
							}
						}
					} else if isNumeric(source.CategoryID) {
						if channel.ParentID == source.CategoryID {
							channels = append(channels, channel.ID)
						}
					}
				}
			}
		}
		// Compile all config channels
		for _, channel := range config.Channels {
			if channel.ChannelIDs != nil {
				channels = append(channels, *channel.ChannelIDs...)
			} else if isNumeric(channel.ChannelID) {
				channels = append(channels, channel.ChannelID)
			}
		}
	}
	return channels
}

//#endregion
