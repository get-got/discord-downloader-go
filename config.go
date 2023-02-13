package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/muhammadmuzzammil1998/jsonc"
	"gopkg.in/ini.v1"
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
	Token    string `json:"token,omitempty"`    // required for bot token (this or login)
	Email    string `json:"email,omitempty"`    // required for login (this or token)
	Password string `json:"password,omitempty"` // required for login (this or token)
	// APIs
	TwitterAccessToken       string `json:"twitterAccessToken,omitempty"`       // optional
	TwitterAccessTokenSecret string `json:"twitterAccessTokenSecret,omitempty"` // optional
	TwitterConsumerKey       string `json:"twitterConsumerKey,omitempty"`       // optional
	TwitterConsumerSecret    string `json:"twitterConsumerSecret,omitempty"`    // optional
	InstagramUsername        string `json:"instagramUsername,omitempty"`        // optional
	InstagramPassword        string `json:"instagramPassword,omitempty"`        // optional
	FlickrApiKey             string `json:"flickrApiKey,omitempty"`             // optional
}

//#endregion

//#region Config, Main

// defConfig_ = Config Default
// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	defConfig_Debug                bool   = false
	defConfig_CommandPrefix        string = "ddg "
	defConfig_ScanOwnMessages      bool   = false
	defConfig_GithubUpdateChecking bool   = true
	// Appearance
	defConfig_PresenceEnabled      bool               = true
	defConfig_PresenceStatus       string             = string(discordgo.StatusIdle)
	defConfig_PresenceType         discordgo.GameType = discordgo.GameTypeGame
	defConfig_ReactWhenDownloaded  bool               = true
	defConfig_InflateDownloadCount int64              = 0

	// These are only defaults to "fix" when loading settings for when people put stupid values
	defConfig_ProcessLimit int = 32

	defConfig_DiscordTimeout   int = 180
	defConfig_DownloadTimeout  int = 60
	defConfig_DownloadRetryMax int = 3

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
			Token:    placeholderToken,
			Email:    placeholderEmail,
			Password: placeholderPassword,
		},

		// Owner Settings
		Admins:        []string{},
		AdminChannels: []configurationAdminChannel{},

		// Program Settings
		ProcessLimit:         defConfig_ProcessLimit,
		Debug:                defConfig_Debug,
		SettingsOutput:       true,
		MessageOutput:        true,
		MessageOutputHistory: false,

		DiscordLogLevel:      discordgo.LogError,
		DiscordTimeout:       defConfig_DiscordTimeout,
		DownloadTimeout:      defConfig_DownloadTimeout,
		DownloadRetryMax:     defConfig_DownloadRetryMax,
		ExitOnBadConnection:  false,
		GithubUpdateChecking: defConfig_GithubUpdateChecking,

		CommandPrefix:        defConfig_CommandPrefix,
		ScanOwnMessages:      defConfig_ScanOwnMessages,
		AllowGeneralCommands: true,
		InflateDownloadCount: &defConfig_InflateDownloadCount,
		EuropeanNumbers:      false,

		CheckupRate:         defConfig_CheckupRate,
		ConnectionCheckRate: defConfig_ConnectionCheckRate,
		PresenceRefreshRate: defConfig_PresenceRefreshRate,

		// Source Setup Defaults
		Save:          true,
		AllowCommands: true,
		ScanEdits:     true,
		IgnoreBots:    true,

		SendErrorMessages:  true,
		SendFileToChannel:  "",
		SendFileToChannels: []string{},
		SendFileDirectly:   true,
		SendFileCaption:    "",

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

		// Rules for Saving
		DivideByYear:           false,
		DivideByMonth:          false,
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
		SavePossibleDuplicates: true,
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
			BlockedPhrases: &[]string{},
		},
	}
}

type configuration struct {
	Constants map[string]string `json:"_constants,omitempty"`

	// Logins
	Credentials configurationCredentials `json:"credentials"` // required

	// Owner Settings
	Admins        []string                    `json:"admins"`        // optional
	AdminChannels []configurationAdminChannel `json:"adminChannels"` // optional

	// Program Settings
	ProcessLimit         int  `json:"processLimit,omitempty"`         // optional, defaults
	Debug                bool `json:"debug,omitempty"`                // optional, defaults
	SettingsOutput       bool `json:"settingsOutput,omitempty"`       // optional, defaults
	MessageOutput        bool `json:"messageOutput,omitempty"`        // optional, defaults
	MessageOutputHistory bool `json:"messageOutputHistory,omitempty"` // optional, defaults

	DiscordLogLevel      int  `json:"discordLogLevel,omitempty"`      // optional, defaults
	DiscordTimeout       int  `json:"discordTimeout,omitempty"`       // optional, defaults
	DownloadTimeout      int  `json:"downloadTimeout,omitempty"`      // optional, defaults
	DownloadRetryMax     int  `json:"downloadRetryMax,omitempty"`     // optional, defaults
	ExitOnBadConnection  bool `json:"exitOnBadConnection,omitempty"`  // optional, defaults
	GithubUpdateChecking bool `json:"githubUpdateChecking,omitempty"` // optional, defaults

	CommandPrefix        string `json:"commandPrefix,omitempty"`        // optional, defaults
	ScanOwnMessages      bool   `json:"scanOwnMessages,omitempty"`      // optional, defaults
	AllowGeneralCommands bool   `json:"allowGeneralCommands,omitempty"` // optional, defaults
	InflateDownloadCount *int64 `json:"inflateDownloadCount,omitempty"` // optional, defaults to 0 if undefined
	EuropeanNumbers      bool   `json:"europeanNumbers,omitempty"`      // optional, defaults

	CheckupRate         int `json:"checkupRate,omitempty"`         // optional, defaults
	ConnectionCheckRate int `json:"connectionCheckRate,omitempty"` // optional, defaults
	PresenceRefreshRate int `json:"presenceRefreshRate,omitempty"` // optional, defaults

	// Source Setup Defaults
	Save          bool `json:"save,omitempty"`          // optional, defaults
	AllowCommands bool `json:"allowCommands,omitempty"` // optional, defaults
	ScanEdits     bool `json:"scanEdits,omitempty"`     // optional, defaults
	IgnoreBots    bool `json:"ignoreBots,omitempty"`    // optional, defaults

	SendErrorMessages  bool     `json:"sendErrorMessages,omitempty"`  // optional, defaults
	SendFileToChannel  string   `json:"sendFileToChannel,omitempty"`  // optional, defaults
	SendFileToChannels []string `json:"sendFileToChannels,omitempty"` // optional, defaults
	SendFileDirectly   bool     `json:"sendFileDirectly,omitempty"`   // optional, defaults
	SendFileCaption    string   `json:"sendFileCaption,omitempty"`    // optional

	FilenameDateFormat string `json:"filenameDateFormat,omitempty"` // optional, defaults
	FilenameFormat     string `json:"filenameFormat,omitempty"`     // optional, defaults

	// Appearance
	PresenceEnabled            bool               `json:"presenceEnabled,omitempty"`            // optional, defaults
	PresenceStatus             string             `json:"presenceStatus,omitempty"`             // optional, defaults
	PresenceType               discordgo.GameType `json:"presenceType,omitempty"`               // optional, defaults
	PresenceLabel              *string            `json:"presenceLabel,omitempty"`              // optional, unused if undefined
	PresenceDetails            *string            `json:"presenceDetails,omitempty"`            // optional, unused if undefined
	PresenceState              *string            `json:"presenceState,omitempty"`              // optional, unused if undefined
	ReactWhenDownloaded        bool               `json:"reactWhenDownloaded,omitempty"`        // optional, defaults
	ReactWhenDownloadedEmoji   *string            `json:"reactWhenDownloadedEmoji,omitempty"`   // optional
	ReactWhenDownloadedHistory bool               `json:"reactWhenDownloadedHistory,omitempty"` // optional, defaults
	HistoryTyping              bool               `json:"historyTyping,omitempty"`              // optional, defaults
	EmbedColor                 *string            `json:"embedColor,omitempty"`                 // optional, defaults to role if undefined, then defaults random if no role color

	// History
	HistoryMaxJobs        int    `json:"historyMaxJobs,omitempty"`        // optional, defaults
	AutoHistory           bool   `json:"autoHistory,omitempty"`           // optional, defaults
	AutoHistoryBefore     string `json:"autoHistoryBefore,omitempty"`     // optional
	AutoHistorySince      string `json:"autoHistorySince,omitempty"`      // optional
	SendAutoHistoryStatus bool   `json:"sendAutoHistoryStatus,omitempty"` // optional, defaults
	SendHistoryStatus     bool   `json:"sendHistoryStatus,omitempty"`     // optional, defaults

	// Rules for Saving
	DivideByYear           bool                        `json:"divideByYear,omitempty"`           // defaults
	DivideByMonth          bool                        `json:"divideByMonth,omitempty"`          // defaults
	DivideByServer         bool                        `json:"divideByServer,omitempty"`         // defaults
	DivideByChannel        bool                        `json:"divideByChannel,omitempty"`        // defaults
	DivideByUser           bool                        `json:"divideByUser,omitempty"`           // defaults
	DivideByType           bool                        `json:"divideByType,omitempty"`           // defaults
	DivideFoldersUseID     bool                        `json:"divideFoldersUseID,omitempty"`     // defaults
	SaveImages             bool                        `json:"saveImages,omitempty"`             // defaults
	SaveVideos             bool                        `json:"saveVideos,omitempty"`             // defaults
	SaveAudioFiles         bool                        `json:"saveAudioFiles,omitempty"`         // defaults
	SaveTextFiles          bool                        `json:"saveTextFiles,omitempty"`          // defaults
	SaveOtherFiles         bool                        `json:"saveOtherFiles,omitempty"`         // defaults
	SavePossibleDuplicates bool                        `json:"savePossibleDuplicates,omitempty"` // defaults
	Filters                *configurationSourceFilters `json:"filters,omitempty"`                // optional

	// Sources
	All                    *configurationSource  `json:"all,omitempty"`
	AllBlacklistUsers      *[]string             `json:"allBlacklistUsers,omitempty"`
	AllBlacklistServers    *[]string             `json:"allBlacklistServers,omitempty"`
	AllBlacklistCategories *[]string             `json:"allBlacklistCategories,omitempty"`
	AllBlacklistChannels   *[]string             `json:"allBlacklistChannels,omitempty"`
	Users                  []configurationSource `json:"users,omitempty"`
	Servers                []configurationSource `json:"servers,omitempty"`
	Categories             []configurationSource `json:"categories,omitempty"`
	Channels               []configurationSource `json:"channels,omitempty"`
}

type constStruct struct {
	Constants map[string]string `json:"_constants,omitempty"`
}

//#endregion

//#region Config, Sources

// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	defSource_Enabled bool = true
)

type configurationSource struct {
	// ~
	UserID            string    `json:"user,omitempty"`              // used for config.Users
	UserIDs           *[]string `json:"users,omitempty"`             // ---> alternative to UserID
	ServerID          string    `json:"server,omitempty"`            // used for config.Servers
	ServerIDs         *[]string `json:"servers,omitempty"`           // ---> alternative to ServerID
	ServerBlacklist   *[]string `json:"serverBlacklist,omitempty"`   // for server.ServerID & server.ServerIDs
	CategoryID        string    `json:"category,omitempty"`          // used for config.Categories
	CategoryIDs       *[]string `json:"categories,omitempty"`        // ---> alternative to CategoryID
	CategoryBlacklist *[]string `json:"categoryBlacklist,omitempty"` // for server.CategoryID & server.CategoryIDs
	ChannelID         string    `json:"channel,omitempty"`           // used for config.Channels
	ChannelIDs        *[]string `json:"channels,omitempty"`          // ---> alternative to ChannelID
	Destination       string    `json:"destination"`                 // required

	// Setup
	Enabled       *bool `json:"enabled"`                 // optional, defaults
	Save          *bool `json:"save"`                    // optional, defaults
	AllowCommands *bool `json:"allowCommands,omitempty"` // optional, defaults
	ScanEdits     *bool `json:"scanEdits,omitempty"`     // optional, defaults
	IgnoreBots    *bool `json:"ignoreBots,omitempty"`    // optional, defaults

	SendErrorMessages  *bool     `json:"sendErrorMessages,omitempty"`  // optional, defaults
	SendFileToChannel  *string   `json:"sendFileToChannel,omitempty"`  // optional, defaults
	SendFileToChannels *[]string `json:"sendFileToChannels,omitempty"` // optional, defaults
	SendFileDirectly   *bool     `json:"sendFileDirectly,omitempty"`   // optional, defaults
	SendFileCaption    *string   `json:"sendFileCaption,omitempty"`    // optional

	FilenameDateFormat *string `json:"filenameDateFormat,omitempty"` // optional
	FilenameFormat     *string `json:"filenameFormat,omitempty"`     // optional

	// Appearance
	PresenceEnabled            *bool     `json:"presenceEnabled,omitempty"`            // optional, defaults
	ReactWhenDownloaded        *bool     `json:"reactWhenDownloaded,omitempty"`        // optional, defaults
	ReactWhenDownloadedEmoji   *string   `json:"reactWhenDownloadedEmoji,omitempty"`   // optional, defaults
	ReactWhenDownloadedHistory *bool     `json:"reactWhenDownloadedHistory,omitempty"` // optional, defaults
	BlacklistReactEmojis       *[]string `json:"blacklistReactEmojis,omitempty"`       // optional
	HistoryTyping              *bool     `json:"historyTyping,omitempty"`              // optional, defaults
	EmbedColor                 *string   `json:"embedColor,omitempty"`                 // optional, defaults to role if undefined, then defaults random if no role color

	// History
	AutoHistory           *bool   `json:"autoHistory,omitempty"`           // optional
	AutoHistoryBefore     *string `json:"autoHistoryBefore,omitempty"`     // optional
	AutoHistorySince      *string `json:"autoHistorySince,omitempty"`      // optional
	SendAutoHistoryStatus *bool   `json:"sendAutoHistoryStatus,omitempty"` // optional, defaults
	SendHistoryStatus     *bool   `json:"sendHistoryStatus,omitempty"`     // optional, defaults

	// Rules for Saving
	DivideByYear           *bool                       `json:"divideByYear,omitempty"`           // optional, defaults
	DivideByMonth          *bool                       `json:"divideByMonth,omitempty"`          // optional, defaults
	DivideByServer         *bool                       `json:"divideByServer,omitempty"`         // optional, defaults
	DivideByChannel        *bool                       `json:"divideByChannel,omitempty"`        // optional, defaults
	DivideByUser           *bool                       `json:"divideByUser,omitempty"`           // optional, defaults
	DivideByType           *bool                       `json:"divideByType,omitempty"`           // optional, defaults
	DivideFoldersUseID     *bool                       `json:"divideFoldersUseID,omitempty"`     // optional, defaults
	SaveImages             *bool                       `json:"saveImages,omitempty"`             // optional, defaults
	SaveVideos             *bool                       `json:"saveVideos,omitempty"`             // optional, defaults
	SaveAudioFiles         *bool                       `json:"saveAudioFiles,omitempty"`         // optional, defaults
	SaveTextFiles          *bool                       `json:"saveTextFiles,omitempty"`          // optional, defaults
	SaveOtherFiles         *bool                       `json:"saveOtherFiles,omitempty"`         // optional, defaults
	SavePossibleDuplicates *bool                       `json:"savePossibleDuplicates,omitempty"` // optional, defaults
	Filters                *configurationSourceFilters `json:"filters,omitempty"`                // optional

	// Misc Rules
	LogLinks    *configurationSourceLog `json:"logLinks,omitempty"`    // optional
	LogMessages *configurationSourceLog `json:"logMessages,omitempty"` // optional
}

type configurationSourceFilters struct {
	BlockedPhrases *[]string `json:"blockedPhrases,omitempty"` // optional
	AllowedPhrases *[]string `json:"allowedPhrases,omitempty"` // optional

	BlockedUsers *[]string `json:"blockedUsers,omitempty"` // optional
	AllowedUsers *[]string `json:"allowedUsers,omitempty"` // optional

	BlockedRoles *[]string `json:"blockedRoles,omitempty"` // optional
	AllowedRoles *[]string `json:"allowedRoles,omitempty"` // optional

	BlockedExtensions *[]string `json:"blockedExtensions,omitempty"` // optional
	AllowedExtensions *[]string `json:"allowedExtensions,omitempty"` // optional

	BlockedDomains *[]string `json:"blockedDomains,omitempty"` // optional
	AllowedDomains *[]string `json:"allowedDomains,omitempty"` // optional
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
	Destination         string  `json:"destination"`                   // required
	DestinationIsFolder *bool   `json:"destinationIsFolder,omitempty"` // optional, defaults
	DivideLogsByServer  *bool   `json:"divideLogsByServer,omitempty"`  // optional, defaults
	DivideLogsByChannel *bool   `json:"divideLogsByChannel,omitempty"` // optional, defaults
	DivideLogsByUser    *bool   `json:"divideLogsByUser,omitempty"`    // optional, defaults
	DivideLogsByStatus  *bool   `json:"divideLogsByStatus,omitempty"`  // optional, defaults
	LogDownloads        *bool   `json:"logDownloads,omitempty"`        // optional, defaults
	LogFailures         *bool   `json:"logFailures,omitempty"`         // optional, defaults
	FilterDuplicates    *bool   `json:"filterDuplicates,omitempty"`    // optional, defaults
	Prefix              *string `json:"prefix,omitempty"`              // optional
	Suffix              *string `json:"suffix,omitempty"`              // optional
	UserData            *bool   `json:"userData,omitempty"`            // optional, defaults
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
	ChannelID      string    `json:"channel"`                  // required
	ChannelIDs     *[]string `json:"channels,omitempty"`       // ---> alternative to ChannelID
	LogProgram     *bool     `json:"logProgram,omitempty"`     // optional, defaults
	LogStatus      *bool     `json:"logStatus,omitempty"`      // optional, defaults
	LogErrors      *bool     `json:"logErrors,omitempty"`      // optional, defaults
	UnlockCommands *bool     `json:"unlockCommands,omitempty"` // optional, defaults
}

//#endregion

//#region Management

func loadConfig() error {
	// Determine json type
	if _, err := os.Stat(configFileBase + ".jsonc"); err == nil {
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
		if configFileC {
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
			if configFileC {
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
			if dupeConfig.Credentials.TwitterAccessToken != "" {
				dupeConfig.Credentials.TwitterAccessToken = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.TwitterAccessTokenSecret != "" {
				dupeConfig.Credentials.TwitterAccessTokenSecret = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.InstagramUsername != "" {
				dupeConfig.Credentials.InstagramUsername = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.InstagramPassword != "" {
				dupeConfig.Credentials.InstagramPassword = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.TwitterConsumerKey != "" {
				dupeConfig.Credentials.TwitterConsumerKey = "STRIPPED_FOR_OUTPUT"
			}
			if dupeConfig.Credentials.TwitterConsumerSecret != "" {
				dupeConfig.Credentials.TwitterConsumerSecret = "STRIPPED_FOR_OUTPUT"
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
	}

	return nil
}

func createConfig() {
	log.Println(lg("Settings", "create", color.YellowString, "Creating new settings file..."))

	enteredBaseChannel := "REPLACE_WITH_DISCORD_CHANNEL_ID_TO_DOWNLOAD_FROM"
	enteredBaseDestination := "REPLACE_WITH_FOLDER_LOCATION_TO_DOWNLOAD_TO"

	// Separate from Defaultconfiguration because there's some elements we want to strip for settings creation
	defaultConfig := configuration{
		Credentials: configurationCredentials{
			Token:    placeholderToken,
			Email:    placeholderEmail,
			Password: placeholderPassword,
		},
		Admins:          []string{"REPLACE_WITH_YOUR_DISCORD_USER_ID"},
		CommandPrefix:   defConfig_CommandPrefix,
		ScanOwnMessages: defConfig_ScanOwnMessages,

		PresenceEnabled:     defConfig_PresenceEnabled,
		PresenceStatus:      defConfig_PresenceStatus,
		PresenceType:        defConfig_PresenceType,
		ReactWhenDownloaded: defConfig_ReactWhenDownloaded,

		GithubUpdateChecking: defConfig_GithubUpdateChecking,
		Debug:                defConfig_Debug,
	}

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
			importKey("twitter", "consumer key", &defaultConfig.Credentials.TwitterConsumerKey, "string")
			importKey("twitter", "consumer secret", &defaultConfig.Credentials.TwitterConsumerSecret, "string")
			importKey("twitter", "access token", &defaultConfig.Credentials.TwitterAccessToken, "string")
			importKey("twitter", "access token secret", &defaultConfig.Credentials.TwitterAccessTokenSecret, "string")

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
		baseChannel := configurationSource{
			ChannelID:   enteredBaseChannel,
			Destination: enteredBaseDestination,

			Enabled:           &defSource_Enabled,
			Save:              &config.Save,
			AllowCommands:     &config.AllowCommands,
			SendErrorMessages: &config.SendErrorMessages,
			ScanEdits:         &config.ScanEdits,
			IgnoreBots:        &config.IgnoreBots,

			PresenceEnabled:            &config.PresenceEnabled,
			ReactWhenDownloadedHistory: &config.ReactWhenDownloadedHistory,

			DivideByType: &config.DivideByType,
			SaveImages:   &config.SaveImages,
			SaveVideos:   &config.SaveVideos,
		}
		defaultConfig.Channels = append(defaultConfig.Channels, baseChannel)

		baseAdminChannel := configurationAdminChannel{
			ChannelID: "REPLACE_WITH_DISCORD_CHANNEL_ID_FOR_ADMIN_COMMANDS",
		}
		defaultConfig.AdminChannels = append(defaultConfig.AdminChannels, baseAdminChannel)

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
				"THERE ARE MANY HIDDEN SETTINGS AVAILABLE, SEE THE GITHUB README github.com/"+projectRepo))
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
	if channel.HistoryTyping == nil {
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
	if channel.DivideByYear == nil {
		channel.DivideByYear = &config.DivideByYear
	}
	if channel.DivideByMonth == nil {
		channel.DivideByMonth = &config.DivideByMonth
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

	// Misc Rules
	if channel.LogLinks == nil {
		channel.LogLinks = &configurationSourceLog{}
	}
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

	if channel.LogMessages == nil {
		channel.LogMessages = &configurationSourceLog{}
	}
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

func getSource(m *discordgo.Message) configurationSource {

	// Channel
	for _, item := range config.Channels {
		// Single Channel Config
		if m.ChannelID == item.ChannelID || isNestedMessage(m, item.ChannelID) {
			return item
		}
		// Multi-Channel Config
		if item.ChannelIDs != nil {
			for _, subchannel := range *item.ChannelIDs {
				if m.ChannelID == subchannel || isNestedMessage(m, subchannel) {
					return item
				}
			}
		}
	}

	// Category Config
	for _, item := range config.Categories {
		if item.CategoryID != "" {
			channel, err := bot.State.Channel(m.ChannelID)
			if err == nil {
				if channel.ParentID == item.CategoryID {
					return item
				}
			}
		}
		// Multi-Category Config
		if item.CategoryIDs != nil {
			for _, subcategory := range *item.CategoryIDs {
				channel, err := bot.State.Channel(m.ChannelID)
				if err == nil {
					if channel.ParentID == subcategory {
						if item.CategoryBlacklist != nil {
							if stringInSlice(channel.ParentID, *item.CategoryBlacklist) {
								return emptyConfig
							}
						}
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
			if err == nil {
				for _, channel := range guild.Channels {
					if m.ChannelID == channel.ID || isNestedMessage(m, channel.ID) {
						// Channel Blacklisting within Server
						if item.ServerBlacklist != nil {
							if stringInSlice(m.ChannelID, *item.ServerBlacklist) {
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
				if err == nil {
					for _, channel := range guild.Channels {
						if m.ChannelID == channel.ID || isNestedMessage(m, channel.ID) {
							// Channel Blacklisting within Servers
							if item.ServerBlacklist != nil {
								if stringInSlice(m.ChannelID, *item.ServerBlacklist) {
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
			if stringInSlice(m.ChannelID, *config.AllBlacklistChannels) {
				return emptyConfig
			}
		}
		if config.AllBlacklistCategories != nil {
			chinf, err := bot.State.Channel(m.ChannelID)
			if err == nil {
				if stringInSlice(chinf.ParentID, *config.AllBlacklistCategories) || stringInSlice(m.ChannelID, *config.AllBlacklistCategories) {
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
	} else if channelConfig := getSource(m); channelConfig != emptyConfig {
		if *channelConfig.AllowCommands || isBotAdmin(m) || m.Author.ID == bot.State.User.ID {
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
				if r := getSource(&discordgo.Message{ChannelID: channel.ID}); r == emptyConfig { // easier than redoing it all but way less efficient, im lazy
					continue
				} else {
					if hasPerms(channel.ID, discordgo.PermissionViewChannel) && hasPerms(channel.ID, discordgo.PermissionReadMessageHistory) {
						channels = append(channels, channel.ID)
					}
				}
			}
		}
	} else { // STANDARD MODE
		// Compile all config channels
		for _, channel := range config.Channels {
			if channel.ChannelIDs != nil {
				channels = append(channels, *channel.ChannelIDs...)
			} else if isNumeric(channel.ChannelID) {
				channels = append(channels, channel.ChannelID)
			}
		}
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
	}
	return channels
}

//#endregion
