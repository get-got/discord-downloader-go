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
)

// `json:",omitempty"` is for settings not to be included into initially written settings file

var (
	PLACEHOLDER_TOKEN    string = "REPLACE_WITH_YOUR_TOKEN_OR_DELETE_LINE"
	PLACEHOLDER_EMAIL    string = "REPLACE_WITH_YOUR_EMAIL_OR_DELETE_LINE"
	PLACEHOLDER_PASSWORD string = "REPLACE_WITH_YOUR_PASSWORD_OR_DELETE_LINE"
)

type configurationCredentials struct {
	// Login
	Token    string `json:"token"`    // required for bot token (this or login)
	Email    string `json:"email"`    // required for login (this or token)
	Password string `json:"password"` // required for login (this or token)
	// APIs
	TwitterAccessToken         string `json:"twitterAccessToken,omitempty"`         // optional
	TwitterAccessTokenSecret   string `json:"twitterAccessTokenSecret,omitempty"`   // optional
	TwitterConsumerKey         string `json:"twitterConsumerKey,omitempty"`         // optional
	TwitterConsumerSecret      string `json:"twitterConsumerSecret,omitempty"`      // optional
	FlickrApiKey               string `json:"flickrApiKey,omitempty"`               // optional
	GoogleDriveCredentialsJSON string `json:"googleDriveCredentialsJSON,omitempty"` // optional
}

// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	// Setup
	SETTING_DEFAULT_DebugOutput          bool   = false
	SETTING_DEFAULT_CommandPrefix        string = "ddg "
	SETTING_DEFAULT_AllowSkipping        bool   = true
	SETTING_DEFAULT_ScanOwnMessages      bool   = false
	SETTING_DEFAULT_DownloadRetryMax     int    = 3
	SETTING_DEFAULT_DownloadTimeout      int    = 60
	SETTING_DEFAULT_GithubUpdateChecking bool   = true
	// Appearance
	SETTING_DEFAULT_PresenceEnabled    bool               = true
	SETTING_DEFAULT_PresenceStatus     string             = string(discordgo.StatusIdle)
	SETTING_DEFAULT_PresenceType       discordgo.GameType = discordgo.GameTypeGame
	SETTING_DEFAULT_FilenameDateFormat string             = "2006-01-02_15-04-05 "
	SETTING_DEFAULT_InflateCount       int64              = 0
)

func DefaultConfiguration() configuration {
	return configuration{
		// Required
		Credentials: configurationCredentials{
			Token:    PLACEHOLDER_TOKEN,
			Email:    PLACEHOLDER_EMAIL,
			Password: PLACEHOLDER_PASSWORD,
		},
		// Setup
		Admins:               []string{},
		AdminChannels:        nil,
		DebugOutput:          SETTING_DEFAULT_DebugOutput,
		CommandPrefix:        SETTING_DEFAULT_CommandPrefix,
		AllowSkipping:        SETTING_DEFAULT_AllowSkipping,
		ScanOwnMessages:      SETTING_DEFAULT_ScanOwnMessages,
		DownloadRetryMax:     SETTING_DEFAULT_DownloadRetryMax,
		DownloadTimeout:      SETTING_DEFAULT_DownloadTimeout,
		GithubUpdateChecking: SETTING_DEFAULT_GithubUpdateChecking,
		// Appearance
		PresenceEnabled:    SETTING_DEFAULT_PresenceEnabled,
		PresenceStatus:     SETTING_DEFAULT_PresenceStatus,
		PresenceType:       SETTING_DEFAULT_PresenceType,
		FilenameDateFormat: SETTING_DEFAULT_FilenameDateFormat,
		InflateCount:       &SETTING_DEFAULT_InflateCount,
		// Channels
		Channels: nil,
	}
}

type configuration struct {
	// Required
	Credentials configurationCredentials `json:"credentials"` // required
	// Setup
	Admins               []string                    `json:"admins"`                     // optional
	AdminChannels        []configurationAdminChannel `json:"adminChannels"`              // optional
	DebugOutput          bool                        `json:"debugOutput"`                // optional, defaults
	CommandPrefix        string                      `json:"commandPrefix"`              // optional, defaults
	AllowSkipping        bool                        `json:"allowSkipping"`              // optional, defaults
	ScanOwnMessages      bool                        `json:"scanOwnMessages"`            // optional, defaults
	DownloadRetryMax     int                         `json:"downloadRetryMax,omitempty"` // optional, defaults
	DownloadTimeout      int                         `json:"downloadTimeout,omitempty"`  // optional, defaults
	GithubUpdateChecking bool                        `json:"githubUpdateChecking"`       // optional, defaults
	// Appearance
	PresenceEnabled          bool               `json:"presenceEnabled"`                    // optional, defaults
	PresenceStatus           string             `json:"presenceStatus"`                     // optional, defaults
	PresenceType             discordgo.GameType `json:"presenceType,omitempty"`             // optional, defaults
	PresenceOverwrite        *string            `json:"presenceOverwrite,omitempty"`        // optional, unused if undefined
	PresenceOverwriteDetails *string            `json:"presenceOverwriteDetails,omitempty"` // optional, unused if undefined
	PresenceOverwriteState   *string            `json:"presenceOverwriteState,omitempty"`   // optional, unused if undefined
	FilenameDateFormat       string             `json:"filenameDateFormat,omitempty"`       // optional, defaults
	EmbedColor               *string            `json:"embedColor,omitempty"`               // optional, defaults to role if undefined, then defaults random if no role color
	InflateCount             *int64             `json:"inflateCount,omitempty"`             // optional, defaults to 0 if undefined
	// Channels
	Channels []configurationChannel `json:"channels"` // required

	/* IDEAS / TODO:

	*

	 */
}

// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	// Setup
	SETTING_DEFAULT_CHANNEL_Enabled       bool = true
	SETTING_DEFAULT_CHANNEL_AllowCommands bool = true
	SETTING_DEFAULT_CHANNEL_ErrorMessages bool = true
	SETTING_DEFAULT_CHANNEL_ScanEdits     bool = true
	// Appearance
	SETTING_DEFAULT_CHANNEL_UpdatePresence           bool     = true
	SETTING_DEFAULT_CHANNEL_ReactWhenDownloaded      bool     = true
	SETTING_DEFAULT_CHANNEL_ReactWhenDownloadedEmoji string   = ""
	SETTING_DEFAULT_CHANNEL_BlacklistReactEmojis     []string = []string{}
	// Rules for Access
	SETTING_DEFAULT_CHANNEL_UsersAllWhitelisted bool = true
	// Rules for Saving
	SETTING_DEFAULT_CHANNEL_DivideFoldersByType    bool = true
	SETTING_DEFAULT_CHANNEL_SaveImages             bool = true
	SETTING_DEFAULT_CHANNEL_SaveVideos             bool = true
	SETTING_DEFAULT_CHANNEL_SaveAudioFiles         bool = false
	SETTING_DEFAULT_CHANNEL_SaveTextFiles          bool = false
	SETTING_DEFAULT_CHANNEL_SaveOtherFiles         bool = false
	SETTING_DEFAULT_CHANNEL_SavePossibleDuplicates bool = true
	SETTING_DEFAULT_CHANNEL_ExtensionBlacklist          = []string{
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
)

type configurationChannel struct {
	// Required
	ChannelID   string    `json:"channel"`            // required
	ChannelIDs  *[]string `json:"channels,omitempty"` // alternative to ChannelID
	Destination string    `json:"destination"`        // required
	// Setup
	Enabled       *bool `json:"enabled,omitempty"`       // optional, defaults
	AllowCommands *bool `json:"allowCommands,omitempty"` // optional, defaults
	ErrorMessages *bool `json:"errorMessages,omitempty"` // optional, defaults
	ScanEdits     *bool `json:"scanEdits,omitempty"`     // optional, defaults
	// Appearance
	UpdatePresence           *bool     `json:"updatePresence,omitempty"`           // optional, defaults
	ReactWhenDownloaded      *bool     `json:"reactWhenDownloaded,omitempty"`      // optional, defaults
	ReactWhenDownloadedEmoji *string   `json:"reactWhenDownloadedEmoji,omitempty"` // optional, defaults
	BlacklistReactEmojis     *[]string `json:"blacklistReactEmojis,omitempty"`     // optional
	// Overwrite Global Settings
	OverwriteFilenameDateFormat *string `json:"overwriteFilenameDateFormat,omitempty"` // optional
	OverwriteAllowSkipping      *bool   `json:"overwriteAllowSkipping,omitempty"`      // optional
	OverwriteEmbedColor         *string `json:"overwriteEmbedColor,omitempty"`         // optional, defaults to role if undefined, then defaults random if no role color
	// Rules for Access
	UsersAllWhitelisted *bool     `json:"usersAllWhitelisted,omitempty"` // optional, defaults to true
	UserWhitelist       *[]string `json:"userWhitelist,omitempty"`       // optional, only relevant if above is false
	UserBlacklist       *[]string `json:"userBlacklist,omitempty"`       // optional
	// Rules for Saving
	DivideFoldersByType    *bool     `json:"divideFoldersByType,omitempty"`    // optional, defaults
	SaveImages             *bool     `json:"saveImages,omitempty"`             // optional, defaults
	SaveVideos             *bool     `json:"saveVideos,omitempty"`             // optional, defaults
	SaveAudioFiles         *bool     `json:"saveAudioFiles,omitempty"`         // optional, defaults
	SaveTextFiles          *bool     `json:"saveTextFiles,omitempty"`          // optional, defaults
	SaveOtherFiles         *bool     `json:"saveOtherFiles,omitempty"`         // optional, defaults
	SavePossibleDuplicates *bool     `json:"savePossibleDuplicates,omitempty"` // optional, defaults
	ExtensionBlacklist     *[]string `json:"extensionBlacklist,omitempty"`     // optional, defaults
	DomainBlacklist        *[]string `json:"domainBlacklist,omitempty"`        // optional, defaults
	SaveAllLinksToFile     *string   `json:"saveAllLinksToFile,omitempty"`     // optional

	/* IDEAS / TODO:

	// These require an efficient way to check roles. I haven't really looked into it.
	* RolesAllWhitelisted *bool     `json:"rolesAllWhitelisted,omitempty"` // optional, defaults to true
	* RoleWhitelist       *[]string `json:"roleWhitelist,omitempty"`       // optional
	* RoleBlacklist       *[]string `json:"roleBlacklist,omitempty"`       // optional

	*/
}

type configurationAdminChannel struct {
	// Required
	ChannelID string `json:"channel"` // required

	/* IDEAS / TODO:

	* UnrestrictAdminCommands *bool `json:"unrestrictAdminCommands,omitempty"` // optional, defaults
	* SendErrorLogs *bool `json:"sendErrorLogs,omitempty"` // optional
	* SendHourlyDigest *bool `json:"sendHourlyDigest,omitempty"` // optional
	* SendDailyDigest *bool `json:"sendDailyDigest,omitempty"` // optional

	 */
}

var (
	config = DefaultConfiguration()
)

func loadConfig() {
	// Load settings
	file, err := os.Open(LOC_CONFIG_FILE)
	defer file.Close()
	if err != nil {
		log.Println(color.HiRedString("Failed to open settings file...\t%s", err))
		createConfig()
		properExit()
	} else {
		decoder := json.NewDecoder(file)
		err := decoder.Decode(&config)
		if err != nil {
			log.Println(color.HiRedString("Settings failed to decode...\t%s", err))
			log.Println(logPrefixHelper, color.MagentaString("Please ensure you're following proper JSON format syntax."))
			properExit()
		}

		// Channel Config Defaults
		// this is dumb but don't see a better way to initialize defaults
		for i := 0; i < len(config.Channels); i++ {
			channelDefault(&config.Channels[i])
		}

		// Debug Output
		if config.DebugOutput {
			s, err := json.MarshalIndent(config, "", "\t")
			if err != nil {
				log.Println(logPrefixDebug, color.HiRedString("Failed to output settings...\t%s", err))
			} else {
				log.Println(logPrefixDebug, color.HiYellowString("Parsed Fixed Settings into JSON:\n\n"),
					color.YellowString(string(s)),
				)
			}
		}

		// Credentials Check
		if (config.Credentials.Token == "" || config.Credentials.Token == PLACEHOLDER_TOKEN) &&
			(config.Credentials.Email == "" || config.Credentials.Email == PLACEHOLDER_EMAIL) &&
			(config.Credentials.Password == "" || config.Credentials.Password == PLACEHOLDER_PASSWORD) {
			log.Println(color.HiRedString("No valid discord login found. Token, Email, and Password are all invalid..."))
			log.Println(color.HiYellowString("Please save your credentials & info into \"%s\" then restart...", LOC_CONFIG_FILE))
			log.Println(logPrefixHelper, color.MagentaString("If your credentials are already properly saved, please ensure you're following proper JSON format syntax."))
			log.Println(logPrefixHelper, color.MagentaString("You DO NOT NEED `Token` *AND* `Email`+`Password`, just one OR the other."))
			properExit()
		}
	}
}

// These have to use the default variables since literal values and consts can't be set to the pointers
func channelDefault(channel *configurationChannel) {
	// Setup
	if channel.Enabled == nil {
		channel.Enabled = &SETTING_DEFAULT_CHANNEL_Enabled
	}
	if channel.AllowCommands == nil {
		channel.AllowCommands = &SETTING_DEFAULT_CHANNEL_AllowCommands
	}
	if channel.ErrorMessages == nil {
		channel.ErrorMessages = &SETTING_DEFAULT_CHANNEL_ErrorMessages
	}
	if channel.ScanEdits == nil {
		channel.ScanEdits = &SETTING_DEFAULT_CHANNEL_ScanEdits
	}
	// Appearance
	if channel.UpdatePresence == nil {
		channel.UpdatePresence = &SETTING_DEFAULT_CHANNEL_UpdatePresence
	}
	if channel.ReactWhenDownloaded == nil {
		channel.ReactWhenDownloaded = &SETTING_DEFAULT_CHANNEL_ReactWhenDownloaded
	}
	if channel.ReactWhenDownloadedEmoji == nil {
		channel.ReactWhenDownloadedEmoji = &SETTING_DEFAULT_CHANNEL_ReactWhenDownloadedEmoji
	}
	if channel.BlacklistReactEmojis == nil {
		channel.BlacklistReactEmojis = &SETTING_DEFAULT_CHANNEL_BlacklistReactEmojis
	}
	// Rules for Access
	if channel.UsersAllWhitelisted == nil {
		channel.UsersAllWhitelisted = &SETTING_DEFAULT_CHANNEL_UsersAllWhitelisted
	}
	// Rules for Saving
	if channel.DivideFoldersByType == nil {
		channel.DivideFoldersByType = &SETTING_DEFAULT_CHANNEL_DivideFoldersByType
	}
	if channel.SaveImages == nil {
		channel.SaveImages = &SETTING_DEFAULT_CHANNEL_SaveImages
	}
	if channel.SaveVideos == nil {
		channel.SaveVideos = &SETTING_DEFAULT_CHANNEL_SaveVideos
	}
	if channel.SaveAudioFiles == nil {
		channel.SaveAudioFiles = &SETTING_DEFAULT_CHANNEL_SaveAudioFiles
	}
	if channel.SaveTextFiles == nil {
		channel.SaveTextFiles = &SETTING_DEFAULT_CHANNEL_SaveTextFiles
	}
	if channel.SaveOtherFiles == nil {
		channel.SaveOtherFiles = &SETTING_DEFAULT_CHANNEL_SaveOtherFiles
	}
	if channel.SavePossibleDuplicates == nil {
		channel.SavePossibleDuplicates = &SETTING_DEFAULT_CHANNEL_SavePossibleDuplicates
	}
	if channel.ExtensionBlacklist == nil {
		channel.ExtensionBlacklist = &SETTING_DEFAULT_CHANNEL_ExtensionBlacklist
	}
}

func createConfig() {
	log.Println(color.YellowString("Creating new settings file..."))

	INPUT_TOKEN := PLACEHOLDER_TOKEN
	INPUT_EMAIL := PLACEHOLDER_EMAIL
	INPUT_PASSWORD := PLACEHOLDER_PASSWORD

	INPUT_ADMIN := "REPLACE_WITH_YOUR_DISCORD_USER_ID"

	INPUT_BASE_CHANNEL_ID := "REPLACE_WITH_DISCORD_CHANNEL_ID_TO_DOWNLOAD_FROM"
	INPUT_BASE_CHANNEL_DEST := "REPLACE_WITH_FOLDER_LOCATION_TO_DOWNLOAD_TO"

	//TODO: Improve, this is very crude, I just wanted *something* for this.
	log.Print(color.HiCyanString("Would you like to enter settings info now? [Y/N]: "))
	reader := bufio.NewReader(os.Stdin)
	read_creds_yn, _ := reader.ReadString('\n')
	read_creds_yn = strings.ReplaceAll(read_creds_yn, "\n", "")
	read_creds_yn = strings.ReplaceAll(read_creds_yn, "\r", "")
	if strings.Contains(strings.ToLower(read_creds_yn), "y") {
	input_creds:
		log.Print(color.HiCyanString("Token or Login? [\"token\"/\"login\"]: "))
		read_creds, _ := reader.ReadString('\n')
		read_creds = strings.ReplaceAll(read_creds, "\n", "")
		read_creds = strings.ReplaceAll(read_creds, "\r", "")
		if strings.Contains(strings.ToLower(read_creds), "token") {
		input_token:
			log.Print(color.HiCyanString("Enter token: "))
			read_token, _ := reader.ReadString('\n')
			read_token = strings.ReplaceAll(read_token, "\n", "")
			read_token = strings.ReplaceAll(read_token, "\r", "")
			if read_token != "" {
				INPUT_TOKEN = read_token
			} else {
				log.Println(color.HiRedString("Please input token..."))
				goto input_token
			}
		} else if strings.Contains(strings.ToLower(read_creds), "login") {
		input_email:
			log.Print(color.HiCyanString("Enter email: "))
			read_email, _ := reader.ReadString('\n')
			read_email = strings.ReplaceAll(read_email, "\n", "")
			read_email = strings.ReplaceAll(read_email, "\r", "")
			if strings.Contains(read_email, "@") {
				INPUT_EMAIL = read_email
			input_password:
				log.Print(color.HiCyanString("Enter password: "))
				read_password, _ := reader.ReadString('\n')
				read_password = strings.ReplaceAll(read_password, "\n", "")
				read_password = strings.ReplaceAll(read_password, "\r", "")
				if read_password != "" {
					INPUT_PASSWORD = read_password
				} else {
					log.Println(color.HiRedString("Please input password..."))
					goto input_password
				}
			} else {
				log.Println(color.HiRedString("Please input email..."))
				goto input_email
			}
		} else {
			log.Println(color.HiRedString("Please input \"token\" or \"login\"..."))
			goto input_creds
		}

	input_admin_id:
		log.Print(color.HiCyanString("Input your Discord User ID: "))
		read_admin, _ := reader.ReadString('\n')
		read_admin = strings.ReplaceAll(read_admin, "\n", "")
		read_admin = strings.ReplaceAll(read_admin, "\r", "")
		if isNumeric(read_admin) {
			INPUT_ADMIN = read_admin
		} else {
			log.Println(color.HiRedString("Please input your Discord User ID..."))
			goto input_admin_id
		}

		//TODO: Base channel setup? Would be kind of annoying and may limit options
		//TODO: Admin channel setup?
	}

	// Separate from Defaultconfiguration because there's some elements we want to strip for settings creation
	defaultConfig := configuration{
		Credentials: configurationCredentials{
			Token:    INPUT_TOKEN,
			Email:    INPUT_EMAIL,
			Password: INPUT_PASSWORD,
		},
		Admins:          []string{INPUT_ADMIN},
		CommandPrefix:   SETTING_DEFAULT_CommandPrefix,
		AllowSkipping:   SETTING_DEFAULT_AllowSkipping,
		ScanOwnMessages: SETTING_DEFAULT_ScanOwnMessages,

		PresenceEnabled: SETTING_DEFAULT_PresenceEnabled,
		PresenceStatus:  SETTING_DEFAULT_PresenceStatus,
		PresenceType:    SETTING_DEFAULT_PresenceType,

		GithubUpdateChecking: SETTING_DEFAULT_GithubUpdateChecking,
		DebugOutput:          SETTING_DEFAULT_DebugOutput,
	}

	baseChannel := configurationChannel{
		ChannelID:   INPUT_BASE_CHANNEL_ID,
		Destination: INPUT_BASE_CHANNEL_DEST,

		Enabled:       &SETTING_DEFAULT_CHANNEL_Enabled,
		AllowCommands: &SETTING_DEFAULT_CHANNEL_AllowCommands,
		ErrorMessages: &SETTING_DEFAULT_CHANNEL_ErrorMessages,
		ScanEdits:     &SETTING_DEFAULT_CHANNEL_ScanEdits,

		UpdatePresence:      &SETTING_DEFAULT_CHANNEL_UpdatePresence,
		ReactWhenDownloaded: &SETTING_DEFAULT_CHANNEL_ReactWhenDownloaded,

		DivideFoldersByType: &SETTING_DEFAULT_CHANNEL_DivideFoldersByType,
		SaveImages:          &SETTING_DEFAULT_CHANNEL_SaveImages,
		SaveVideos:          &SETTING_DEFAULT_CHANNEL_SaveVideos,
	}
	defaultConfig.Channels = append(defaultConfig.Channels, baseChannel)

	baseAdminChannel := configurationAdminChannel{
		ChannelID: "REPLACE_WITH_DISCORD_CHANNEL_ID_FOR_ADMIN_COMMANDS",
	}
	defaultConfig.AdminChannels = append(defaultConfig.AdminChannels, baseAdminChannel)

	log.Println(logPrefixHelper, color.MagentaString("The default settings will be missing some options to avoid clutter."))
	log.Println(logPrefixHelper, color.HiMagentaString("There are MANY MORE SETTINGS! If you would like to maximize customization, see the GitHub README for all available settings."))

	defaultJSON, err := json.MarshalIndent(defaultConfig, "", "\t")
	if err != nil {
		log.Println(color.HiRedString("Failed to format new settings...\t%s", err))
	} else {
		err := ioutil.WriteFile(LOC_CONFIG_FILE, defaultJSON, 0644)
		if err != nil {
			log.Println(color.HiRedString("Failed to save new settings file...\t%s", err))
		} else {
			log.Println(color.HiYellowString("Created new settings file..."))
			log.Println(color.HiYellowString("Please save your credentials & info into \"%s\" then restart...", LOC_CONFIG_FILE))
			log.Println(logPrefixHelper, color.MagentaString("You DO NOT NEED `Token` *AND* `Email`+`Password`, just one OR the other."))
			log.Println(logPrefixHelper, color.MagentaString("See README on GitHub for help and more info..."))
		}
	}
}

func isChannelRegistered(ChannelID string) bool {
	for _, item := range config.Channels {
		// Single Channel Config
		if ChannelID == item.ChannelID {
			return true
		}
		// Multi-Channel Config
		if item.ChannelIDs != nil {
			for _, subchannel := range *item.ChannelIDs {
				if ChannelID == subchannel {
					return true
				}
			}
		}
	}
	return false
}

func getChannelConfig(ChannelID string) configurationChannel {
	for _, item := range config.Channels {
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
	return configurationChannel{}
}

func isAdminChannelRegistered(ChannelID string) bool {
	if config.AdminChannels != nil {
		for _, item := range config.AdminChannels {
			if ChannelID == item.ChannelID {
				return true
			}
		}
	}
	return false
}

func getAdminChannelConfig(ChannelID string) configurationAdminChannel {
	if config.AdminChannels != nil {
		for _, item := range config.AdminChannels {
			if ChannelID == item.ChannelID {
				return item
			}
		}
	}
	return configurationAdminChannel{}
}

func isCommandableChannel(m *discordgo.Message) bool {
	if isAdminChannelRegistered(m.ChannelID) {
		return true
	} else if isChannelRegistered(m.ChannelID) {
		channelConfig := getChannelConfig(m.ChannelID)
		if *channelConfig.AllowCommands || isBotAdmin(m) {
			return true
		}
	}
	return false
}

func getBoundChannelsCount() int {
	var channels []string
	for _, item := range config.Channels {
		if item.ChannelID != "" {
			if !stringInSlice(item.ChannelID, channels) {
				channels = append(channels, item.ChannelID)
			}
		} else if *item.ChannelIDs != nil {
			for _, subchannel := range *item.ChannelIDs {
				if !stringInSlice(subchannel, channels) {
					channels = append(channels, subchannel)
				}
			}
		}
	}
	return len(channels)
}
