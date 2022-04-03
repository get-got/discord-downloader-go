<h1 align="center">
    Discord Downloader <i>Go</i>
</h1>
<p align="center">
    <a href="https://travis-ci.com/get-got/discord-downloader-go" alt="Travis Build">
        <img src="https://travis-ci.com/get-got/discord-downloader-go.svg?branch=master" />
    </a>
    <a href="https://hub.docker.com/r/getgot/discord-downloader-go" alt="Docker Build">
        <img src="https://img.shields.io/docker/cloud/build/getgot/discord-downloader-go" />
    </a>
    <a href="https://goreportcard.com/report/github.com/get-got/discord-downloader-go" alt="Go Report Card">
        <img src="https://goreportcard.com/badge/github.com/get-got/discord-downloader-go" />
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
<h3 align="center">
    <a href="https://github.com/get-got/discord-downloader-go/releases/latest">
        <b>DOWNLOAD LATEST RELEASE</b>
    </a>
    <br/><br/>
    <a href="https://discord.com/invite/6Z6FJZVaDV">
        <b>Need help? Have suggestions? Join the Discord server!</b>
    </a>
</h3>

This is a program that connects to a Discord Bot or User to locally download files posted in Discord channels in real-time as well as old messages. It can download any directly linked files or Discord attachments, as well as the highest possible quality files from specific sources _(see list below)_. It also supports extensive channel-specific configuration and customization. _See [Features](#Features) below for full list!_

<h3 align="center">
    <b>This project is a fork of <a href="https://github.com/Seklfreak/discord-image-downloader-go">Seklfreak's <i>discord-image-downloader-go</i></a></b>
</h3>
<h4 align="center">
    For list of differences and why I made an independent project, <a href="#differences-from-seklfreaks-discord-image-downloader-go--why-i-made-this"><b>see below</b></a>
</h4>

---

### Sections
* [**List of Features**](#features)
* [**Getting Started**](#getting-started)
* [**Guide: Downloading History _(Old Messages)_**](#guide-downloading-history-old-messages)
* [**Guide: Settings / Configuration**](#guide-settings--configuration)
* [**List of Settings**](#list-of-settings)
* [**FAQ (Frequently Asked Questions)**](#faq)
* [**Development, Credits, Dependencies**](#development)

---

## Features

<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> LIST OF FEATURES & COMMANDS</b></summary>

### Supported Download Sources
* Discord File Attachments
* Direct Links to Files
* Twitter _(requires API key, see config section)_
* Instagram
* Reddit
* Imgur _(Single Posts & Albums)_
* Flickr _(requires API key, see config section)_
* Google Drive _(requires API Credentials, see config section)_
* Mastodon
* Tistory
* Streamable
* Gfycat
  
### Commands
Commands are used as `ddg <command> <?arguments?>` _(unless you've changed the prefix)_
Command     | Arguments? | Description
---         | ---   | ---
`help`, `commands`  | No    | Lists all commands.
`ping`, `test`      | No    | Pings the bot.
`info`      | No    | Displays relevant Discord info.
`status`    | No    | Shows the status of the bot.
`stats`     | No    | Shows channel stats.
`history`   | [**SEE HISTORY SECTION**](#guide-downloading-history-old-messages) | **(BOT AND SERVER ADMINS ONLY)** Processes history for old messages in channel.
`exit`, `kill`, `reload`    | No    | **(BOT ADMINS ONLY)** Exits the bot _(or restarts if using a keep-alive process manager)_.
`emojis`    | Optionally specify server IDs to download emojis from; separate by commas | **(BOT ADMINS ONLY)** Saves all emojis for channel.

</details>

---

## **WARNING!** Discord does not allow Automated User Accounts (Self-Bots/User-Bots)
[Read more in Discord Trust & Safety Team's Official Statement...](https://support.discordapp.com/hc/en-us/articles/115002192352-Automated-user-accounts-self-bots-)

While this project works for user logins, I do not reccomend it as you risk account termination. If you can, [use a proper Discord Bot user for this program.](https://discord.com/developers/applications)

> _NOTE: This only applies to real User Accounts, not Bot users. This program currently works for either._

---

## Getting Started
<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> GETTING STARTED, HOW-TO, OTHER INFO...</b></summary>

_Confused? Try looking at [the step-by-step list.](#getting-started-step-by-step)_

Depending on your purpose for this program, there are various ways you can run it.
- [Run the executable file for your platform. _(Process managers like **pm2** work well for this)_](https://github.com/get-got/discord-downloader-go/releases/latest)
- [Run automated image builds in Docker.](https://hub.docker.com/r/getgot/discord-downloader-go) _(Google it)._
  - Mount your settings.json to ``/root/settings.json``
  - Mount a folder named "database" to ``/root/database``
  - Mount your save folders or the parent of your save folders within ``/root/``
    - _i.e. ``X:\My Folder`` to ``/root/My Folder``_
- Install Golang and compile/run the source code yourself. _(Google it)_

You can either create a `settings.json` following the examples & variables listed below, or have the program create a default file (if it is missing when you run the program, it will make one, and ask you if you want to enter in basic info for the new file).
- [Ensure you follow proper JSON syntax to avoid any unexpected errors.](https://www.w3schools.com/js/js_json_syntax.asp)
- [Having issues? Try this JSON Validator to ensure it's correctly formatted.](https://jsonformatter.curiousconcept.com/)

### Getting Started Step-by-Step
1. Download & put executable within it's own folder.
2. Configure Main Settings (or run once to have settings generated). [_(SEE BELOW)_](#list-of-settings)
3. Enter your login credentials in the `"credentials"` section. [_(SEE BELOW)_](#list-of-settings)
4. Put your Discord User ID as in the `"admins"` list of the settings. [_(SEE BELOW)_](#list-of-settings)
5. Put a Discord Channel ID for a private channel you have access to into the `"adminChannels"`. [_(SEE BELOW)_](#list-of-settings)
6. Put your desired Discord Channel IDs into the `"channels"` section. [_(SEE BELOW)_](#list-of-settings)
- I know it can be confusing if you don't have experience with programming or JSON in general, but this was the ideal setup for extensive configuration like this. Just be careful with comma & quote placement and you should be fine. [See examples below for help.](#settings-examples)

### Bot Login Credentials...
* If using a **Bot Application,** enter the token into the `"token"` setting. Remove the lines for `"username"` and `"password"` or leave blank (`""`). **To create a Bot User,** go to [discord.com/developers/applications](https://discord.com/developers/applications) and create a `New Application`. Once created, go to `Bot` and create. The token can be found on the `Bot` page. To invite to your server(s), go to `OAuth2` and check `"bot"`, copy the url, paste into browser and follow prompts for adding to server(s).
* If using a **User Account (Self-Bot),** fill out the `"username"` and `"password"` settings. Remove the line for `"token"` or leave blank (`""`).
* If using a **User Account (Self-Bot) with 2FA (Two-Factor Authentication),** enter the token into the `"token"` setting. Remove the lines for `"username"` and `"password"` or leave blank (`""`). Token can be found from `Developer Tools` in browser under `localStorage.token` or in the Discord client `Ctrl+Shift+I (Windows)`/`Cmd+Option+I (Mac)` under `Application → Local Storage → https://discordapp.com → "token"`. **You must also set `userBot` within the `credentials` section of the settings.json to `true`.**

### Bot Permissions in Discord...
* In order to perform basic downloading functions, the bot will need `Read Message` permissions in the server(s) of your designated channel(s).
* In order to respond to commands, the bot will need `Send Message` permissions in the server(s) of your designated channel(s). If executing commands via an Admin Channel, the bot will only need `Send Message` permissions for that channel, and that permission will not be required for the source channel.
* In order to process history commands, the bot will need `Read Message History` permissions in the server(s) of your designated channel(s).

### How to Find Discord IDs...
* ***Use the info command!***
* **Discord Developer Mode:** Enable `Developer Mode` in Discord settings under `Appearance`.
* **Finding Channel ID:** _Enable Discord Developer Mode (see above),_ right click on the channel and `Copy ID`.
* **Finding User ID:** _Enable Discord Developer Mode (see above),_ right click on the user and `Copy ID`.
* **Finding Emoji ID:** _Enable Discord Developer Mode (see above),_ right click on the emoji and `Copy ID`.
* **Finding DM/PM ID:** Inspect Element on the DM icon for the desired user. Look for `href="/channels/@me/CHANNEL_ID_HERE"`. Using this ID in place of a normal channel ID should work perfectly fine.

---

### Differences from [Seklfreak's _discord-image-downloader-go_](https://github.com/Seklfreak/discord-image-downloader-go) & Why I made this
* _Better command formatting & support_
* Configuration is JSON-based rather than ini to allow more elaborate settings and better organization. With this came many features such as channel-specific settings.
* Channel-specific control of downloaded filetypes / content types (considers things like .mov as videos as well, rather than ignore them), Optional dividing of content types into separate folders.
* **Download Support for Reddit & Mastodon.**
* (Optional) Reactions upon download success.
* (Optional) Discord messages upon encountered errors.
* Extensive bot status/presence customization.
* Consistent Log Formatting, Color-Coded Logging
* Somewhat different organization than original project; initially created from scratch then components ported over.
* _Various fixes, improvements, and dependency updates that I also contributed to Seklfreak's original project._

> I've been a user of Seklfreak's project since ~2018 and it's been great for my uses, but there were certain aspects I wanted to expand upon, one of those being customization of channel configuration, and other features like message reactions upon success, differently formatted statuses, etc. If some aspects are rudimentary or messy, please make a pull request, as this is my first project using Go and I've learned everything from observation & Stack Overflow.

</details>

---

## Guide: Downloading History (Old Messages)
<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> HISTORY GUIDE</b></summary>

> This guide is to show you how to make the bot go through all old messages in a channel and catalog them as though they were being sent right now, in order to download them all.

### Command Arguments
If no channel IDs are specified, it will try and use the channel ID for the channel you're using the command in.

Argument / Flag         | Details
---                     | ---
**channel ID(s)**       | One or more channel IDs, separated by commas if multiple.
`all`                   | Use all available registered channels.
`cancel` or `stop`      | Stop downloading history for specified channel(s).
`--since=YYYY-MM-DD`    | Will process messages sent after this date.
`--since=message_id`    | Will process messages sent after this message.
`--before=YYYY-MM-DD`   | Will process messages sent before this date.
`--before=message_id`   | Will process messages sent before this message.

***Order of arguments does not matter.***

#### Examples
* `ddg history`
* `ddg history cancel`
* `ddg history all`
* `ddg history stop all`
* `ddg history 000111000111000`
* `ddg history 000111000111000, 000222000222000`
* `ddg history 000111000111000,000222000222000,000333000333000`
* `ddg history 000111000111000, 000333000333000 cancel`
* `ddg history 000111000111000 --before=000555000555000`
* `ddg history 000111000111000 --since=2020-01-02`
* `ddg history 000111000111000 --since=2020-10-12 --before=2021-05-06`
* `ddg history 000111000111000 --since=000555000555000 --before=2021-05-06`

</details>

---

## Guide: Settings / Configuration
> I tried to make the configuration as user friendly as possible, though you still need to follow proper JSON syntax (watch those commas). All settings specified below labeled `[DEFAULTS]` will use default values if missing from the settings file, and those labeled `[OPTIONAL]` will not be used if missing from the settings file.

When initially launching the bot it will create a default settings file if you do not create your own `settings.json` manually. All JSON settings follow camelCase format.

**If you have a ``config.ini`` from _Seklfreak's discord-image-downloader-go_, it will import settings if it's in the same folder as the program.**

### Settings Examples
The following example is for a Bot Application _(using a token)_, bound to 1 channel.

This setup exempts many options so they will use default values _(see below)_. It shows the bare minimum required settings for the bot to function.

<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> SETTINGS EXAMPLE - Barebones:</b></summary>

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

</details>

<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> SETTINGS EXAMPLE - Selfbot:</b></summary>

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

</details>

<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> SETTINGS EXAMPLE - Advanced:</b></summary>

```javascript
{
    "credentials": {
        "token": "YOUR_TOKEN",
        "twitterAccessToken": "aaa",
        "twitterAccessTokenSecret": "bbb",
        "twitterConsumerKey": "ccc",
        "twitterConsumerSecret": "ddd"
    },
    "admins": [ "YOUR_DISCORD_USER_ID", "YOUR_FRIENDS_DISCORD_USER_ID" ],
    "adminChannels": [
        {
            "channel": "CHANNEL_ID_FOR_ADMIN_CONTROL"
        }
    ],
    "debugOutput": true,
    "commandPrefix": "downloader_",
    "allowSkipping": true,
    "allowGlobalCommands": true,
    "asyncHistory": false,
    "downloadRetryMax": 5,
    "downloadTimeout": 120,
    "githubUpdateChecking": true,
    "discordLogLevel": 2,
    "filterDuplicateImages": true,
    "filterDuplicateImagesThreshold": 75,
    "presenceEnabled": true,
    "presenceStatus": "dnd",
    "presenceType": 3,
    "presenceOverwrite": "{{count}} files",
    "filenameDateFormat": "2006.01.02-15.04.05 ",
    "embedColor": "#EE22CC",
    "inflateCount": 12345,
    "channels": [
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
            "destination": "stealthy",
            "allowCommands": false,
            "errorMessages": false,
            "updatePresence": false,
            "reactWhenDownloaded": false
        },
        {
            "channels": [ "CHANNEL_1", "CHANNEL_2", "CHANNEL_3", "CHANNEL_4", "CHANNEL_5" ],
            "destination": "stuff",
            "allowCommands": false,
            "errorMessages": false,
            "updatePresence": false,
            "reactWhenDownloaded": false
        }
    ]
}
```

</details>

<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> SETTINGS EXAMPLE - Pretty Much Everything:</b></summary>

```javascript
{
    "_constants": {
        "DOWNLOAD_FOLDER":              "X:/Discord Downloads",
        "MY_TOKEN":                     "aaabbbccc111222333",
        "TWITTER_ACCESS_TOKEN_SECRET":  "aaabbbccc111222333",
        "TWITTER_ACCESS_TOKEN":         "aaabbbccc111222333",
        "TWITTER_CONSUMER_KEY":         "aaabbbccc111222333",
        "TWITTER_CONSUMER_SECRET":      "aaabbbccc111222333",
        "FLICKR_API_KEY":               "aaabbbccc111222333",
        "GOOGLE_DRIVE_CREDS":           "googleDriveCreds.json",

        "MY_USER_ID":       "000111222333444555",
        "BOBS_USER_ID":     "000111222333444555",

        "SERVER_MAIN":               "000111222333444555",
        "CHANNEL_MAIN_GENERAL":      "000111222333444555",
        "CHANNEL_MAIN_MEMES":        "000111222333444555",
        "CHANNEL_MAIN_SPAM":         "000111222333444555",
        "CHANNEL_MAIN_PHOTOS":       "000111222333444555",
        "CHANNEL_MAIN_ARCHIVE":      "000111222333444555",
        "CHANNEL_MAIN_BOT_ADMIN":    "000111222333444555",

        "SERVER_BOBS":              "000111222333444555",
        "CHANNEL_BOBS_GENERAL":     "000111222333444555",
        "CHANNEL_BOBS_MEMES":       "000111222333444555",
        "CHANNEL_BOBS_SPAM":        "000111222333444555",
        "CHANNEL_BOBS_BOT_ADMIN":   "000111222333444555",

        "SERVER_GAMERZ":                "000111222333444555",
        "CHANNEL_GAMERZ_GENERAL":       "000111222333444555",
        "CHANNEL_GAMERZ_MEMES":         "000111222333444555",
        "CHANNEL_GAMERZ_VIDEOS":        "000111222333444555",
        "CHANNEL_GAMERZ_SPAM":          "000111222333444555",
        "CHANNEL_GAMERZ_SCREENSHOTS":   "000111222333444555"
    },
    "credentials": {
        "token": "MY_TOKEN",
        "userBot": true,
        "twitterAccessToken": "TWITTER_ACCESS_TOKEN",
        "twitterAccessTokenSecret": "TWITTER_ACCESS_TOKEN_SECRET",
        "twitterConsumerKey": "TWITTER_CONSUMER_KEY",
        "twitterConsumerSecret": "TWITTER_CONSUMER_SECRET",
        "flickrApiKey": "FLICKR_API_KEY",
        "googleDriveCredentialsJSON": "GOOGLE_DRIVE_CREDS"
    },
    "admins": [ "MY_USER_ID", "BOBS_USER_ID" ],
    "adminChannels": [
        {
            "channel": "CHANNEL_MAIN_BOT_ADMIN"
        },
        {
            "channel": "CHANNEL_BOBS_BOT_ADMIN"
        }
    ],
    "debugOutput": true,
    "commandPrefix": "d_",
    "allowSkipping": true,
    "scanOwnMessages": true,
    "checkPermissions": false,
    "allowGlobalCommands": false,
    "autorunHistory": true,
    "asyncHistory": false,
    "downloadRetryMax": 5,
    "downloadTimeout": 120,
    "discordLogLevel": 3,
    "githubUpdateChecking": false,
    "filterDuplicateImages": true,
    "filterDuplicateImagesThreshold": 50,
    "presenceEnabled": true,
    "presenceStatus": "idle",
    "presenceType": 3,
    "presenceOverwrite": "{{count}} things",
    "presenceOverwriteDetails": "these are my details",
    "presenceOverwriteState": "this is my state",
    "filenameDateFormat": "2006.01.02_15.04.05_",
    "embedColor": "#FF0000",
    "inflateCount": 69,
    "numberFormatEuropean": true,
    "all": {
        "destination": "DOWNLOAD_FOLDER/Unregistered",
        "allowCommands": false,
        "errorMessages": false,
        "scanEdits": true,
        "ignoreBots": false,
        "overwriteAutorunHistory": false,
        "updatePresence": false,
        "reactWhenDownloaded": false,
        "typeWhileProcessing": false,
        "divideFoldersByServer": true,
        "divideFoldersByChannel": true,
        "divideFoldersByUser": false,
        "divideFoldersByType": false,
        "saveImages": true,
        "saveVideos": true,
        "saveAudioFiles": true,
        "saveTextFiles": false,
        "saveOtherFiles": true,
        "savePossibleDuplicates": true,
        "filters": {
            "blockedExtensions": [
                ".htm",
                ".html",
                ".php",
                ".bat",
                ".sh",
                ".jar",
                ".exe"
            ]
        },
        "logLinks": {
            "destination": "log_links",
            "destinationIsFolder": true,
            "divideLogsByServer": true,
            "divideLogsByChannel": true,
            "divideLogsByUser": true,
            "userData": true
        },
        "logMessages": {
            "destination": "log_messages",
            "destinationIsFolder": true,
            "divideLogsByServer": true,
            "divideLogsByChannel": true,
            "divideLogsByUser": true,
            "userData": true
        }
    },
    "allBlacklistChannels": [ "CHANNEL_I_DONT_LIKE", "OTHER_CHANNEL_I_DONT_LIKE" ],
    "allBlacklistServers": [ "SERVER_MAIN", "SERVER_BOBS" ],
    "servers": [
        {
            "server": "SERVER_MAIN",
            "destination": "DOWNLOAD_FOLDER/- My Server",
            "divideFoldersByChannel": true
        },
        {
            "servers": [ "SERVER_BOBS", "SERVER_GAMERZ" ],
            "destination": "DOWNLOAD_FOLDER/- Friends Servers",
            "divideFoldersByServer": true,
            "divideFoldersByChannel": true
        }
    ],
    "channels": [
        {
            "channel": "CHANNEL_MAIN_SPAM",
            "destination": "DOWNLOAD_FOLDER/Spam",
            "overwriteAllowSkipping": false,
            "saveImages": true,
            "saveVideos": true,
            "saveAudioFiles": true,
            "saveTextFiles": false,
            "saveOtherFiles": false
        },
        {
            "channel": "CHANNEL_BOBS_SPAM",
            "destination": "DOWNLOAD_FOLDER/Spam - Bob",
            "overwriteAllowSkipping": false,
            "saveImages": true,
            "saveVideos": true,
            "saveAudioFiles": true,
            "saveTextFiles": false,
            "saveOtherFiles": false
        },
        {
            "channels": [ "CHANNEL_MAIN_MEMES", "CHANNEL_BOBS_MEMES", "CHANNEL_GAMERZ_MEMES" ],
            "destination": "DOWNLOAD_FOLDER/Our Memes",
            "allowCommands": true,
            "errorMessages": true,
            "updatePresence": true,
            "reactWhenDownloaded": true,
            "saveImages": true,
            "saveVideos": true,
            "saveAudioFiles": false,
            "saveTextFiles": false,
            "saveOtherFiles": true
        }
    ]
}
```

</details>

---

## List of Settings
:small_red_triangle: means the setting (or alternative) is **required**.

:small_blue_diamond: means the setting defaults to a prespecified value. List below should say all default values.

:small_orange_diamond: means the setting is optional and the feature(s) related to the setting will not be implemented if missing.

<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> LIST OF ALL SETTINGS</b></summary>

* :small_orange_diamond: **"_constants"**
    * — _settings.\_constants : list of name:value strings_
    * Use constants to replace values throughout the rest of the settings.
        * ***Note:*** _If a constants name is used within another longer constants name, make sure the longer one is higher in order than the shorter one, otherwise the longer one will not be used properly. (i.e. if you have MY\_CONSTANT and MY\_CONSTANT\_TWO, put MY\_CONSTANT\_TWO above MY\_CONSTANT)_
    * **Basic Example:**
    ```json
    {
        "_constants": {
            "MY_TOKEN": "my token here",
            "ADMIN_CHANNEL": "123456789"
        },
        "credentials": {
            "token": "MY_TOKEN"
        },
        "adminChannels": {
            "channel": "ADMIN_CHANNEL"
        }
    }
    ```
---
* :small_red_triangle: **"credentials"**
    * — _settings.credentials : setting:value list_
    * :small_red_triangle: **"token"**
        * — _settings.credentials.token : string_
        * _REQUIRED FOR BOT APPLICATION LOGIN OR USER LOGIN WITH 2FA, don't include if using User Login without 2FA._
    * :small_red_triangle: **"email"**
        * — _settings.credentials.email : string_
        * _REQUIRED FOR USER LOGIN WITHOUT 2FA, don't include if using Bot Application Login._
    * :small_red_triangle: **"password"**
        * — _settings.credentials.password : string_
        * _REQUIRED FOR USER LOGIN WITHOUT 2FA, don't include if using Bot Application Login._
    * :small_blue_diamond: **"userBot"**
        * — _settings.credentials.userBot : boolean_
        * _Default:_ `false`
        * _SET TO `true` FOR A USER LOGIN WITH 2FA, keep as `false` if using a Bot Application._
    ---
    * :small_orange_diamond: "twitterAccessToken"
        * — _settings.credentials.twitterAccessToken : string_
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * :small_orange_diamond: "twitterAccessTokenSecret"
        * — _settings.credentials.twitterAccessTokenSecret : string_
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * :small_orange_diamond: "twitterConsumerKey"
        * — _settings.credentials.twitterConsumerKey : string_
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * :small_orange_diamond: "twitterConsumerSecret"
        * — _settings.credentials.twitterConsumerSecret : string_
        * _Won't use Twitter API for fetching media from tweets if credentials are missing._
    * :small_orange_diamond: "flickrApiKey"
        * — _settings.credentials.flickrApiKey : string_
        * _Won't use Flickr API for fetching media from posts/albums if credentials are missing._
    * :small_orange_diamond: "googleDriveCredentialsJSON"
        * — _settings.credentials.googleDriveCredentialsJSON : string_
        * _Path for Google Drive API credentials JSON file._
        * _Won't use Google Drive API for fetching files if credentials are missing._
---
* :small_orange_diamond: "admins"
    * — _settings.admins : list of strings_
    * List of User ID strings for users allowed to use admin commands
* :small_orange_diamond: "adminChannels"
    * — _settings.adminChannels : list of setting:value groups_
    * :small_red_triangle: **"channel"** _`[USE THIS OR "channels"]`_
        * — _settings.adminChannel.channel : string_
        * _Channel ID for admin commands & logging._
    * :small_red_triangle: **"channels"** _`[USE THIS OR "channel"]`_
        * — _settings.adminChannel.channels : list of strings_
        * Channel IDs to monitor, for if you want the same configuration for multiple channels.
    * :small_blue_diamond: "logStatus"
        * — _settings.adminChannel.logStatus : boolean_
        * _Default:_ `true`
        * _Send status messages to admin channel(s) upon launch._
    * :small_blue_diamond: "logErrors"
        * — _settings.adminChannel.logErrors : boolean_
        * _Default:_ `true`
        * _Send error messages to admin channel(s) when encountering errors._
    * :small_blue_diamond: "unlockCommands"
        * — _settings.adminChannel.unlockCommands : boolean_
        * _Default:_ `false`
        * _Unrestrict admin commands so anyone can use within this admin channel._
---
* :small_blue_diamond: "debugOutput"
    * — _settings.debugOutput : boolean_
    * _Default:_ `false`
    * Output debugging information.
* :small_blue_diamond: "messageOutput"
    * — _settings.messageOutput : boolean_
    * _Default:_ `true`
    * Output handled Discord messages.
* :small_blue_diamond: "commandPrefix"
    * — _settings.commandPrefix : string_
    * _Default:_ `"ddg "`
* :small_blue_diamond: "allowSkipping"
    * — _settings.allowSkipping : boolean_
    * _Default:_ `true`
    * Allow scanning for keywords to skip content downloading.
    * `"skip", "ignore", "don't save", "no save"`
* :small_blue_diamond: "scanOwnMessages"
    * — _settings.scanOwnMessages : boolean_
    * _Default:_ `false`
    * Scans the bots own messages for content to download, only useful if using as a selfbot.
* :small_blue_diamond: "checkPermissions"
    * — _settings.checkPermissions : boolean_
    * _Default:_ `true`
    * Checks Discord permissions before attempting requests/actions.
* :small_blue_diamond: "allowGlobalCommands"
    * — _settings.allowGlobalCommands : boolean_
    * _Default:_ `true`
    * Allow certain commands to be used even if not registered in `channels` or `adminChannels`.
* :small_orange_diamond: "autorunHistory"
    * — _settings.autorunHistory : boolean_
    * Autorun history for all registered channels in background upon launch.
    * _This can take anywhere between 2 minutes and 2 hours. It depends on how many channels your bot monitors and how many messages it has to go through. It can help to disable it by-channel for channels that don't require it (see `overwriteAutorunHistory` in channel options)._
* :small_orange_diamond: "asyncHistory"
    * — _settings.asyncHistory : boolean_
    * Runs history commands simultaneously rather than one after the other.
      * **WARNING!!! May result in Discord API Rate Limiting with many channels**, difficulty troubleshooting, exploding CPUs, melted RAM.
* :small_blue_diamond: "downloadRetryMax"
    * — _settings.downloadRetryMax : number_
    * _Default:_ `3`
* :small_blue_diamond: "downloadTimeout"
    * — _settings.downloadTimeout : number_
    * _Default:_ `60`
* :small_blue_diamond: "githubUpdateChecking"
    * — _settings.githubUpdateChecking : boolean_
    * _Default:_ `true`
    * Check for updates from this repo.
* :small_blue_diamond: "discordLogLevel"
    * — _settings.discordLogLevel : number_
    * _Default:_ `0`
    * 0 = LogError
    * 1 = LogWarning
    * 2 = LogInformational
    * 3 = LogDebug _(everything)_
* :small_blue_diamond: "filterDuplicateImages"
    * — _settings.filterDuplicateImages : boolean_
    * _Default:_ `false`
    * **Experimental** feature to filter out images that are too similar to other cached images.
    * _Caching of image data is stored via a database file; it will not read all pre-existing images._
* :small_blue_diamond: "filterDuplicateImagesThreshold"
    * — _settings.filterDuplicateImagesThreshold : number with decimals_
    * _Default:_ `0`
    * Threshold for what the bot considers too similar of an image comparison score. Lower = more similar (lowest is around -109.7), Higher = less similar (does not really have a maximum, would require your own testing).
---
* :small_blue_diamond: "presenceEnabled"
    * — _settings.presenceEnabled : boolean_
    * _Default:_ `true`
* :small_blue_diamond: "presenceStatus"
    * — _settings.presenceStatus : string_
    * _Default:_ `"idle"`
    * Presence status type.
    * `"online"`, `"idle"`, `"dnd"`, `"invisible"`, `"offline"`
* :small_blue_diamond: "presenceType"
    * — _settings.presenceType : number_
    * _Default:_ `0`
    * Presence label type. _("Playing \<activity\>", "Listening to \<activity\>", etc)_
    * `Game = 0, Streaming = 1, Listening = 2, Watching = 3, Custom = 4`
        * If Bot User, Streaming & Custom won't work properly.
* :small_orange_diamond: "presenceOverwrite"
    * — _settings.presenceOverwrite : string_
    * _Unused by Default_
    * Replace counter status with custom string.
    * [see Presence Placeholders for customization...](#presence-placeholders-for-settings)
* :small_orange_diamond: "presenceOverwriteDetails"
    * — _settings.presenceOverwriteDetails : string_
    * _Unused by Default_
    * Replace counter status details with custom string (only works for User, not Bot).
    * [see Presence Placeholders for customization...](#presence-placeholders-for-settings)
* :small_orange_diamond: "presenceOverwriteState"
    * — _settings.presenceOverwriteState : string_
    * _Unused by Default_
    * Replace counter status state with custom string (only works for User, not Bot).
    * [see Presence Placeholders for customization...](#presence-placeholders-for-settings)
---
    * :small_blue_diamond: "reactWhenDownloaded"
        * — _settings.reactWhenDownloaded : boolean_
        * _Default:_ `true`
        * Confirmation reaction that file(s) successfully downloaded. Is overwritten by the channel/server equivelant of this setting.
---
* :small_blue_diamond: "filenameDateFormat"
    * — _settings.filenameDateFormat : string_
    * _Default:_ `"2006-01-02_15-04-05 "`
    * [see this Stack Overflow post regarding Golang date formatting.](https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format)
* :small_orange_diamond: "embedColor"
    * — _settings.embedColor : string_
    * _Unused by Default_
    * Supports `random`/`rand`, `role`/`user`, or RGB in hex or int format (ex: #FF0000 or 16711680).
* :small_orange_diamond: "inflateCount"
    * — _settings.inflateCount : number_
    * _Unused by Default_
    * Inflates the count of total files downloaded by the bot. I only added this for my own personal use to represent an accurate total amount of files downloaded by previous bots I used.
* :small_blue_diamond: "numberFormatEuropean"
    * — _settings.numberFormatEuropean : boolean_
    * _Default:_ false
    * Formats numbers as `123.456,78`/`123.46k` rather than `123,456.78`/`123,46k`.
---
* :small_orange_diamond: **"all"**
    * — _settings.all : list of setting:value options_
    * **Follow `channels` below for variables, except channel & server ID(s) are not used.**
    * If a pre-existing config for the channel or server is not found, it will download from **any and every channel** it has access to using your specified settings.
* :small_orange_diamond: "allBlacklistServers"
    * — _settings.allBlacklistServers : list of strings_
    * _Unused by Default_
    * Blacklists servers (by ID) from `all`.
* :small_orange_diamond: "allBlacklistChannels"
    * — _settings.allBlacklistChannels : list of strings_
    * _Unused by Default_
    * Blacklists channels (by ID) from `all`.
---
* :small_red_triangle: **"servers"** _`[USE THIS OR "channels"]`_
    * — _settings.servers : list of setting:value groups_
    * :small_red_triangle: **"server"** _`[USE THIS OR "servers"]`_
        * — _settings.servers[].server : string_
        * Server ID to monitor.
    * :small_red_triangle: **"servers"** _`[USE THIS OR "server"]`_
        * — _settings.servers[].servers : list of strings_
        * Server IDs to monitor, for if you want the same configuration for multiple servers.
    * :small_orange_diamond: "blacklistChannels"
        * — _settings.servers[].blacklistChannels : list of strings_
        * Blacklist specific channels from the encompassing server(s).
    * **ALL OTHER VARIABLES ARE SAME AS "channels" BELOW**
* :small_red_triangle: **"channels"** _`[USE THIS OR "servers"]`_
    * — _settings.channels : list of setting:value groups_
    * :small_red_triangle: **"channel"** _`[USE THIS OR "channels"]`_
        * — _settings.channels[].channel : string_
        * Channel ID to monitor.
    * :small_red_triangle: **"channels"** _`[USE THIS OR "channel"]`_
        * — _settings.channels[].channels : list of strings_
        * Channel IDs to monitor, for if you want the same configuration for multiple channels.
    ---
    * :small_red_triangle: **"destination"**
        * — _settings.channels[].destination : string_
        * Folder path for saving files, can be full path or local subfolder.
    * :small_blue_diamond: "enabled"
        * — _settings.channels[].enabled : boolean_
        * _Default:_ `true`
        * Toggles bot functionality for channel.
    * :small_blue_diamond: "save"
        * — _settings.channels[].save : boolean_
        * _Default:_ `true`
        * Toggles whether the files actually get downloaded/saved.
    * :small_blue_diamond: "allowCommands"
        * — _settings.channels[].allowCommands : boolean_
        * _Default:_ `true`
        * Allow use of commands like ping, help, etc.
    * :small_blue_diamond: "errorMessages"
        * — _settings.channels[].errorMessages : boolean_
        * _Default:_ `true`
        * Send response messages when downloads fail or other download-related errors are encountered.
    * :small_blue_diamond: "scanEdits"
        * — _settings.channels[].scanEdits : boolean_
        * _Default:_ `true`
        * Check edits for un-downloaded media.
    * :small_blue_diamond: "ignoreBots"
        * — _settings.channels[].ignoreBots : boolean_
        * _Default:_ `false`
        * Ignores messages from Bot users.
    * :small_orange_diamond: overwriteAutorunHistory
        * — _settings.channels[].overwriteAutorunHistory : boolean_
        * Overwrite global setting for autorunning history for all registered channels in background upon launch.
    * :small_orange_diamond: "sendFileToChannel"
        * — _settings.channels[].sendFileToChannel : string_
        * Forwards/crossposts/logs downloaded files to specified channel (or channels if used as `sendFileToChannels` below). By default will send as the actual file
    * :small_orange_diamond: "sendFileToChannels"
        * — _settings.channels[].sendFileToChannels : list of strings_
        * List form of `sendFileToChannel` above.
    * :small_blue_diamond: "sendFileDirectly"
        * — _settings.channels[].sendFileDirectly : boolean_
        * _Default:_ `true`
        * Sends raw file to channel(s) rather than embedded download link.
    ---
    * :small_blue_diamond: "updatePresence"
        * — _settings.channels[].updatePresence : boolean_
        * _Default:_ `true`
        * Update Discord Presence when download succeeds within this channel.
    * :small_blue_diamond: "reactWhenDownloaded"
        * — _settings.channels[].reactWhenDownloaded : boolean_
        * _Default:_ `true`
        * Confirmation reaction that file(s) successfully downloaded.
    * :small_orange_diamond: "reactWhenDownloadedEmoji"
        * — _settings.channels[].reactWhenDownloadedEmoji : string_
        * _Unused by Default_
        * Uses specified emoji rather than random server emojis. Simply pasting a standard emoji will work, for custom Discord emojis use "name:ID" format.
    * :small_blue_diamond: "reactWhenDownloadedHistory"
        * — _settings.channels[].reactWhenDownloadedHistory : boolean_
        * _Default:_ `false`
        * Reacts to old messages when processing history.
    * :small_blue_diamond: "blacklistReactEmojis"
        * — _settings.channels[].blacklistReactEmojis : list of strings_
        * _Unused by Default_
        * Block specific emojis from being used for reacts. Simply pasting a standard emoji will work, for custom Discord emojis use "name:ID" format.
    * :small_blue_diamond: "typeWhileProcessing"
        * — _settings.channels[].typeWhileProcessing : boolean_
        * _Default:_ `false`
        * Shows _"<name> is typing..."_ while processing things that aren't processed instantly, like history cataloging.
    * :small_orange_diamond: "overwriteFilenameDateFormat"
        * — _settings.channels[].overwriteFilenameDateFormat : string_
        * _Unused by Default_
        * Overwrites the global setting `filenameDateFormat` _(see above)_
        * [see this Stack Overflow post regarding Golang date formatting.](https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format)
    * :small_orange_diamond: "overwriteAllowSkipping"
        * — _settings.channels[].overwriteAllowSkipping : boolean_
        * _Unused by Default_
        * Allow scanning for keywords to skip content downloading.
        * `"skip", "ignore", "don't save", "no save"`
    * :small_orange_diamond: "overwriteEmbedColor"
        * — _settings.channels[].overwriteEmbedColor : string_
        * _Unused by Default_
        * Supports `random`/`rand`, `role`/`user`, or RGB in hex or int format (ex: #FF0000 or 16711680).
    ---
    * :small_blue_diamond: "divideFoldersByServer"
        * — _settings.channels[].divideFoldersByServer : boolean_
        * _Default:_ `false`
        * Separate files into subfolders by server of origin _(e.g. "My Server", "My Friends Server")_
    * :small_blue_diamond: "divideFoldersByChannel"
        * — _settings.channels[].divideFoldersByChannel : boolean_
        * _Default:_ `false`
        * Separate files into subfolders by channel of origin _(e.g. "my-channel", "my-other-channel")_
    * :small_blue_diamond: "divideFoldersByUser"
        * — _settings.channels[].divideFoldersByUser : boolean_
        * _Default:_ `false`
        * Separate files into subfolders by user who sent _(e.g. "Me#1234", "My Friend#0000")_
    * :small_blue_diamond: "divideFoldersByType"
        * — _settings.channels[].divideFoldersByType : boolean_
        * _Default:_ `true`
        * Separate files into subfolders by type _(e.g. "images", "video", "audio", "text", "other")_
    * :small_blue_diamond: "divideFoldersUseID"
        * — _settings.channels[].divideFoldersUseID : boolean_
        * _Default:_ `false`
        * Uses ID rather than Name for `"divideFoldersByServer"`, `"divideFoldersByChannel"`, `"divideFoldersByUser"`. I would recommend this if any servers you download from have server/channel/usernames changed frequently.
    * :small_blue_diamond: "saveImages"
        * — _settings.channels[].saveImages : boolean_
        * _Default:_ `true`
    * :small_blue_diamond: "saveVideos"
        * — _settings.channels[].saveVideos : boolean_
        * _Default:_ `true`
    * :small_blue_diamond: "saveAudioFiles"
        * — _settings.channels[].saveAudioFiles : boolean_
        * _Default:_ `false`
    * :small_blue_diamond: "saveTextFiles"
        * — _settings.channels[].saveTextFiles : boolean_
        * _Default:_ `false`
    * :small_blue_diamond: "saveOtherFiles"
        * — _settings.channels[].saveOtherFiles : boolean_
        * _Default:_ `false`
    * :small_blue_diamond: "savePossibleDuplicates"
        * — _settings.channels[].savePossibleDuplicates : boolean_
        * _Default:_ `false`
        * Save file even if exact filename already exists or exact URL is already recorded in database.
    ---
    * :small_orange_diamond: "filters"
        * — _settings.channels[].filters : setting:value group_
        * _Filter prioritizes Users before Roles before Phrases._
        * :small_blue_diamond: "blockedPhrases"
            * — _settings.channels[].filters.blockedPhrases : list of strings_
            * List of phrases to make the bot ignore this message.
            * Will ignore any message containing a blocked phrase UNLESS it also has an allowed phrase. Messages will be processed by default.
            * _Default:_ `[ "skip", "ignore", "don't save", "no save" ]`
        * :small_orange_diamond: "allowedPhrases"
            * — _settings.channels[].filters.allowedPhrases : list of strings_
            * List of phrases to allow the bot to process the message.
            * _If used without blockedPhrases,_ no messages will be processed unless they contain an allowed phrase.
        * :small_orange_diamond: "blockedUsers"
            * — _settings.channels[].filters.blockedUsers : list of strings_
            * Will ignore messages from the following users.
        * :small_orange_diamond: "allowedUsers"
            * — _settings.channels[].filters.allowedUsers : list of strings_
            * Will ONLY process messages if they were sent from the following users.
        * :small_orange_diamond: "blockedRoles"
            * — _settings.channels[].filters.blockedRoles : list of strings_
            * Will ignore messages from users with any of the following roles.
        * :small_orange_diamond: "allowedRoles"
            * — _settings.channels[].filters.allowedRoles : list of strings_
            * Will ONLY process messages if they were sent from users with any of the following roles.
        * :small_blue_diamond: "blockedExtensions"
            * — _settings.channels[].filters.blockedExtensions : list of strings_
            * List of file extensions for the bot to ignore (include periods).
            * _Default:_ `[ ".htm", ".html", ".php", ".exe", ".dll", ".bin", ".cmd", ".sh", ".py", ".jar" ]`
        * :small_orange_diamond: "allowedExtensions"
            * — _settings.channels[].filters.allowedExtensions : list of strings_
            * Will ONLY process files if they have the following extensions (include periods).
        * :small_orange_diamond: "blockedDomains"
            * — _settings.channels[].filters.blockedDomains : list of strings_
            * List of file source domains (websites) for the bot to ignore.
        * :small_orange_diamond: "allowedDomains"
            * — _settings.channels[].filters.allowedDomains : list of strings_
            * Will ONLY process files if they were sent from any of the following domains (websites).
    ---
    * :small_orange_diamond: "logLinks"
        * — _settings.channels[].logLinks : setting:value group_
        * :small_red_triangle: "destination"
            * — _settings.channels[].logLinks.destination : string_
            * Filepath for single log file to be stored, or directory path for multiple logs to be stored.
        * :small_blue_diamond: "destinationIsFolder"
            * — _settings.channels[].logLinks.destinationIsFolder : bool_
            * _Default:_ `false`
            * `true` if `"destination"` above is for a directory for multiple logs.
        * :small_blue_diamond: "divideLogsByServer"
            * — _settings.channels[].logLinks.divideLogsByServer : bool_
            * _Default:_ `true`
            * *ONLY USED IF `"destinationIsFolder"` ABOVE IS `true`*
            * Separates log files by Server ID.
        * :small_blue_diamond: "divideLogsByChannel"
            * — _settings.channels[].logLinks.divideLogsByChannel : bool_
            * _Default:_ `true`
            * *ONLY USED IF `"destinationIsFolder"` ABOVE IS `true`*
            * Separates log files by Channel ID.
        * :small_blue_diamond: "divideLogsByUser"
            * — _settings.channels[].logLinks.divideLogsByUser : bool_
            * _Default:_ `false`
            * *ONLY USED IF `"destinationIsFolder"` ABOVE IS `true`*
            * Separates log files by User ID.
        * :small_blue_diamond: "divideLogsByStatus"
            * — _settings.channels[].logLinks.divideLogsByStatus : bool_
            * _Default:_ `false`
            * *ONLY USED IF `"destinationIsFolder"` ABOVE IS `true`*
            * Separates log files download status.
            * *DOES NOT APPLY TO `"logMessages"` BELOW*
        * :small_blue_diamond: "logDownloads"
            * — _settings.channels[].logLinks.logDownloads : bool_
            * _Default:_ `true`
            * Includes successfully downloaded links in logs.
            * *DOES NOT APPLY TO `"logMessages"` BELOW*
        * :small_blue_diamond: "logFailures"
            * — _settings.channels[].logLinks.logFailures : bool_
            * _Default:_ `true`
            * Includes failed/skipped/ignored links in logs.
            * *DOES NOT APPLY TO `"logMessages"` BELOW*
        * :small_blue_diamond: "filterDuplicates"
            * — _settings.channels[].logLinks.filterDuplicates : bool_
            * _Default:_ `false`
            * Filters out duplicate links (or messages) from being logged if already present in log file.
        * :small_orange_diamond: "prefix"
            * — _settings.channels[].logLinks.prefix : string_
            * Prepend log line with string.
        * :small_orange_diamond: "suffix"
            * — _settings.channels[].logLinks.suffix : string_
            * Append log line with string.
        * :small_blue_diamond: "userData"
            * — _settings.channels[].logLinks.userData : bool_
            * _Default:_ `false`
            * Include additional data such as SERVER/CHANNEL/USER ID's for logged files/messages.
    * :small_orange_diamond: "logMessages"
        * ***Identical to `"logLinks"` above unless noted otherwise.***

</details>

---

### Presence Placeholders for Settings
_For `presenceOverwrite`, `presenceOverwriteDetails`, `presenceOverwriteState`_
<details>
<summary><b><i>(COLLAPSABLE SECTION)</i></b></summary>

Key | Description
--- | ---
`{{dgVersion}}`             | discord-go version
`{{ddgVersion}}`            | Project version
`{{apiVersion}}`            | Discord API version
`{{countNoCommas}}`         | Raw total count of downloads (without comma formatting)
`{{count}}`                 | Raw total count of downloads
`{{countShort}}`            | Shortened total count of downloads
`{{numServers}}`            | Number of servers bot is in
`{{numBoundServers}}`       | Number of bound servers
`{{numBoundChannels}}`      | Number of bound channels
`{{numAdminChannels}}`      | Number of admin channels
`{{numAdmins}}`             | Number of designated admins
`{{timeSavedShort}}`        | Last save time formatted as `3:04pm`
`{{timeSavedShortTZ}}`      | Last save time formatted as `3:04pm MST`
`{{timeSavedMid}}`          | Last save time formatted as `3:04pm MST 1/2/2006`
`{{timeSavedLong}}`         | Last save time formatted as `3:04:05pm MST - January 2, 2006`
`{{timeSavedShort24}}`      | Last save time formatted as `15:04`
`{{timeSavedShortTZ24}}`    | Last save time formatted as `15:04 MST`
`{{timeSavedMid24}}`        | Last save time formatted as `15:04 MST 2/1/2006`
`{{timeSavedLong24}}`       | Last save time formatted as `15:04:05 MST - 2 January, 2006`
`{{timeNowShort}}`          | Current time formatted as `3:04pm`
`{{timeNowShortTZ}}`        | Current time formatted as `3:04pm MST`
`{{timeNowMid}}`            | Current time formatted as `3:04pm MST 1/2/2006`
`{{timeNowLong}}`           | Current time formatted as `3:04:05pm MST - January 2, 2006`
`{{timeNowShort24}}`        | Current time formatted as `15:04`
`{{timeNowShortTZ24}}`      | Current time formatted as `15:04 MST`
`{{timeNowMid24}}`          | Current time formatted as `15:04 MST 2/1/2006`
`{{timeNowLong24}}`         | Current time formatted as `15:04:05 MST - 2 January, 2006`
`{{uptime}}`                | Shortened duration of bot uptime

</details>

---

## FAQ
* ***Q: How do I install?***
* **A: [SEE #getting-started](#getting-started)** 
---
* ***Q: How do I convert from Seklfreak's discord-image-downloader-go?***
* **A: Place your config.ini from that program in the same directory as this program and delete any settings.json file if present. The program will import your settings from the old project and make a new settings.json. It will still re-download files that DIDG already downloaded, as the database layout is different and the old database is not imported.**

---

## Development
* I'm a complete amateur with Golang. If anything's bad please make a pull request.
* Versioning is `[MAJOR].[MINOR].[PATCH]`

<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> CREDITS & SOURCES</b></summary>

### Credits & Dependencies
* [github.com/Seklfreak/discord-image-downloader-go - the original project this originated from](https://github.com/Seklfreak/discord-image-downloader-go)

#### Core Dependencies
* [github.com/bwmarrin/discordgo](https://github.com/bwmarrin/discordgo)
* [github.com/Necroforger/dgrouter](https://github.com/Necroforger/dgrouter)
* [github.com/HouzuoGuo/tiedot/db](https://github.com/HouzuoGuo/tiedot)
* [github.com/fatih/color](https://github.com/fatih/color)

#### Other Dependencies
* [github.com/AvraamMavridis/randomcolor](https://github.com/AvraamMavridis/randomcolor)
* [github.com/ChimeraCoder/anaconda](https://github.com/ChimeraCoder/anaconda)
* [github.com/ChimeraCoder/tokenbucket](https://github.com/ChimeraCoder/tokenbucket)
* [github.com/Jeffail/gabs](https://github.com/Jeffail/gabs)
* [github.com/PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery)
* [github.com/azr/backoff](https://github.com/azr/backoff)
* [github.com/dustin/go-jsonpointer](https://github.com/dustin/go-jsonpointer)
* [github.com/dustin/gojson](https://github.com/dustin/gojson)
* [github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify)
* [github.com/garyburd/go-oauth](https://github.com/garyburd/go-oauth)
* [github.com/hako/durafmt](https://github.com/hako/durafmt)
* [github.com/hashicorp/go-version](https://github.com/hashicorp/go-version)
* [github.com/kennygrant/sanitize](https://github.com/kennygrant/sanitize)
* [github.com/nfnt/resize](https://github.com/nfnt/resize)
* [github.com/rivo/duplo](https://github.com/rivo/duplo)
* [golang.org/x/net](https://golang.org/x/net)
* [golang.org/x/oauth2](https://golang.org/x/oauth2)
* [google.golang.org/api](https://google.golang.org/api)
* [gopkg.in/ini.v1](https://gopkg.in/ini.v1)
* [mvdan.cc/xurls/v2](https://mvdan.cc/xurls/v2)
  
</details>
