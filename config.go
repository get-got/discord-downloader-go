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
	placeholderToken    string = "REPLACE_WITH_YOUR_TOKEN_OR_DELETE_LINE"
	placeholderEmail    string = "REPLACE_WITH_YOUR_EMAIL_OR_DELETE_LINE"
	placeholderPassword string = "REPLACE_WITH_YOUR_PASSWORD_OR_DELETE_LINE"
)

type configurationCredentials struct {
	// Login
	Token    string `json:"token"`    // required for bot token (this or login)
	Email    string `json:"email"`    // required for login (this or token)
	Password string `json:"password"` // required for login (this or token)
	UserBot  bool   `json:"userBot"`  // required
	// APIs
	TwitterAccessToken         string `json:"twitterAccessToken,omitempty"`         // optional
	TwitterAccessTokenSecret   string `json:"twitterAccessTokenSecret,omitempty"`   // optional
	TwitterConsumerKey         string `json:"twitterConsumerKey,omitempty"`         // optional
	TwitterConsumerSecret      string `json:"twitterConsumerSecret,omitempty"`      // optional
	FlickrApiKey               string `json:"flickrApiKey,omitempty"`               // optional
	GoogleDriveCredentialsJSON string `json:"googleDriveCredentialsJSON,omitempty"` // optional
}

// cd = Config Default
// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	// Setup
	cdDebugOutput          bool   = false
	cdCommandPrefix        string = "ddg "
	cdAllowSkipping        bool   = true
	cdScanOwnMessages      bool   = false
	cdGithubUpdateChecking bool   = true
	// Appearance
	cdPresenceEnabled bool               = true
	cdPresenceStatus  string             = string(discordgo.StatusIdle)
	cdPresenceType    discordgo.GameType = discordgo.GameTypeGame
	cdInflateCount    int64              = 0
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
		Admins:                         []string{},
		DebugOutput:                    cdDebugOutput,
		CommandPrefix:                  cdCommandPrefix,
		AllowSkipping:                  cdAllowSkipping,
		ScanOwnMessages:                cdScanOwnMessages,
		FilterDuplicateImages:          false,
		FilterDuplicateImagesThreshold: 25,
		DownloadRetryMax:               3,
		DownloadTimeout:                60,
		GithubUpdateChecking:           cdGithubUpdateChecking,
		// Appearance
		PresenceEnabled:      cdPresenceEnabled,
		PresenceStatus:       cdPresenceStatus,
		PresenceType:         cdPresenceType,
		FilenameDateFormat:   "2006-01-02_15-04-05 ",
		InflateCount:         &cdInflateCount,
		NumberFormatEuropean: false,
	}
}

type configuration struct {
	// Required
	Credentials configurationCredentials `json:"credentials"` // required
	// Setup
	Admins                         []string                    `json:"admins"`                                   // optional
	AdminChannels                  []configurationAdminChannel `json:"adminChannels"`                            // optional
	DebugOutput                    bool                        `json:"debugOutput"`                              // optional, defaults
	CommandPrefix                  string                      `json:"commandPrefix"`                            // optional, defaults
	AllowSkipping                  bool                        `json:"allowSkipping"`                            // optional, defaults
	ScanOwnMessages                bool                        `json:"scanOwnMessages"`                          // optional, defaults
	FilterDuplicateImages          bool                        `json:"filterDuplicateImages,omitempty"`          // optional, defaults
	FilterDuplicateImagesThreshold float64                     `json:"filterDuplicateImagesThreshold,omitempty"` // optional, defaults
	DownloadRetryMax               int                         `json:"downloadRetryMax,omitempty"`               // optional, defaults
	DownloadTimeout                int                         `json:"downloadTimeout,omitempty"`                // optional, defaults
	GithubUpdateChecking           bool                        `json:"githubUpdateChecking"`                     // optional, defaults
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
	NumberFormatEuropean     bool               `json:"numberFormatEuropean,omitempty"`     // optional, defaults
	// Channels
	Channels []configurationChannel `json:"channels"` // required

	/* IDEAS / TODO:

	*

	 */
}

// ccd = Channel Config Default
// Needed for settings used without redundant nil checks, and settings defaulting + creation
var (
	// Setup
	ccdEnabled       bool = true
	ccdAllowCommands bool = true
	ccdErrorMessages bool = true
	ccdScanEdits     bool = true
	// Appearance
	ccdUpdatePresence           bool     = true
	ccdReactWhenDownloaded      bool     = true
	ccdReactWhenDownloadedEmoji string   = ""
	ccdBlacklistReactEmojis     []string = []string{}
	// Rules for Access
	ccdUsersAllWhitelisted bool = true
	// Rules for Saving
	ccdDivideFoldersByServer  bool = false
	ccdDivideFoldersByChannel bool = false
	ccdDivideFoldersByUser    bool = false
	ccdDivideFoldersByType    bool = true
	ccdSaveImages             bool = true
	ccdSaveVideos             bool = true
	ccdSaveAudioFiles         bool = false
	ccdSaveTextFiles          bool = false
	ccdSaveOtherFiles         bool = false
	ccdSavePossibleDuplicates bool = false
	ccdExtensionBlacklist          = []string{
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
	// Main
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
	DivideFoldersByServer  *bool     `json:"divideFoldersByServer,omitempty"`  // optional, defaults
	DivideFoldersByChannel *bool     `json:"divideFoldersByChannel,omitempty"` // optional, defaults
	DivideFoldersByUser    *bool     `json:"divideFoldersByUser,omitempty"`    // optional, defaults
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
	config = defaultConfiguration()
)

func loadConfig() {
	// Load settings
	file, err := os.Open(configPath)
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
		if (config.Credentials.Token == "" || config.Credentials.Token == placeholderToken) &&
			(config.Credentials.Email == "" || config.Credentials.Email == placeholderEmail) &&
			(config.Credentials.Password == "" || config.Credentials.Password == placeholderPassword) {
			log.Println(color.HiRedString("No valid discord login found. Token, Email, and Password are all invalid..."))
			log.Println(color.HiYellowString("Please save your credentials & info into \"%s\" then restart...", configPath))
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
		channel.Enabled = &ccdEnabled
	}
	if channel.AllowCommands == nil {
		channel.AllowCommands = &ccdAllowCommands
	}
	if channel.ErrorMessages == nil {
		channel.ErrorMessages = &ccdErrorMessages
	}
	if channel.ScanEdits == nil {
		channel.ScanEdits = &ccdScanEdits
	}
	// Appearance
	if channel.UpdatePresence == nil {
		channel.UpdatePresence = &ccdUpdatePresence
	}
	if channel.ReactWhenDownloaded == nil {
		channel.ReactWhenDownloaded = &ccdReactWhenDownloaded
	}
	if channel.ReactWhenDownloadedEmoji == nil {
		channel.ReactWhenDownloadedEmoji = &ccdReactWhenDownloadedEmoji
	}
	if channel.BlacklistReactEmojis == nil {
		channel.BlacklistReactEmojis = &ccdBlacklistReactEmojis
	}
	// Rules for Access
	if channel.UsersAllWhitelisted == nil {
		channel.UsersAllWhitelisted = &ccdUsersAllWhitelisted
	}
	// Rules for Saving
	if channel.DivideFoldersByServer == nil {
		channel.DivideFoldersByServer = &ccdDivideFoldersByServer
	}
	if channel.DivideFoldersByChannel == nil {
		channel.DivideFoldersByChannel = &ccdDivideFoldersByChannel
	}
	if channel.DivideFoldersByUser == nil {
		channel.DivideFoldersByUser = &ccdDivideFoldersByUser
	}
	if channel.DivideFoldersByType == nil {
		channel.DivideFoldersByType = &ccdDivideFoldersByType
	}
	if channel.SaveImages == nil {
		channel.SaveImages = &ccdSaveImages
	}
	if channel.SaveVideos == nil {
		channel.SaveVideos = &ccdSaveVideos
	}
	if channel.SaveAudioFiles == nil {
		channel.SaveAudioFiles = &ccdSaveAudioFiles
	}
	if channel.SaveTextFiles == nil {
		channel.SaveTextFiles = &ccdSaveTextFiles
	}
	if channel.SaveOtherFiles == nil {
		channel.SaveOtherFiles = &ccdSaveOtherFiles
	}
	if channel.SavePossibleDuplicates == nil {
		channel.SavePossibleDuplicates = &ccdSavePossibleDuplicates
	}
	if channel.ExtensionBlacklist == nil {
		channel.ExtensionBlacklist = &ccdExtensionBlacklist
	}
}

func createConfig() {
	log.Println(color.YellowString("Creating new settings file..."))

	enteredToken := placeholderToken
	enteredEmail := placeholderEmail
	enteredPassword := placeholderPassword

	enteredAdmin := "REPLACE_WITH_YOUR_DISCORD_USER_ID"

	enteredBaseChannel := "REPLACE_WITH_DISCORD_CHANNEL_ID_TO_DOWNLOAD_FROM"
	enteredBaseDestination := "REPLACE_WITH_FOLDER_LOCATION_TO_DOWNLOAD_TO"

	//TODO: Improve, this is very crude, I just wanted *something* for this.
	log.Print(color.HiCyanString("Would you like to enter settings info now? [Y/N]: "))
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
				enteredToken = inputToken
			} else {
				log.Println(color.HiRedString("Please input token..."))
				goto EnterToken
			}
		} else if strings.Contains(strings.ToLower(inputCreds), "login") {
		EnterEmail:
			log.Print(color.HiCyanString("Enter email: "))
			inputEmail, _ := reader.ReadString('\n')
			inputEmail = strings.ReplaceAll(inputEmail, "\n", "")
			inputEmail = strings.ReplaceAll(inputEmail, "\r", "")
			if strings.Contains(inputEmail, "@") {
				enteredEmail = inputEmail
			EnterPassword:
				log.Print(color.HiCyanString("Enter password: "))
				inputPassword, _ := reader.ReadString('\n')
				inputPassword = strings.ReplaceAll(inputPassword, "\n", "")
				inputPassword = strings.ReplaceAll(inputPassword, "\r", "")
				if inputPassword != "" {
					enteredPassword = inputPassword
				} else {
					log.Println(color.HiRedString("Please input password..."))
					goto EnterPassword
				}
			} else {
				log.Println(color.HiRedString("Please input email..."))
				goto EnterEmail
			}
		} else {
			log.Println(color.HiRedString("Please input \"token\" or \"login\"..."))
			goto EnterCreds
		}

	EnterAdmin:
		log.Print(color.HiCyanString("Input your Discord User ID: "))
		inputAdmin, _ := reader.ReadString('\n')
		inputAdmin = strings.ReplaceAll(inputAdmin, "\n", "")
		inputAdmin = strings.ReplaceAll(inputAdmin, "\r", "")
		if isNumeric(inputAdmin) {
			enteredAdmin = inputAdmin
		} else {
			log.Println(color.HiRedString("Please input your Discord User ID..."))
			goto EnterAdmin
		}

		//TODO: Base channel setup? Would be kind of annoying and may limit options
		//TODO: Admin channel setup?
	}

	// Separate from Defaultconfiguration because there's some elements we want to strip for settings creation
	defaultConfig := configuration{
		Credentials: configurationCredentials{
			Token:    enteredToken,
			Email:    enteredEmail,
			Password: enteredPassword,
		},
		Admins:          []string{enteredAdmin},
		CommandPrefix:   cdCommandPrefix,
		AllowSkipping:   cdAllowSkipping,
		ScanOwnMessages: cdScanOwnMessages,

		PresenceEnabled: cdPresenceEnabled,
		PresenceStatus:  cdPresenceStatus,
		PresenceType:    cdPresenceType,

		GithubUpdateChecking: cdGithubUpdateChecking,
		DebugOutput:          cdDebugOutput,
	}

	baseChannel := configurationChannel{
		ChannelID:   enteredBaseChannel,
		Destination: enteredBaseDestination,

		Enabled:       &ccdEnabled,
		AllowCommands: &ccdAllowCommands,
		ErrorMessages: &ccdErrorMessages,
		ScanEdits:     &ccdScanEdits,

		UpdatePresence:      &ccdUpdatePresence,
		ReactWhenDownloaded: &ccdReactWhenDownloaded,

		DivideFoldersByType: &ccdDivideFoldersByType,
		SaveImages:          &ccdSaveImages,
		SaveVideos:          &ccdSaveVideos,
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
		err := ioutil.WriteFile(configPath, defaultJSON, 0644)
		if err != nil {
			log.Println(color.HiRedString("Failed to save new settings file...\t%s", err))
		} else {
			log.Println(color.HiYellowString("Created new settings file..."))
			log.Println(color.HiYellowString("Please save your credentials & info into \"%s\" then restart...", configPath))
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
