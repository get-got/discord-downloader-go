[![Build Status](https://travis-ci.com/get-got/discord-downloader-go.svg?branch=master)](https://travis-ci.com/get-got/discord-downloader-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/get-got/discord-downloader-go)](https://goreportcard.com/report/github.com/get-got/discord-downloader-go)

# discord-downloader-go

### **This project is a fork of [Seklfreak's _discord-image-downloader-go_](https://github.com/Seklfreak/discord-image-downloader-go)**
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
    * Twitter _(requires API key, see config section)_
    * Instagram
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
* _discordgo 0.22.0 rather than 0.16.1_
* _Implements dgrouter for commands_
* Configuration is JSON-based rather than ini to allow more elaborate settings and better organization. With this came many features such as channel-specific settings.
* Log Formatting, Color-Coded Logging
* Somewhat different organization than original project; initially created from scratch then features ported over.
* **Added Support For:** Facebook Videos
* **Removed Support For:** Tistory, Google Drive - These seemed to require some extra configuration and be for rather specific-case use, so I took them out. If anyone can neatly integrate them, feel free to make a pull request.
* Fixed Compatability Issue with `xurls` that required people to edit the project, regarding `xurls.Strict.FindAllString`. The issue was due to some people having xurls v2 installed while the projects go.mod required v1.1.

> I've been a user of Seklfreak's project since ~2018 and it's been great for my uses, but there were certain aspects I wanted to expand upon, one of those being customization of channel configuration, and other features like message reactions upon success, differently formatted statuses, etc. If some aspects are rudimentary or messy, please make a pull request, as this is my first project using Go and I've learned everything from observation & Stack Overflow.

## History Cataloging Guide
> This guide is to show you how to make the bot go through all old messages in a channel and catalog them as though they were being sent right now, in order to download them all.

You will need the Channel ID (see bottom of [Setup](#Setup)) if attempting to catalog history from a specific channel or group of channels, within an admin channel.

* `<prefix>history` to catalog the current channel the command is sent in (must be registered in `channels` in settings).
* `<prefix>history cancel` to stop cataloging the current channel the command is sent in (must be registered in `channels` in settings).
* `<prefix>history <Channel ID(s)>` to catalog specified channels from within a designated Admin Channel (must be registered in `adminChannels` in settings). You can do multiple channels per command if desired, separated by commas.
* `<prefix>history cancel <Channel ID(s)>` to stop cataloging specified channels from within a designated Admin Channel (must be registered in `adminChannels` in settings). You can do multiple channels per command if desired, separated by commas.

## Settings / Configuration Guide
> I tried to make the configuration as user friendly as possible, though you still need to follow proper JSON syntax (watch those commas). All settings specified below labeled `[DEFAULTS]` will use default values if missing from the settings file, and those labeled `[OPTIONAL]` will not be used if missing from the settings file.

All JSON settings follow camelCase format.

* **credentials** `[key/value object]`
    * **token** `[string]`
        * _Required for Bot Login or User Login with 2FA, don't include if using User Login without 2FA._
    * **email** `[string]`
        * _Required for User Login without 2FA, don't include if using Bot Login._
    * **password** `[string]`
        * _Required for User Login without 2FA, don't include if using Bot Login._
    * _`[OPTIONAL]`_ twitterAccessToken `[string]`
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * _`[OPTIONAL]`_ twitterAccessTokenSecret `[string]`
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * _`[OPTIONAL]`_ twitterConsumerKey `[string]`
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * _`[OPTIONAL]`_ twitterConsumerSecret `[string]`
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * _`[OPTIONAL]`_ flickrApiKey `[string]`
        * _Won't use Flickr API for fetching media from posts/albums if credentials are missing._
* _`[OPTIONAL]`_ **admins** `[array of strings]`
    * Array of User ID strings for users allowed to use admin commands
* _`[DEFAULTS]`_ downloadRetryMax `[int]`
    * _Default:_ `3`
* _`[DEFAULTS]`_ downloadTimeout `[int]`
    * _Default:_ `60`
* _`[DEFAULTS]`_ commandPrefix `[string]`
    * _Default:_ `"ddg "`
* _`[DEFAULTS]`_ allowSkipping `[bool]`
    * _Default:_ `true`
    * Allow scanning for keywords to skip content downloading.
    * `"skip", "ignore", "don't save", "no save"`
* _`[DEFAULTS]`_ scanOwnMessages `[bool]`
    * _Default:_ `false`
    * Scans the bots own messages for content to download, only useful if using as a selfbot.
* _`[DEFAULTS]`_ presenceEnabled `[bool]`
    * _Default:_ `true`
* _`[DEFAULTS]`_ presenceStatus `[string]`
    * _Default:_ `"idle"`
    * Presence status type.
    * `"online"`, `"idle"`, `"dnd"`, `"invisible"`, `"offline"`
* _`[DEFAULTS]`_ presenceType `[int]`
    * _Default:_ `0`
    * Presence label type. _("Playing \<activity\>", "Listening to \<activity\>", etc)_
    * `Game = 0, Streaming = 1, Listening = 2, Watching = 3, Custom = 4`
        * If Bot User, Streaming & Custom won't work properly.
* _`[OPTIONAL]`_ presenceOverwrite `[string]`
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
* _`[OPTIONAL]`_ presenceOverwriteDetails `[string]`
    * _Unused by Default_
    * Replace counter status details with custom string (only works for User, not Bot).
    * Embedded Placeholders:
        * _See `presenceOverwrite` above..._
* _`[OPTIONAL]`_ presenceOverwriteState `[string]`
    * _Unused by Default_
    * Replace counter status state with custom string (only works for User, not Bot).
    * Embedded Placeholders:
        * _See `presenceOverwrite` above..._
* _`[DEFAULTS]`_ filenameDateFormat `[string]`
    * _Default:_ `"2006-01-02_15-04-05"`
    * See [this Stack Overflow post regarding Golang date formatting.](https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format)
* _`[DEFAULTS]`_ githubUpdateChecking `[bool]`
    * _Default:_ `true`
    * Check for updates from this repo.
* _`[DEFAULTS]`_ debugOutput `[bool]`
    * _Default:_ `false`
    * Output debugging information.
* _`[OPTIONAL]`_ embedColor `[string]`
    * _Unused by Default_
    * Supports `random`/`rand`, `role`/`user`, or RGB in hex or int format (ex: #FF0000 or 16711680).
* _`[OPTIONAL]`_ inflateCount `[int]`
    * _Unused by Default_
    * Inflates the count of total files downloaded by the bot. I only added this for my own personal use to represent an accurate total amount of files downloaded by previous bots I used.
* _`[OPTIONAL]`_ **adminChannels** `[array of key/value objects]`
    * **channel** `[string]`
* **channels** `[array of key/value objects]`
    * **channel** `[string]`
        * Channel ID to monitor.
    * **destination** `[string]`
        * Folder path for saving files, can be full path or local subfolder.
    * _`[DEFAULTS]`_ enabled `[bool]`
        * _Default:_ `true`
        * Toggles bot functionality for channel.
    * _`[DEFAULTS]`_ allowCommands `[bool]`
        * _Default:_ `true`
        * Allow use of commands like ping, help, etc.
    * _`[DEFAULTS]`_ errorMessages `[bool]`
        * _Default:_ `true`
        * Send response messages when downloads fail or other download-related errors are encountered.
    * _`[DEFAULTS]`_ scanEdits `[bool]`
        * _Default:_ `true`
        * Check edits for un-downloaded media.
    * _`[DEFAULTS]`_ updatePresence `[bool]`
        * _Default:_ `true`
        * Update Discord Presence when download succeeds within this channel.
    * _`[DEFAULTS]`_ reactWhenDownloaded `[bool]`
        * _Default:_ `true`
        * Confirmation reaction that file(s) successfully downloaded.
    * _`[OPTIONAL]`_ reactWhenDownloadedEmoji `[string]`
        * _Unused by Default_
        * Uses specified emoji rather than random server emojis. Simply pasting a standard emoji will work, for custom Discord emojis use "name:ID" format.
    * _`[OPTIONAL]`_ overwriteFilenameDateFormat `[string]`
        * _Unused by Default_
        * Overwrites the global setting `FilenameDateFormat` _(see above)_
        * See [this Stack Overflow post regarding Golang date formatting.](https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format)
    * _`[OPTIONAL]`_ overwriteAllowSkipping `[bool]`
        * _Unused by Default_
        * Allow scanning for keywords to skip content downloading.
        * `"skip", "ignore", "don't save", "no save"`
    * _`[OPTIONAL]`_ overwriteEmbedColor `[string]`
        * _Unused by Default_
        * Supports `random`/`rand`, `role`/`user`, or RGB in hex or int format (ex: #FF0000 or 16711680).
    * _`[DEFAULTS]`_ divideFoldersByType `[bool]`
        * _Default:_ `true`
        * Separate files into subfolders by type _(e.g. "images", "video", "audio", "text", "other")_
    * _`[DEFAULTS]`_ saveImages `[bool]`
        * _Default:_ `true`
    * _`[DEFAULTS]`_ saveVideos `[bool]`
        * _Default:_ `true`
    * _`[DEFAULTS]`_ saveAudioFiles `[bool]`
        * _Default:_ `false`
    * _`[DEFAULTS]`_ saveTextFiles `[bool]`
        * _Default:_ `false`
    * _`[DEFAULTS]`_ saveOtherFiles `[bool]`
        * _Default:_ `false`
    * _`[DEFAULTS]`_ savePossibleDuplicates `[bool]`
        * _Default:_ `true`
    * _`[DEFAULTS]`_ blacklistedExtensions `[array of strings]`
        * _Default:_ `[ ".htm", ".html", ".php", ".exe", ".dll", ".bin", ".cmd", ".sh", ".py", ".jar" ]`
        * Ignores files containing specified extensions. Ensure you use proper formatting.