package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/muhammadmuzzammil1998/jsonc"
	"gopkg.in/ini.v1"
)

var (
	config = defaultConfiguration()
)

//#region Credentials

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
	TwitterAccessToken         string `json:"twitterAccessToken,omitempty"`         // optional
	TwitterAccessTokenSecret   string `json:"twitterAccessTokenSecret,omitempty"`   // optional
	TwitterConsumerKey         string `json:"twitterConsumerKey,omitempty"`         // optional
	TwitterConsumerSecret      string `json:"twitterConsumerSecret,omitempty"`      // optional
	FlickrApiKey               string `json:"flickrApiKey,omitempty"`               // optional
	GoogleDriveCredentialsJSON string `json:"googleDriveCredentialsJSON,omitempty"` // optional
}

//#endregion

//#region Configuration

// defConfig_ = Config Default
// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	// Setup
	defConfig_DebugOutput          bool   = false
	defConfig_MessageOutput        bool   = true
	defConfig_CommandPrefix        string = "ddg "
	defConfig_ScanOwnMessages      bool   = false
	defConfig_AllowGlobalCommands  bool   = true
	defConfig_GithubUpdateChecking bool   = true
	// Appearance
	defConfig_PresenceEnabled            bool               = true
	defConfig_PresenceStatus             string             = string(discordgo.StatusIdle)
	defConfig_PresenceType               discordgo.GameType = discordgo.GameTypeGame
	defConfig_ReactWhenDownloaded        bool               = true
	defConfig_ReactWhenDownloadedHistory bool               = false
	defConfig_InflateCount               int64              = 0
)

func defaultConfiguration() configuration {
	return configuration{
		// Required
		Credentials: configurationCredentials{
			Token:    placeholderToken,
			Email:    placeholderEmail,
			Password: placeholderPassword,
		},
		// Setup
		Admins: []string{},

		DiscordLogLevel:     discordgo.LogError,
		DebugOutput:         defConfig_DebugOutput,
		MessageOutput:       defConfig_MessageOutput,
		CheckupRate:         30,
		ConnectionCheckRate: 5,
		PresenceRefreshRate: 3,

		CommandPrefix:       defConfig_CommandPrefix,
		ScanOwnMessages:     defConfig_ScanOwnMessages,
		AllowGlobalCommands: defConfig_AllowGlobalCommands,

		AutorunHistory:           false,
		AutorunHistoryBefore:     "",
		AutorunHistorySince:      "",
		SendHistoryStatus:        true,
		SendAutorunHistoryStatus: false,

		ExitOnBadConnection: false,
		DownloadRetryMax:    3,
		DownloadTimeout:     60,
		DiscordTimeout:      180,

		GithubUpdateChecking:           defConfig_GithubUpdateChecking,
		FilterDuplicateImages:          false,
		FilterDuplicateImagesThreshold: 0,
		// Appearance
		PresenceEnabled:            defConfig_PresenceEnabled,
		PresenceStatus:             defConfig_PresenceStatus,
		PresenceType:               defConfig_PresenceType,
		ReactWhenDownloaded:        defConfig_ReactWhenDownloaded,
		ReactWhenDownloadedHistory: defConfig_ReactWhenDownloadedHistory,
		FilenameDateFormat:         "2006-01-02_15-04-05",
		FilenameFormat:             "{{date}} {{shortID}} {{file}}",
		InflateCount:               &defConfig_InflateCount,
		NumberFormatEuropean:       false,
	}
}

type configuration struct {
	Constants map[string]string `json:"_constants,omitempty"`
	// Required
	Credentials configurationCredentials `json:"credentials"` // required
	// Setup
	Admins        []string                    `json:"admins"`        // optional
	AdminChannels []configurationAdminChannel `json:"adminChannels"` // optional
	// Main
	DiscordLogLevel     int  `json:"discordLogLevel,omitempty"`     // optional, defaults
	DebugOutput         bool `json:"debugOutput"`                   // optional, defaults
	MessageOutput       bool `json:"messageOutput"`                 // optional, defaults
	CheckupRate         int  `json:"checkupRate,omitempty"`         // optional, defaults
	ConnectionCheckRate int  `json:"connectionCheckRate,omitempty"` // optional, defaults
	PresenceRefreshRate int  `json:"presenceRefreshRate,omitempty"` // optional, defaults

	CommandPrefix       string `json:"commandPrefix"`                  // optional, defaults
	ScanOwnMessages     bool   `json:"scanOwnMessages"`                // optional, defaults
	AllowGlobalCommands bool   `json:"allowGlobalCommmands,omitempty"` // optional, defaults

	AutorunHistory           bool   `json:"autorunHistory,omitempty"`           // optional, defaults
	AutorunHistoryBefore     string `json:"autorunHistoryBefore,omitempty"`     // optional
	AutorunHistorySince      string `json:"autorunHistorySince,omitempty"`      // optional
	SendAutorunHistoryStatus bool   `json:"sendAutorunHistoryStatus,omitempty"` // optional, defaults
	SendHistoryStatus        bool   `json:"sendHistoryStatus,omitempty"`        // optional, defaults

	DiscordTimeout      int  `json:"discordTimeout,omitempty"`      // optional, defaults
	ExitOnBadConnection bool `json:"exitOnBadConnection,omitempty"` // optional, defaults

	DownloadTimeout  int `json:"downloadTimeout,omitempty"`  // optional, defaults
	DownloadRetryMax int `json:"downloadRetryMax,omitempty"` // optional, defaults

	GithubUpdateChecking bool `json:"githubUpdateChecking"` // optional, defaults

	FilterDuplicateImages          bool    `json:"filterDuplicateImages,omitempty"`          // optional, defaults
	FilterDuplicateImagesThreshold float64 `json:"filterDuplicateImagesThreshold,omitempty"` // optional, defaults
	// Appearance
	PresenceEnabled            bool               `json:"presenceEnabled"`                      // optional, defaults
	PresenceStatus             string             `json:"presenceStatus"`                       // optional, defaults
	PresenceType               discordgo.GameType `json:"presenceType,omitempty"`               // optional, defaults
	PresenceOverwrite          *string            `json:"presenceOverwrite,omitempty"`          // optional, unused if undefined
	PresenceOverwriteDetails   *string            `json:"presenceOverwriteDetails,omitempty"`   // optional, unused if undefined
	PresenceOverwriteState     *string            `json:"presenceOverwriteState,omitempty"`     // optional, unused if undefined
	ReactWhenDownloaded        bool               `json:"reactWhenDownloaded,omitempty"`        // optional, defaults
	ReactWhenDownloadedHistory bool               `json:"reactWhenDownloadedHistory,omitempty"` // optional, defaults
	FilenameDateFormat         string             `json:"filenameDateFormat,omitempty"`         // optional, defaults
	FilenameFormat             string             `json:"filenameFormat,omitempty"`             // optional, defaults
	EmbedColor                 *string            `json:"embedColor,omitempty"`                 // optional, defaults to role if undefined, then defaults random if no role color
	InflateCount               *int64             `json:"inflateCount,omitempty"`               // optional, defaults to 0 if undefined
	NumberFormatEuropean       bool               `json:"numberFormatEuropean,omitempty"`       // optional, defaults
	// Sources
	All                  *configurationSource  `json:"all,omitempty"`                  // optional, defaults
	AllBlacklistServers  *[]string             `json:"allBlacklistServers,omitempty"`  // optional
	AllBlacklistChannels *[]string             `json:"allBlacklistChannels,omitempty"` // optional
	Users                []configurationSource `json:"users"`                          // required
	Servers              []configurationSource `json:"servers"`                        // required
	Categories           []configurationSource `json:"categories"`                     // required
	Channels             []configurationSource `json:"channels"`                       // required
}

type constStruct struct {
	Constants map[string]string `json:"_constants,omitempty"`
}

//#endregion

//#region Sources

// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	// Setup
	defSource_Enabled           bool = true
	defSource_Save              bool = true
	defSource_AllowCommands     bool = true
	defSource_ScanEdits         bool = true
	defSource_IgnoreBots        bool = false
	defSource_SendErrorMessages bool = true
	defSource_SendFileDirectly  bool = true
	// Appearance
	defSource_UpdatePresence             bool     = true
	defSource_ReactWhenDownloadedEmoji   string   = ""
	defSource_ReactWhenDownloaded        bool     = false
	defSource_ReactWhenDownloadedHistory bool     = false
	defSource_BlacklistReactEmojis       []string = []string{}
	defSource_TypeWhileProcessing        bool     = false
	// Rules for Saving
	defSource_DivideFoldersByYear    bool = false
	defSource_DivideFoldersByMonth   bool = false
	defSource_DivideFoldersByServer  bool = false
	defSource_DivideFoldersByChannel bool = false
	defSource_DivideFoldersByUser    bool = false
	defSource_DivideFoldersByType    bool = true
	defSource_DivideFoldersUseID     bool = false
	defSource_SaveImages             bool = true
	defSource_SaveVideos             bool = true
	defSource_SaveAudioFiles         bool = false
	defSource_SaveTextFiles          bool = false
	defSource_SaveOtherFiles         bool = false
	defSource_SavePossibleDuplicates bool = false
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
	Enabled                           *bool     `json:"enabled"`                                     // optional, defaults
	Save                              *bool     `json:"save"`                                        // optional, defaults
	AllowCommands                     *bool     `json:"allowCommands,omitempty"`                     // optional, defaults
	ScanEdits                         *bool     `json:"scanEdits,omitempty"`                         // optional, defaults
	IgnoreBots                        *bool     `json:"ignoreBots,omitempty"`                        // optional, defaults
	OverwriteAutorunHistory           *bool     `json:"overwriteAutorunHistory,omitempty"`           // optional
	OverwriteAutorunHistoryBefore     *string   `json:"overwriteAutorunHistoryBefore,omitempty"`     // optional
	OverwriteAutorunHistorySince      *string   `json:"overwriteAutorunHistorySince,omitempty"`      // optional
	OverwriteSendHistoryStatus        *bool     `json:"overwriteSendHistoryStatus,omitempty"`        // optional, defaults
	OverwriteSendAutorunHistoryStatus *bool     `json:"overwriteSendAutorunHistoryStatus,omitempty"` // optional, defaults
	SendErrorMessages                 *bool     `json:"sendErrorMessages,omitempty"`                 // optional, defaults
	SendFileToChannel                 *string   `json:"sendFileToChannel,omitempty"`                 // optional, defaults
	SendFileToChannels                *[]string `json:"sendFileToChannels,omitempty"`                // optional, defaults
	SendFileDirectly                  *bool     `json:"sendFileDirectly,omitempty"`                  // optional, defaults
	SendFileCaption                   *string   `json:"sendFileCaption,omitempty"`                   // optional
	// Appearance
	UpdatePresence             *bool     `json:"updatePresence,omitempty"`             // optional, defaults
	ReactWhenDownloaded        *bool     `json:"reactWhenDownloaded,omitempty"`        // optional, defaults
	ReactWhenDownloadedEmoji   *string   `json:"reactWhenDownloadedEmoji,omitempty"`   // optional, defaults
	ReactWhenDownloadedHistory *bool     `json:"reactWhenDownloadedHistory,omitempty"` // optional, defaults
	BlacklistReactEmojis       *[]string `json:"blacklistReactEmojis,omitempty"`       // optional
	TypeWhileProcessing        *bool     `json:"typeWhileProcessing,omitempty"`        // optional, defaults
	// Overwrite Global Settings
	OverwriteFilenameDateFormat *string `json:"overwriteFilenameDateFormat,omitempty"` // optional
	OverwriteFilenameFormat     *string `json:"overwriteFilenameFormat,omitempty"`     // optional
	OverwriteEmbedColor         *string `json:"overwriteEmbedColor,omitempty"`         // optional, defaults to role if undefined, then defaults random if no role color
	// Rules for Saving
	DivideFoldersByYear    *bool `json:"divideFoldersByYear,omitempty"`    // optional, defaults
	DivideFoldersByMonth   *bool `json:"divideFoldersByMonth,omitempty"`   // optional, defaults
	DivideFoldersByServer  *bool `json:"divideFoldersByServer,omitempty"`  // optional, defaults
	DivideFoldersByChannel *bool `json:"divideFoldersByChannel,omitempty"` // optional, defaults
	DivideFoldersByUser    *bool `json:"divideFoldersByUser,omitempty"`    // optional, defaults
	DivideFoldersByType    *bool `json:"divideFoldersByType,omitempty"`    // optional, defaults
	DivideFoldersUseID     *bool `json:"divideFoldersUseID,omitempty"`     // optional, defaults
	SaveImages             *bool `json:"saveImages,omitempty"`             // optional, defaults
	SaveVideos             *bool `json:"saveVideos,omitempty"`             // optional, defaults
	SaveAudioFiles         *bool `json:"saveAudioFiles,omitempty"`         // optional, defaults
	SaveTextFiles          *bool `json:"saveTextFiles,omitempty"`          // optional, defaults
	SaveOtherFiles         *bool `json:"saveOtherFiles,omitempty"`         // optional, defaults
	SavePossibleDuplicates *bool `json:"savePossibleDuplicates,omitempty"` // optional, defaults
	// Misc Rules
	Filters     *configurationSourceFilters `json:"filters,omitempty"`     // optional
	LogLinks    *configurationSourceLog     `json:"logLinks,omitempty"`    // optional
	LogMessages *configurationSourceLog     `json:"logMessages,omitempty"` // optional
}

var (
	defSourceFilter_BlockedExtensions = []string{
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
	}
	defSourceFilter_BlockedPhrases = []string{}
)

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

//#region Admin Channels

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

func initConfig() {
	if _, err := os.Stat(configFileBase + ".jsonc"); err == nil {
		configFile = configFileBase + ".jsonc"
		configFileC = true
	} else {
		configFile = configFileBase + ".json"
		configFileC = false
	}
}

func loadConfig() {
	// Determine json type
	if _, err := os.Stat(configFileBase + ".jsonc"); err == nil {
		configFile = configFileBase + ".jsonc"
		configFileC = true
	} else {
		configFile = configFileBase + ".json"
		configFileC = false
	}
	// .
	log.Println(lg("Settings", "loadConfig", color.YellowString, "Loading from \"%s\"...", configFile))
	// Load settings
	configContent, err := ioutil.ReadFile(configFile)
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

		// Channel Config Defaults
		// this is dumb but don't see a better way to initialize defaults
		for i := 0; i < len(config.Channels); i++ {
			channelDefault(&config.Channels[i])
		}
		for i := 0; i < len(config.Categories); i++ {
			channelDefault(&config.Categories[i])
		}
		for i := 0; i < len(config.Servers); i++ {
			channelDefault(&config.Servers[i])
		}
		for i := 0; i < len(config.Users); i++ {
			channelDefault(&config.Users[i])
		}
		if config.All != nil {
			channelDefault(config.All)
		}

		for i := 0; i < len(config.AdminChannels); i++ {
			adminChannelDefault(&config.AdminChannels[i])
		}

		// Debug Output
		if config.DebugOutput {
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
			log.Println(lg("Settings", "loadConfig", color.HiRedString, "No valid discord login found. Token, Email, and Password are all invalid..."))
			log.Println(lg("Settings", "loadConfig", color.HiYellowString, "Please save your credentials & info into \"%s\" then restart...", configFile))
			log.Println(lg("Settings", "loadConfig", color.MagentaString, "If your credentials are already properly saved, please ensure you're following proper JSON format syntax."))
			log.Println(lg("Settings", "loadConfig", color.MagentaString, "You DO NOT NEED `Token` *AND* `Email`+`Password`, just one OR the other."))
			properExit()
		}
	}
}

func createConfig() {
	log.Println(lg("Settings", "createConfig", color.YellowString, "Creating new settings file..."))

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
		DebugOutput:          defConfig_DebugOutput,
	}

	// Import old config
	if _, err := os.Stat("config.ini"); err == nil {
		log.Println(lg("Settings", "createConfig", color.HiGreenString,
			"Detected config.ini from Seklfreak's discord-image-downloader-go, importing..."))
		cfg, err := ini.Load("config.ini")
		if err != nil {
			log.Println(lg("Settings", "createConfig", color.HiRedString,
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
					log.Println(lg("Settings", "createConfig", color.GreenString, "IMPORTED %s - %s:\t\t\t%s", section, key, outVar))
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
			importKey("google", "client credentials json", &defaultConfig.Credentials.GoogleDriveCredentialsJSON, "string")
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
				log.Println(lg("Settings", "createConfig", color.GreenString, "IMPORTED Admin Channel:\t\t%s", key))
				defaultConfig.AdminChannels = append(defaultConfig.AdminChannels, newChannel)
			}
			ChannelWhitelist := cfg.Section("channels").KeysHash()
			for key, value := range ChannelWhitelist {
				newChannel := configurationSource{
					ChannelID:   key,
					Destination: value,
				}
				log.Println(lg("Settings", "createConfig", color.GreenString, "IMPORTED Channel:\t\t\t%s to \"%s\"", key, value))
				defaultConfig.Channels = append(defaultConfig.Channels, newChannel)
			}
		}
		log.Println(lg("Settings", "createConfig", color.HiGreenString,
			"Finished importing config.ini from Seklfreak's discord-image-downloader-go!"))
	} else {
		baseChannel := configurationSource{
			ChannelID:   enteredBaseChannel,
			Destination: enteredBaseDestination,

			Enabled:           &defSource_Enabled,
			Save:              &defSource_Save,
			AllowCommands:     &defSource_AllowCommands,
			SendErrorMessages: &defSource_SendErrorMessages,
			ScanEdits:         &defSource_ScanEdits,
			IgnoreBots:        &defSource_IgnoreBots,

			UpdatePresence:             &defSource_UpdatePresence,
			ReactWhenDownloadedEmoji:   &defSource_ReactWhenDownloadedEmoji,
			ReactWhenDownloadedHistory: &defSource_ReactWhenDownloadedHistory,

			DivideFoldersByType: &defSource_DivideFoldersByType,
			SaveImages:          &defSource_SaveImages,
			SaveVideos:          &defSource_SaveVideos,
		}
		defaultConfig.Channels = append(defaultConfig.Channels, baseChannel)

		baseAdminChannel := configurationAdminChannel{
			ChannelID: "REPLACE_WITH_DISCORD_CHANNEL_ID_FOR_ADMIN_COMMANDS",
		}
		defaultConfig.AdminChannels = append(defaultConfig.AdminChannels, baseAdminChannel)

		//TODO: Improve, this is very crude, I just wanted *something* for this.
		log.Print(lg("Settings", "createConfig", color.HiCyanString, "Would you like to enter settings info now? [Y/N]: "))
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
					log.Println(lg("Settings", "createConfig", color.HiRedString, "Please input token..."))
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
						log.Println(lg("Settings", "createConfig", color.HiRedString, "Please input password..."))
						goto EnterPassword
					}
				} else {
					log.Println(lg("Settings", "createConfig", color.HiRedString, "Please input email..."))
					goto EnterEmail
				}
			} else {
				log.Println(lg("Settings", "createConfig", color.HiRedString, "Please input \"token\" or \"login\"..."))
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
				log.Println(lg("Settings", "createConfig", color.HiRedString, "Please input your Discord User ID..."))
				goto EnterAdmin
			}

			//TODO: Base channel setup? Would be kind of annoying and may limit options
			//TODO: Admin channel setup?
		}
	}

	log.Println(lg("Settings", "createConfig", color.MagentaString,
		"The default settings will be missing some options to avoid clutter."))
	log.Println(lg("Settings", "createConfig", color.HiMagentaString,
		"There are MANY MORE SETTINGS! If you would like to maximize customization, see the GitHub README for all available settings."))

	defaultJSON, err := json.MarshalIndent(defaultConfig, "", "\t")
	if err != nil {
		log.Println(lg("Settings", "createConfig", color.HiRedString, "Failed to format new settings...\t%s", err))
	} else {
		err := ioutil.WriteFile(configFile, defaultJSON, 0644)
		if err != nil {
			log.Println(lg("Settings", "createConfig", color.HiRedString, "Failed to save new settings file...\t%s", err))
		} else {
			log.Println(lg("Settings", "createConfig", color.HiYellowString, "Created new settings file..."))
			log.Println(lg("Settings", "createConfig", color.HiYellowString,
				"Please save your credentials & info into \"%s\" then restart...", configFile))
			log.Println(lg("Settings", "createConfig", color.MagentaString,
				"You DO NOT NEED `Token` *AND* `Email`+`Password`, just one OR the other."))
			log.Println(lg("Settings", "createConfig", color.MagentaString,
				"See README on GitHub for help and more info..."))
		}
	}
}

func channelDefault(channel *configurationSource) {
	// These have to use the default variables since literal values and consts can't be set to the pointers

	// Setup
	if channel.Enabled == nil {
		channel.Enabled = &defSource_Enabled
	}
	if channel.Save == nil {
		channel.Save = &defSource_Save
	}
	if channel.AllowCommands == nil {
		channel.AllowCommands = &defSource_AllowCommands
	}
	if channel.SendErrorMessages == nil {
		channel.SendErrorMessages = &defSource_SendErrorMessages
	}
	if channel.ScanEdits == nil {
		channel.ScanEdits = &defSource_ScanEdits
	}
	if channel.IgnoreBots == nil {
		channel.IgnoreBots = &defSource_IgnoreBots
	}
	if channel.SendFileDirectly == nil {
		channel.SendFileDirectly = &defSource_SendFileDirectly
	}
	// Appearance
	if channel.UpdatePresence == nil {
		channel.UpdatePresence = &defSource_UpdatePresence
	}
	if channel.ReactWhenDownloadedEmoji == nil {
		channel.ReactWhenDownloadedEmoji = &defSource_ReactWhenDownloadedEmoji
	}
	if channel.ReactWhenDownloadedHistory == nil {
		channel.ReactWhenDownloadedHistory = &defSource_ReactWhenDownloadedHistory
	}
	if channel.BlacklistReactEmojis == nil {
		channel.BlacklistReactEmojis = &defSource_BlacklistReactEmojis
	}
	if channel.TypeWhileProcessing == nil {
		channel.TypeWhileProcessing = &defSource_TypeWhileProcessing
	}
	// Rules for Saving
	if channel.DivideFoldersByYear == nil {
		channel.DivideFoldersByYear = &defSource_DivideFoldersByYear
	}
	if channel.DivideFoldersByMonth == nil {
		channel.DivideFoldersByMonth = &defSource_DivideFoldersByMonth
	}
	if channel.DivideFoldersByServer == nil {
		channel.DivideFoldersByServer = &defSource_DivideFoldersByServer
	}
	if channel.DivideFoldersByChannel == nil {
		channel.DivideFoldersByChannel = &defSource_DivideFoldersByChannel
	}
	if channel.DivideFoldersByUser == nil {
		channel.DivideFoldersByUser = &defSource_DivideFoldersByUser
	}
	if channel.DivideFoldersByType == nil {
		channel.DivideFoldersByType = &defSource_DivideFoldersByType
	}
	if channel.DivideFoldersUseID == nil {
		channel.DivideFoldersUseID = &defSource_DivideFoldersUseID
	}
	if channel.SaveImages == nil {
		channel.SaveImages = &defSource_SaveImages
	}
	if channel.SaveVideos == nil {
		channel.SaveVideos = &defSource_SaveVideos
	}
	if channel.SaveAudioFiles == nil {
		channel.SaveAudioFiles = &defSource_SaveAudioFiles
	}
	if channel.SaveTextFiles == nil {
		channel.SaveTextFiles = &defSource_SaveTextFiles
	}
	if channel.SaveOtherFiles == nil {
		channel.SaveOtherFiles = &defSource_SaveOtherFiles
	}
	if channel.SavePossibleDuplicates == nil {
		channel.SavePossibleDuplicates = &defSource_SavePossibleDuplicates
	}

	if channel.Filters == nil {
		channel.Filters = &configurationSourceFilters{}
	}
	if channel.Filters.BlockedExtensions == nil {
		channel.Filters.BlockedExtensions = &defSourceFilter_BlockedExtensions
	}
	if channel.Filters.BlockedPhrases == nil {
		channel.Filters.BlockedPhrases = &defSourceFilter_BlockedPhrases
	}

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

//#region Channel Checks/Returns

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
		if config.AllBlacklistServers != nil {
			if stringInSlice(m.GuildID, *config.AllBlacklistServers) {
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
	if config.AllowGlobalCommands {
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
				if config.AllBlacklistChannels != nil {
					if stringInSlice(channel.ID, *config.AllBlacklistChannels) {
						continue
					}
				}
				if hasPerms(channel.ID, discordgo.PermissionReadMessages) && hasPerms(channel.ID, discordgo.PermissionReadMessageHistory) {
					channels = append(channels, channel.ID)
				}
			}
		}
	} else { // STANDARD MODE
		// Compile all config channels
		for _, channel := range config.Channels {
			if channel.ChannelIDs != nil {
				for _, subchannel := range *channel.ChannelIDs {
					channels = append(channels, subchannel)
				}
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
