[![Build Status](https://travis-ci.com/get-got/discord-downloader-go.svg?branch=master)](https://travis-ci.com/get-got/discord-downloader-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/get-got/discord-downloader-go)](https://goreportcard.com/report/github.com/get-got/discord-downloader-go)

# discord-downloader-go

### **This project is a rework of [Seklfreak's _discord-image-downloader-go_](https://github.com/Seklfreak/discord-image-downloader-go)**
#### For list of differences and why I made an independent project, [**see below**](#key-differences-from-seklfreaks-discord-image-downloader-go--why-i-made-this).

## [**DOWNLOAD LATEST RELEASE BUILDS**](https://github.com/get-got/discord-downloader-go/releases/latest)

This is a Discord bot program to download files posted in specified Discord channels to local folders. It can fetch highest possible quality files from various sources (listed below), aside from downloading any file directly linked or attached. _See [Features](#Features) below for full list._

## **WARNING:** Discord does not allow Automated User Accounts (Self-Bots)
[Read more in Discord Trust & Safety Team's Official Statement...](https://support.discordapp.com/hc/en-us/articles/115002192352-Automated-user-accounts-self-bots-)

> _NOTE: This only applies to real User Accounts, not Bot users. This program currently works for either._

## Setup
Edit the `settings.json` file and enter your credentials & configuration. If the file is missing or critically corrupt, it will replace it with a new file. Ensure you follow proper JSON syntax to avoid any unexpected errors.

If using a **Bot User,** enter the token into the `"Token"` setting. Remove the lines for `"Username"` and `"Password"` or leave blank (`""`). **To create a Bot User,** go to [discord.com/developers/applications](https://discord.com/developers/applications) and create a `New Application`. Once created, go to `Bot` and create. The token can be found on the `Bot` page. To invite to your server(s), go to `OAuth2` and check `"bot"`, copy the url, paste into browser and follow prompts for adding to server(s).

If using a **Real User (Self-Bot),** fill out the `"Username"` and `"Password"` settings. Remove the line for `"Token"` or leave blank (`""`).

If using a **Real User (Self-Bot) with 2FA (Two-Factor Authentication),** enter the token into the `"Token"` setting. Remove the lines for `"Username"` and `"Password"` or leave blank (`""`). Token can be found from `Developer Tools` in browser under `localStorage.token` or in the Discord client `Ctrl+Shift+I (Windows)`/`Cmd+Option+I (Mac)` under `Application → Local Storage → https://discordapp.com → "token"`.

* **Discord Developer Mode:** Enable `Developer Mode` in Discord settings under `Appearance`.
* **Finding Channel ID:** _Enable Discord Developer Mode (see above),_ right click on the channel and `Copy ID`.
* **Finding User ID:** _Enable Discord Developer Mode (see above),_ right click on the user and `Copy ID`.
* **Finding Emoji ID:** _Enable Discord Developer Mode (see above),_ right click on the emoji and `Copy ID`.

## Features
* ***Supported File Downloading:***
    * Discord File Attachments
    * Direct Links to Files
    * Twitter _(Images & Videos)_
    * Instagram _(Images & Videos)_
    * Facebook Videos
    * Imgur _(Single Posts & Albums)_
    * Flickr _(requires API key, see config section)_
    * Streamable
    * Gfycat
* ***Commands:***
    * Help _(<prefix>help - Alias: commands)_
    * Ping _(<prefix>ping - Alias: test)_
    * Status: Get an output of the current status of the bot _(<prefix>status - Alias: info)_
    * Stats: Have the bot dump stats _(<prefix>stats)_
    * **[Must be Bot or Server Admin]** History: Process all old messages in channel _(<prefix>history - Aliases: catalog, cache)_
    * **[Must be Bot Admin]** Exit (nice for process managers like pm2 for instant reload) _(<prefix>exit - Aliases: reload, kill)_

### Key Differences from [Seklfreak's _discord-image-downloader-go_](https://github.com/Seklfreak/discord-image-downloader-go) & Why I made this
* _Go 1.15 rather than 1.13_
* _discordgo 0.19.0 rather than 0.16.1_
* Configuration is JSON-based rather than ini to allow more elaborate settings and better organization. With this came many features such as channel-specific settings.
* Somewhat different organization than original project; was initially created from scratch then features ported over.
* Cleaner Logging, Color-Coded Logging
* **Added Support For:** Facebook Videos
* **Removed Support For:** Tistory, Google Drive - These seemed to require some extra configuration and be for rather specific-case use, so I took them out. If anyone can neatly integrate them, feel free to make a pull request.
* Fixed Compatability Issue with `xurls` that required people to edit the project, regarding `xurls.Strict.FindAllString`. Was due to some people having xurls v2 installed while the project required v1.

> I've been a user of Seklfreak's project since ~2018 and it's been great for my uses, but there were certain aspects I wanted to expand upon, one of those being customization of channel configuration, and other features like message reactions upon success, differently formatted statuses, etc. If some aspects are rudimentary or messy, please make a pull request, as this is my first project using Go and I've learned everything from observation & Stack Overflow.

## Settings / Configuration Guide
> I tried to make the config as user friendly as possible, though you still need to follow proper JSON syntax (watch those commas). All options specified below labeled "OPTIONAL" will use default values if they're missing from the settings file.

* **Credentials...** `[key/value object]`
    * **Token** `[string]`
        * _Required for Bot Login or User Login with 2FA, don't include if using User Login without 2FA._
    * **Email** `[string]`
        * _Required for User Login without 2FA, don't include if using Bot Login._
    * **Password** `[string]`
        * _Required for User Login without 2FA, don't include if using Bot Login._
    * _OPTIONAL:_ TwitterAccessToken `[string]`
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * _OPTIONAL:_ TwitterAccessTokenSecret `[string]`
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * _OPTIONAL:_ TwitterConsumerKey `[string]`
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * _OPTIONAL:_ TwitterConsumerSecret `[string]`
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * _OPTIONAL:_ FlickrApiKey `[string]`
        * _Won't use Flickr API for fetching media from posts/albums if credentials are missing._
* _OPTIONAL:_ **Admins** `[array of strings]`
    * Array of User ID strings for users allowed to use admin commands
* _OPTIONAL:_ DownloadRetryMax `[int]`
    * _Default:_ `3`
* _OPTIONAL:_ DownloadTimeout `[int]`
    * _Default:_ `60`
* _OPTIONAL:_ CommandPrefix `[string]`
    * _Default:_ `"ddg "`
* _OPTIONAL:_ AllowSkipping `[bool]`
    * _Default:_ `true`
    * Allow scanning for keywords to skip content downloading.
    * `"skip", "ignore", "don't save", "no save"`
* _OPTIONAL:_ ScanOwnMessages `[bool]`
    * _Default:_ `false`
    * Scans the bots own messages for content to download, only useful if using as a selfbot.
* _OPTIONAL:_ PresenceEnabled `[bool]`
    * _Default:_ `true`
* _OPTIONAL:_ PresenceStatus `[string]`
    * _Default:_ `"idle"`
    * Presence status type.
    * `"online"`, `"idle"`, `"dnd"`, `"invisible"`, `"offline"`
* _OPTIONAL:_ PresenceType `[int]`
    * _Default:_ `0`
    * Presence label type. _(Playing <thing>, Listening to <thing>, etc)_
    * `Game = 0, Streaming = 1, Listening = 2, Watching = 3, Custom = 4`
        * If Bot User, Streaming & Custom won't work properly.
* _OPTIONAL:_ PresenceOverwrite `[string]`
    * _Unused by Default_
    * Replace counter status with custom string.
    * Embedded Placeholders:
        * `{{dgVersion}}`: discord-go version
        * `{{ddgVersion}}`: Project version
        * `{{apiVersion}}`: Discord API version
        * `{{count}}`: Raw total count of downloads
        * `{{countShort}}`: Shortened total count of downloads
        * `{{numGuilds}}`: Number of guilds bot is a member of
        * `{{numChannels}}`: Number of bound channels
        * `{{numAdminChannels}}`: Number of admin channels
        * `{{numAdmins}}`: Number of designated admins
        * `{{timeSavedShort}}`: Last save time formatted as `3:04pm`
        * `{{timeSavedLong}}`: Last save time formatted as `3:04:05pm MST - January 1, 2006`
        * `{{timeSavedShort24}}`: Last save time formatted as `15:04`
        * `{{timeSavedLong24}}`: Last save time formatted as `15:04:05 MST - 1 January, 2006`
        * `{{timeNowShort}}`: Current time formatted as `3:04pm`
        * `{{timeNowLong}}`: Current time formatted as `3:04:05pm MST - January 1, 2006`
        * `{{timeNowShort24}}`: Current time formatted as `15:04`
        * `{{timeNowLong24}}`: Current time formatted as `15:04:05 MST - 1 January, 2006`
        * `{{uptime}}`: Shortened duration of bot uptime
* _OPTIONAL:_ PresenceOverwriteDetails `[string]`
    * _Unused by Default_
    * Replace counter status details with custom string (only works for User, not Bot).
    * Embedded Placeholders:
        * _See `PresenceOverwrite` above..._
* _OPTIONAL:_ PresenceOverwriteState `[string]`
    * _Unused by Default_
    * Replace counter status state with custom string (only works for User, not Bot).
    * Embedded Placeholders:
        * _See `PresenceOverwrite` above..._
* _OPTIONAL:_ FilenameDateFormat `[string]`
    * _Default:_ `"2006-01-02_15-04-05"`
    * See [this Stack Overflow post regarding Golang date formatting.](https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format)
* _OPTIONAL:_ GithubUpdateChecking `[bool]`
    * _Default:_ `true`
    * Check for updates from this repo.
* _OPTIONAL:_ DebugOutput `[bool]`
    * _Default:_ `false`
    * Output debugging information.
* _OPTIONAL:_ EmbedColor `[string]`
    * _Unused by Default_
    * Supports `random`/`rand`, `role`/`user`, or RGB in hex or int format (ex: #FF0000 or 16711680).
* _OPTIONAL:_ InflateCount `[int]`
    * _Unused by Default_
    * Inflates the count of total files downloaded by the bot. I only added this for my own personal use to represent an accurate total amount of files downloaded by previous bots I used.
* **AdminChannels...** `[array of key/value objects]`
    * **ChannelID** `[string]`
* **Channels...** `[array of key/value objects]`
    * **ChannelID** `[string]`
    * **Destination** `[string]`
        * Can be full path or local subfolder.
    * _OPTIONAL:_ Enabled `[bool]`
        * _Default:_ `true`
        * Toggles bot functionality for channel.
    * _OPTIONAL:_ AllowCommands `[bool]`
        * _Default:_ `true`
        * Allow use of commands like ping, help, etc.
    * _OPTIONAL:_ ErrorMessages `[bool]`
        * _Default:_ `true`
        * Send response messages when downloads fail or other download-related errors are encountered.
    * _OPTIONAL:_ ScanEdits `[bool]`
        * _Default:_ `true`
        * Check edits for un-downloaded media.
    * _OPTIONAL:_ UpdatePresence `[bool]`
        * _Default:_ `true`
        * Update Discord Presence when download succeeds within this channel.
    * _OPTIONAL:_ ReactWhenDownloaded `[bool]`
        * _Default:_ `true`
        * Confirmation reaction that file(s) successfully downloaded.
    * _OPTIONAL:_ ReactWhenDownloadedEmoji `[string]`
        * _Unused by Default_
        * Uses specified emoji rather than random server emojis. Simply pasting a standard emoji will work, for custom Discord emojis use "name:ID" format.
    * _OPTIONAL:_ OverwriteFilenameDateFormat `[string]`
        * _Unused by Default_
        * Overwrites the global setting `FilenameDateFormat` _(see above)_
        * See [this Stack Overflow post regarding Golang date formatting.](https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format)
    * _OPTIONAL:_ OverwriteAllowSkipping `[bool]`
        * _Unused by Default_
        * Allow scanning for keywords to skip content downloading.
        * `"skip", "ignore", "don't save", "no save"`
    * _OPTIONAL:_ OverwriteEmbedColor `[string]`
        * _Unused by Default_
        * Supports `random`/`rand`, `role`/`user`, or RGB in hex or int format (ex: #FF0000 or 16711680).
    * _OPTIONAL:_ DivideFoldersByType `[bool]`
        * _Default:_ `true`
        * Separate files into subfolders by type _(e.g. "images", "video", "audio", "text", "other")_
    * _OPTIONAL:_ SaveImages `[bool]`
        * _Default:_ `true`
    * _OPTIONAL:_ SaveVideos `[bool]`
        * _Default:_ `true`
    * _OPTIONAL:_ SaveAudioFiles `[bool]`
        * _Default:_ `false`
    * _OPTIONAL:_ SaveTextFiles `[bool]`
        * _Default:_ `false`
    * _OPTIONAL:_ SaveOtherFiles `[bool]`
        * _Default:_ `false`
    * _OPTIONAL:_ SavePossibleDuplicates `[bool]`
        * _Default:_ `true`
    * _OPTIONAL:_ BlacklistedExtensions `[array of strings]`
        * _Default:_ `[ ".htm", ".html", ".php", ".exe", ".dll", ".bin", ".cmd", ".sh", ".py", ".jar" ]`
        * Ignores files containing specified extensions. Ensure you use proper formatting.