<h1 align="center">
    Discord Downloader 💾
</h1>
<p align="center">
    <a href="https://travis-ci.com/get-got/discord-downloader-go" alt="Build Status">
        <img src="https://travis-ci.com/get-got/discord-downloader-go.svg?branch=master" />
    </a>
    <a href="https://goreportcard.com/report/github.com/get-got/discord-downloader-go" alt="Go Report Card">
        <img src="https://goreportcard.com/badge/github.com/get-got/discord-downloader-go" />
    </a>
    <a href="" alt="Go Version">
        <img src="https://img.shields.io/github/go-mod/go-version/get-got/discord-downloader-go" />
    </a>
    <br>
    <a href="https://github.com/get-got/discord-downloader-go/releases" alt="All Releases">
        <img src="https://img.shields.io/github/downloads/get-got/discord-downloader-go/total?label=all-releases&logo=GitHub" />
    </a>
    <a href="https://github.com/get-got/discord-downloader-go/releases/latest" alt="Latest Release">
        <img src="https://img.shields.io/github/downloads/get-got/discord-downloader-go/latest/total?label=latest-release&logo=GitHub" />
    </a>
    <br>
    <a href="https://discord.gg/6Z6FJZVaDV">
        <img src="https://img.shields.io/discord/780985109608005703?logo=discord"alt="Join the Discord">
    </a>
</p>
<h2 align="center">
    <a href="https://github.com/get-got/discord-downloader-go/releases/latest">
        <b>DOWNLOAD LATEST RELEASE BUILDS</b>
    </a>
</h2>

### **This project is a fork of [Seklfreak's _discord-image-downloader-go_](https://github.com/Seklfreak/discord-image-downloader-go)**
#### For list of differences and why I made an independent project, [**see below**](#differences-from-seklfreaks-discord-image-downloader-go--why-i-made-this).

This is a Discord bot program to download files posted in specified Discord channels to local folders. It can fetch highest possible quality files from various sources (listed below), aside from downloading any file directly linked or attached. _See [Features](#Features) below for full list._

## **WARNING:** Discord does not allow Automated User Accounts (Self-Bots)
[Read more in Discord Trust & Safety Team's Official Statement...](https://support.discordapp.com/hc/en-us/articles/115002192352-Automated-user-accounts-self-bots-)

> _NOTE: This only applies to real User Accounts, not Bot users. This program currently works for either._

## Getting Started (Basic Setup)
Obviously, download the bot program (or compile yourself with Go) and put it in a folder of your choice.

You can either create a `settings.json` following the examples & variables listed below, or have the program create a default file (if it is missing when you run the program, it will make one, and ask you if you want to enter in basic info for the new file). Ensure you follow proper JSON syntax to avoid any unexpected errors.

### For Credentials...
* If using a **Bot User,** enter the token into the `"Token"` setting. Remove the lines for `"Username"` and `"Password"` or leave blank (`""`). **To create a Bot User,** go to [discord.com/developers/applications](https://discord.com/developers/applications) and create a `New Application`. Once created, go to `Bot` and create. The token can be found on the `Bot` page. To invite to your server(s), go to `OAuth2` and check `"bot"`, copy the url, paste into browser and follow prompts for adding to server(s).
* If using a **Real User (Self-Bot),** fill out the `"Username"` and `"Password"` settings. Remove the line for `"Token"` or leave blank (`""`).
* If using a **Real User (Self-Bot) with 2FA (Two-Factor Authentication),** enter the token into the `"Token"` setting. Remove the lines for `"Username"` and `"Password"` or leave blank (`""`). Token can be found from `Developer Tools` in browser under `localStorage.token` or in the Discord client `Ctrl+Shift+I (Windows)`/`Cmd+Option+I (Mac)` under `Application → Local Storage → https://discordapp.com → "token"`.

### Bot Permissions in Channels/Servers
* In order to perform basic downloading functions, the bot will need `Read Message` permissions in the server(s) of your designated channel(s).
* In order to respond to commands, the bot will need `Send Message` permissions in the server(s) of your designated channel(s). If executing commands via an Admin Channel, the bot will only need `Send Message` permissions for that channel, and that permission will not be required for the source channel.
* In order to process history commands, the bot will need `Read Message History` permissions in the server(s) of your designated channel(s).

#### How to Find Discord ID's
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
    * Google Drive _(requires API Credentials, see config section)_
    * Tistory
    * Streamable
    * Gfycat
* ***Commands:***
    * Help _(<prefix>help - Alias: commands)_
    * Ping _(<prefix>ping - Alias: test)_
    * Status: Get an output of the current status of the bot _(<prefix>status - Alias: info)_
    * Stats: Have the bot dump stats _(<prefix>stats)_
    * **[Must be Bot or Server Admin]** History: Process all old messages in channel _(<prefix>history - Aliases: catalog, cache)_
    * **[Must be Bot Admin]** Exit (nice for process managers like pm2 for instant reload) _(<prefix>exit - Aliases: reload, kill)_

### Differences from [Seklfreak's _discord-image-downloader-go_](https://github.com/Seklfreak/discord-image-downloader-go) & Why I made this
* _Go 1.15 rather than 1.13_
* _discordgo 0.22.0 rather than 0.16.1_
* _Implements dgrouter for commands_
* Configuration is JSON-based rather than ini to allow more elaborate settings and better organization. With this came many features such as channel-specific settings.
* Channel-specific control of downloaded filetypes / content types (considers things like .mov as videos as well, rather than ignore them), Optional dividing of content types into separate folders.
* (Optional) Reactions upon download success.
* (Optional) Discord messages upon encountered errors.
* Extensive bot status/presence customization.
* Consistent Log Formatting, Color-Coded Logging
* Somewhat different organization than original project; initially created from scratch then components ported over.
* **Added Download Support for:** Facebook Videos
* Fixed Compatability Issue with `xurls` that required people to edit the project, regarding `xurls.Strict.FindAllString`. The issue was due to some people having xurls v2 installed while the projects go.mod required v1.1; changing go.mod to require v2 specifically seems to be the correct fix.

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

### Basic Settings Example
The following example is for a Bot Application _(using a token)_, bound to 1 channel.

This setup exempts many options so they will use default values _(see below)_. It shows the bare minimum required settings for the bot to function.

When initially launching the bot it will create a default settings file if you do not create your own `settings.json` beforehand.

`Example - Barebones settings.json:`
```javascript
{
    "credentials": {
        "token": "YOUR_TOKEN"
    },
    "channels": [
        {
            "channel": "DISCORD_CHANNEL_ID_TO_DOWNLOAD_FROM",
            "destination": "FOLDER_LOCATION_TO_DOWNLOAD_TO"
        }
    ]
}
```

`Example - Selfbot settings.json:`
```javascript
{
    "credentials": {
        "email": "REPLACE_WITH_YOUR_EMAIL",
        "password": "REPLACE_WITH_YOUR_PASSWORD"
    },
    "scanOwnMessages": true,
    "presenceEnabled": false,
    "channels": [
        {
            "channel": "DISCORD_CHANNEL_ID_TO_DOWNLOAD_FROM",
            "destination": "FOLDER_LOCATION_TO_DOWNLOAD_TO",
            "allowCommands": false,
            "errorMessages": false,
            "reactWhenDownloaded": false
        }
    ]
}
```

`Example - Advanced settings.json:`
```javascript
{
    "credentials": {
        "token": "YOUR_TOKEN",
        "twitterAccessToken": "",
        "twitterAccessTokenSecret": "",
        "twitterConsumerKey": "",
        "twitterConsumerSecret": ""
    },
    "admins": [ "YOUR_DISCORD_USER_ID", "YOUR_FRIENDS_DISCORD_USER_ID" ],
    "adminChannels": [
        {
            "channel": "CHANNEL_ID_FOR_ADMIN_CONTROL"
        }
    ],
    "debugOutput": true,
    "commandPrefix": "downloader ",
    "allowSkipping": true,
    "downloadRetryMax": 5,
    "downloadTimeout": 120,
    "githubUpdateChecking": true,
    "presenceStatus": "dnd",
    "presenceType": 3,
    "presenceOverwrite": "{{count}} files",
    "filenameDateFormat": "2006.01.02-15.04.05 ",
    "embedColor": "#29BEB0",
    "inflateCount": 1000,
    "channels": [
        {
            "channel": "THIS_CHANNEL_DOWNLOADS_EVERYTHING",
            "destination": "EVERYTHING",
            "overwriteEmbedColor": "#FF0000",
            "userBlacklist": [ "USER_ID_FOR_PERSON_I_DONT_LIKE" ],
            "divideFoldersByType": false,
            "saveImages": true,
            "saveVideos": true,
            "saveAudioFiles": true,
            "saveTextFiles": true,
            "saveOtherFiles": true,
            "savePossibleDuplicates": true,
            "extensionBlacklist": [
                ".htm",
                ".html",
                ".php"
            ]
        },
        {
            "channel": "THIS_CHANNEL_ONLY_DOWNLOADS_MEDIA",
            "destination": "media",
            "overwriteAllowSkipping": false,
            "saveImages": true,
            "saveVideos": true,
            "saveAudioFiles": true,
            "saveTextFiles": false,
            "saveOtherFiles": false
        },
        {
            "channel": "THIS_CHANNEL_IS_STEALTHY",
            "destination": "stealthy files",
            "allowCommands": false,
            "errorMessages": false,
            "updatePresence": false,
            "reactWhenDownloaded": false
        }
    ]
}
```

All JSON settings follow camelCase format.

### List of Settings
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
    * _`[OPTIONAL]`_ googleDriveCredentialsJSON `[string]`
        * _Path for Google Drive API credentials JSON file._
        * _Won't use Google Drive API for fetching files if credentials are missing._
* _`[OPTIONAL]`_ admins `[array of strings]`
    * Array of User ID strings for users allowed to use admin commands
* _`[OPTIONAL]`_ adminChannels `[array of key/value objects]`
    * **channel** `[string]`
* _`[DEFAULTS]`_ debugOutput `[bool]`
    * _Default:_ `false`
    * Output debugging information.
* _`[DEFAULTS]`_ commandPrefix `[string]`
    * _Default:_ `"ddg "`
* _`[DEFAULTS]`_ allowSkipping `[bool]`
    * _Default:_ `true`
    * Allow scanning for keywords to skip content downloading.
    * `"skip", "ignore", "don't save", "no save"`
* _`[DEFAULTS]`_ scanOwnMessages `[bool]`
    * _Default:_ `false`
    * Scans the bots own messages for content to download, only useful if using as a selfbot.
* _`[DEFAULTS]`_ downloadRetryMax `[int]`
    * _Default:_ `3`
* _`[DEFAULTS]`_ downloadTimeout `[int]`
    * _Default:_ `60`
* _`[DEFAULTS]`_ githubUpdateChecking `[bool]`
    * _Default:_ `true`
    * Check for updates from this repo.
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
        * `{{timeSavedShortTZ}}`: Last save time formatted as `3:04pm MST`
        * `{{timeSavedMid}}`: Last save time formatted as `3:04pm MST 1/2/2006`
        * `{{timeSavedLong}}`: Last save time formatted as `3:04:05pm MST - January 1, 2006`
        * `{{timeSavedShort24}}`: Last save time formatted as `15:04`
        * `{{timeSavedShortTZ24}}`: Last save time formatted as `15:04 MST`
        * `{{timeSavedMid24}}`: Last save time formatted as `15:04 MST 2/1/2006`
        * `{{timeSavedLong24}}`: Last save time formatted as `15:04:05 MST - 1 January, 2006`
        * `{{timeNowShort}}`: Current time formatted as `3:04pm`
        * `{{timeNowShortTZ}}`: Current time formatted as `3:04pm MST`
        * `{{timeNowMid}}`: Current time formatted as `3:04pm MST 1/2/2006`
        * `{{timeNowLong}}`: Current time formatted as `3:04:05pm MST - January 1, 2006`
        * `{{timeNowShort24}}`: Current time formatted as `15:04`
        * `{{timeNowShortTZ24}}`: Current time formatted as `15:04 MST`
        * `{{timeNowMid24}}`: Current time formatted as `15:04 MST 2/1/2006`
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
    * _Default:_ `"2006-01-02_15-04-05 "`
    * See [this Stack Overflow post regarding Golang date formatting.](https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format)
* _`[OPTIONAL]`_ embedColor `[string]`
    * _Unused by Default_
    * Supports `random`/`rand`, `role`/`user`, or RGB in hex or int format (ex: #FF0000 or 16711680).
* _`[OPTIONAL]`_ inflateCount `[int]`
    * _Unused by Default_
    * Inflates the count of total files downloaded by the bot. I only added this for my own personal use to represent an accurate total amount of files downloaded by previous bots I used.
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
        * Overwrites the global setting `filenameDateFormat` _(see above)_
        * See [this Stack Overflow post regarding Golang date formatting.](https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format)
    * _`[OPTIONAL]`_ overwriteAllowSkipping `[bool]`
        * _Unused by Default_
        * Allow scanning for keywords to skip content downloading.
        * `"skip", "ignore", "don't save", "no save"`
    * _`[OPTIONAL]`_ overwriteEmbedColor `[string]`
        * _Unused by Default_
        * Supports `random`/`rand`, `role`/`user`, or RGB in hex or int format (ex: #FF0000 or 16711680).
    * _`[DEFAULTS]`_ usersAllWhitelisted `[bool]`
        * _Default:_ `true`
        * Allow messages from all users to be handled. Set to `false` if you wish to use `userWhitelist` to only permit specific users messages to be handled.
    * _`[OPTIONAL]`_ userWhitelist `[array of strings]`
        * Use with `usersAllWhitelisted` as `false` to only permit specific users to have their messages handled by the bot. **Only accepts User ID's in the array.**
    * _`[OPTIONAL]`_ userBlacklist `[array of strings]`
        * Use with `usersAllWhitelisted` as the default `true` to block certain users messages from being handled by the bot. **Only accepts User ID's in the array.**
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
    * _`[DEFAULTS]`_ extensionBlacklist `[array of strings]`
        * _Default:_ `[ ".htm", ".html", ".php", ".exe", ".dll", ".bin", ".cmd", ".sh", ".py", ".jar" ]`
        * Ignores files containing specified extensions. Ensure you use proper formatting.
    * _`[OPTIONAL]`_ domainBlacklist `[array of strings]`
        * Ignores files from specified domains. Ensure you use proper formatting.
    * _`[OPTIONAL]`_ saveAllLinksToFile `[string]`
        * Saves all sent links to file, does not account for any filetypes or duplicates, it just simply appends every raw link sent in the channel to the specified file.

## Info for Developers
* I'm a complete amateur with Golang. If anything's bad please make a pull request.
* Versioning is `[MAJOR].[MINOR].[PATCH]`
* I try to be consistent with annotation but it's not perfect.
* Logging generally follows certain standards and patterns with formatting and color-coding.