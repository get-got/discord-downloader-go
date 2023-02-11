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
    <a href="https://hub.docker.com/r/getgot/discord-downloader-go" alt="Docker Pulls">
        <img src="https://img.shields.io/docker/pulls/getgot/discord-downloader-go?label=docker-pulls&logo=Docker" />
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
    <a href="https://discord.com/invite/6Z6FJZVaDV">
        <b>Need help? Have suggestions? Join the Discord server!</b>
    </a>
    <br/><br/>
    <a href="https://github.com/get-got/discord-downloader-go/releases/latest">
        <b>DOWNLOAD LATEST RELEASE</b>
    </a>
</h2>
<div align="center">

| Operating System  | Architectures _( ? = available but untested )_    |
| -----------------:|:----------------------------------------------- |
| Windows           | **amd64**, arm64 _(?)_, armv7/6/5 _(?)_, 386 _(?)_
| Linux             | **amd64**, **arm64**, **armv7/6/5**,<br/>risc-v64 _(?)_, mips64/64le _(?)_, s390x _(?)_, 386 _(?)_
| Darwin (Mac)      | amd64 _(?)_, arm64 _(?)_
| FreeBSD           | amd64 _(?)_, arm64 _(?)_, armv7/6/5 _(?)_, 386 _(?)_
| OpenBSD           | amd64 _(?)_, arm64 _(?)_, armv7/6/5 _(?)_, 386 _(?)_
| NetBSD            | amd64 _(?)_, arm64 _(?)_, armv7/6/5 _(?)_, 386 _(?)_

</div><br>

This project is a cross-platform cli single-file program to interact with a Discord Bot (genuine bot application or user account, limitations apply to both respectively) to locally download files posted from Discord in real-time as well as a full archive of old messages. It can download any directly sent Discord attachments or linked files and supports fetching highest possible quality files from specific sources _([see list below](#supported-download-sources))._ It also supports **very extensive** settings configurations and customization, applicable globally or per-server/category/channel/user. Tailor the bot to your exact needs and runtime environment. See the [Features](#-features) list below for the full list. See the [List of Settings](#-list-of-settings) below for a settings breakdown. See [Getting Started](#%EF%B8%8F-getting-started) or anything else in the table of contents right under this to learn more!

<h3 align="center">
    <b>Originally a fork of <a href="https://github.com/Seklfreak/discord-image-downloader-go">Seklfreak's <i>discord-image-downloader-go</i></a></b>
</h3>
<h4 align="center">
    The original project was abandoned, for a list of differences and why I made an independent project, <a href="#differences-from-seklfreaks-discord-image-downloader-go--why-i-made-this"><b>see below</b></a>
</h4>

---

- [⚠️ **WARNING!** Discord does not allow Automated User Accounts (Self-Bots/User-Bots)](#️-warning-discord-does-not-allow-automated-user-accounts-self-botsuser-bots)
- [🤖 Features](#-features)
  - [Supported Download Sources](#supported-download-sources)
  - [Commands](#commands)
- [✔️ Getting Started](#️-getting-started)
  - [Getting Started Step-by-Step](#getting-started-step-by-step)
  - [Bot Login Credentials](#bot-login-credentials)
  - [Bot Permissions in Discord](#bot-permissions-in-discord)
    - [NOTE: GENUINE DISCORD BOTS REQUIRE PERMISSIONS ENABLED](#note-genuine-discord-bots-require-permissions-enabled)
  - [How to Find Discord IDs](#how-to-find-discord-ids)
  - [Differences from Seklfreak's _discord-image-downloader-go_ \& Why I made this](#differences-from-seklfreaks-discord-image-downloader-go--why-i-made-this)
- [📚 Guide: Downloading History (Old Messages)](#-guide-downloading-history-old-messages)
  - [Command Arguments](#command-arguments)
    - [Examples](#examples)
- [🔨 Guide: Settings / Configuration](#-guide-settings--configuration)
- [🛠 List of Settings](#-list-of-settings)
      - [\_constants](#_constants)
      - [credentials](#credentials)
        - [token](#token)
        - [email](#email)
        - [password](#password)
        - [twitterAccessToken](#twitteraccesstoken)
        - [twitterAccessTokenSecret](#twitteraccesstokensecret)
        - [twitterConsumerKey](#twitterconsumerkey)
        - [twitterConsumerSecret](#twitterconsumersecret)
        - [instagramUsername](#instagramusername)
        - [instagramPassword](#instagrampassword)
        - [flickrApiKey](#flickrapikey)
      - [admins](#admins)
      - [adminChannels](#adminchannels)
      - [processLimit](#processlimit)
      - [debug](#debug)
      - [settingsOutput](#settingsoutput)
      - [messageOutput](#messageoutput)
      - [messageOutputHistory](#messageoutputhistory)
      - [discordLogLevel](#discordloglevel)
      - [discordTimeout](#discordtimeout)
      - [\_constants](#_constants-1)
      - [\_constants](#_constants-2)
      - [\_constants](#_constants-3)
      - [\_constants](#_constants-4)
      - [\_constants](#_constants-5)
      - [\_constants](#_constants-6)
      - [\_constants](#_constants-7)
      - [\_constants](#_constants-8)
      - [\_constants](#_constants-9)
      - [\_constants](#_constants-10)
      - [\_constants](#_constants-11)
      - [\_constants](#_constants-12)
      - [\_constants](#_constants-13)
      - [\_constants](#_constants-14)
      - [\_constants](#_constants-15)
      - [\_constants](#_constants-16)
      - [\_constants](#_constants-17)
      - [\_constants](#_constants-18)
      - [\_constants](#_constants-19)
      - [\_constants](#_constants-20)
      - [\_constants](#_constants-21)
      - [\_constants](#_constants-22)
      - [\_constants](#_constants-23)
      - [\_constants](#_constants-24)
      - [\_constants](#_constants-25)
      - [\_constants](#_constants-26)
      - [\_constants](#_constants-27)
      - [\_constants](#_constants-28)
      - [\_constants](#_constants-29)
      - [\_constants](#_constants-30)
      - [\_constants](#_constants-31)
      - [\_constants](#_constants-32)
      - [\_constants](#_constants-33)
      - [\_constants](#_constants-34)
      - [\_constants](#_constants-35)
      - [\_constants](#_constants-36)
      - [\_constants](#_constants-37)
      - [\_constants](#_constants-38)
      - [\_constants](#_constants-39)
      - [\_constants](#_constants-40)
      - [\_constants](#_constants-41)
      - [\_constants](#_constants-42)
      - [\_constants](#_constants-43)
      - [\_constants](#_constants-44)
      - [\_constants](#_constants-45)
      - [\_constants](#_constants-46)
      - [\_constants](#_constants-47)
      - [\_constants](#_constants-48)
      - [\_constants](#_constants-49)
      - [\_constants](#_constants-50)
      - [\_constants](#_constants-51)
      - [\_constants](#_constants-52)
      - [\_constants](#_constants-53)
      - [\_constants](#_constants-54)
      - [\_constants](#_constants-55)
      - [\_constants](#_constants-56)
      - [\_constants](#_constants-57)
      - [\_constants](#_constants-58)
      - [\_constants](#_constants-59)
      - [\_constants](#_constants-60)
      - [\_constants](#_constants-61)
      - [\_constants](#_constants-62)
      - [\_constants](#_constants-63)
      - [\_constants](#_constants-64)
      - [\_constants](#_constants-65)
      - [\_constants](#_constants-66)
      - [\_constants](#_constants-67)
      - [\_constants](#_constants-68)
      - [\_constants](#_constants-69)
  - [settings - main](#settings---main)
  - [settings - credentials group](#settings---credentials-group)
  - [settings - adminChannels group](#settings---adminchannels-group)
  - [settings - source group](#settings---source-group)
  - [settings - source / filters group](#settings---source--filters-group)
  - [settings - source / log group](#settings---source--log-group)
- [🤖 Settings Examples](#-settings-examples)
  - [example - minimum bot app](#example---minimum-bot-app)
  - [example - minimum user account without 2FA](#example---minimum-user-account-without-2fa)
  - [example - minimum user account with 2FA](#example---minimum-user-account-with-2fa)
  - [example - server with friends](#example---server-with-friends)
  - [example - scraping public servers (user)](#example---scraping-public-servers-user)
  - [example - scraping public server (bot app, as admin)](#example---scraping-public-server-bot-app-as-admin)
- [❔ FAQ](#-faq)
- [⚙️ Development](#️-development)

---

## ⚠️ **WARNING!** Discord does not allow Automated User Accounts (Self-Bots/User-Bots)

[Read more in Discord Trust & Safety Team's Official Statement...](https://support.discordapp.com/hc/en-us/articles/115002192352-Automated-user-accounts-self-bots-)

While this project works for user logins, I do not reccomend it as you risk account termination. If you can, [use a proper Discord Bot user for this program.](https://discord.com/developers/applications)

> _NOTE: This only applies to real User Accounts, not Bot users. This program currently works for either._

Now that that's out of the way...

---

## 🤖 Features

### Supported Download Sources

- Direct Links to Files
- Discord File Attachments
- Twitter _(requires API key, see config section)_
- Instagram _(requires account login, see config section)_
- Reddit
- Imgur
- Streamable
- Gfycat
- Tistory
- Flickr _(requires API key, see config section)_
- _I'll always welcome requests but some sources can be tricky to parse..._
  
### Commands

Commands are used as `ddg <command> <?arguments?>` _(unless you've changed the prefix)_
Command     | Arguments? | Description
---         | ---   | ---
`help`, `commands`  | No    | Lists all commands.
`ping`, `test`      | No    | Pings the bot.
`info`      | No    | Displays relevant Discord info.
`status`    | No    | Shows the status of the bot.
`stats`     | No    | Shows channel stats.
`history`   | [**SEE HISTORY SECTION**](#-guide-downloading-history-old-messages) | **(BOT AND SERVER ADMINS ONLY)** Processes history for old messages in channel.
`exit`, `kill`, `reload`    | No    | **(BOT ADMINS ONLY)** Exits the bot _(or restarts if using a keep-alive process manager)_.
`emojis`    | Optionally specify server IDs to download emojis from; separate by commas | **(BOT ADMINS ONLY)** Saves all emojis for channel.

---

## ✔️ Getting Started

_Confused? Try looking at [the step-by-step list.](#getting-started-step-by-step)_

Depending on your purpose for this program, there are various ways you can run it.

- [Run the executable file for your platform. _(Process managers like **pm2** work well for this)_](https://github.com/get-got/discord-downloader-go/releases/latest)
- [Run the executable file via command prompt. _(`discord-downloader-go.exe settings2` or similar to run multiple instances sharing a database with separate settings files)_](https://github.com/get-got/discord-downloader-go/releases/latest)
- [Run automated image builds in Docker.](https://hub.docker.com/r/getgot/discord-downloader-go) _(Google it)._
  - Mount your settings.json to ``/root/settings.json``
  - Mount a folder named "database" to ``/root/database``
  - Mount your save folders or the parent of your save folders within ``/root/``
    - _i.e. ``X:\My Folder`` to ``/root/My Folder``_
- Install Golang and compile/run the source code yourself. _(Google it)_

You can either create a `settings.json` following the examples & variables listed below, or have the program create a default file (if it is missing when you run the program, it will make one, and ask you if you want to enter in basic info for the new file).

- [Ensure you follow proper JSON syntax to avoid any unexpected errors.](https://www.w3schools.com/js/js_json_syntax.asp)
- [Having issues? Try this JSON Validator to ensure it's correctly formatted.](https://jsonformatter.curiousconcept.com/)

[![Tutorial Video](http://img.youtube.com/vi/06UUXDQ80f8/0.jpg)](http://www.youtube.com/watch?v=06UUXDQ80f8)

### Getting Started Step-by-Step

1. Download & put executable within it's own folder.
2. Configure Main Settings (or run once to have settings generated). [_(SEE BELOW)_](#-list-of-settings)
3. Enter your login credentials in the `"credentials"` section. [_(SEE BELOW)_](#-list-of-settings)
4. Put your Discord User ID as in the `"admins"` list of the settings. [_(SEE BELOW)_](#-list-of-settings)
5. Put a Discord Channel ID for a private channel you have access to into the `"adminChannels"`. [_(SEE BELOW)_](#-list-of-settings)
6. Put your desired Discord Channel IDs into th6e `"channels"` section. [_(SEE BELOW)_](#-list-of-settings)
   - I know it can be confusing if you don't have experience with programming or JSON in general, but this was the ideal setup for extensive configuration like this. Just be careful with comma & quote placement and you should be fine. [See examples below for help.](#-settings-examples)

### Bot Login Credentials

- If using a **Bot Application,** enter the token into the `"token"` setting. Remove the lines for `"username"` and `"password"` or leave blank (`""`). **To create a Bot User,** go to [discord.com/developers/applications](https://discord.com/developers/applications) and create a `New Application`. Once created, go to `Bot` and create. The token can be found on the `Bot` page. To invite to your server(s), go to `OAuth2` and check `"bot"`, copy the url, paste into browser and follow prompts for adding to server(s).
- If using a **User Account (Self-Bot) WITHOUT 2FA (2-Factor Authentication),** fill out the `"username"` and `"password"` settings. Remove the line for `"token"` or leave blank (`""`).
- If using a **User Account (Self-Bot) WITH 2FA (2-Factor Authentication),** enter the token into the `"token"` setting. Remove the lines for `"username"` and `"password"` or leave blank (`""`). Your account token can be found by `opening the browser console / dev tools / inspect element` > `Network` tab > `filter for "library"` and reload the page if nothing appears. Assuming there is an item that looks like the below screenshot, click it and `find "Authorization" within the "Request Headers" section` of the Headers tab. The random text is your token.

<img src="https://i.imgur.com/2BdaJSH.png"> <img src="https://i.imgur.com/i9DItcH.png">

### Bot Permissions in Discord

- In order to perform basic downloading functions, the bot will need `Read Message` permissions in the server(s) of your designated channel(s).
- In order to respond to commands, the bot will need `Send Message` permissions in the server(s) of your designated channel(s). If executing commands via an Admin Channel, the bot will only need `Send Message` permissions for that channel, and that permission will not be required for the source channel.
- In order to process history commands, the bot will need `Read Message History` permissions in the server(s) of your designated channel(s).

#### NOTE: GENUINE DISCORD BOTS REQUIRE PERMISSIONS ENABLED

- Go to the Discord Application management page, choose your application, go to the `Bot` category, and ensure `Message Content Intent` is enabled.

<img src="https://i.imgur.com/2GcyA2B.png"/>

### How to Find Discord IDs

- **_Use the info command!_**
- **Discord Developer Mode:** Enable `Developer Mode` in Discord settings under `Appearance`.
- **Finding Channel ID:** _Enable Discord Developer Mode (see above),_ right click on the channel and `Copy ID`.
- **Finding User ID:** _Enable Discord Developer Mode (see above),_ right click on the user and `Copy ID`.
- **Finding Emoji ID:** _Enable Discord Developer Mode (see above),_ right click on the emoji and `Copy ID`.
- **Finding DM/PM ID:** Inspect Element on the DM icon for the desired user. Look for `href="/channels/@me/CHANNEL_ID_HERE"`. Using this ID in place of a normal channel ID should work perfectly fine.

---

### Differences from [Seklfreak's _discord-image-downloader-go_](https://github.com/Seklfreak/discord-image-downloader-go) & Why I made this

- _Better command formatting & support_
- Configuration is JSON-based rather than ini to allow more elaborate settings and better organization. With this came many features such as channel-specific settings.
- Channel-specific control of downloaded filetypes / content types (considers things like .mov as videos as well, rather than ignore them), Optional dividing of content types into separate folders.
- (Optional) Reactions upon download success.
- (Optional) Discord messages upon encountered errors.
- Extensive bot status/presence customization.
- Consistent Log Formatting, Color-Coded Logging
- Somewhat different organization than original project; initially created from scratch then components ported over.
- _Various fixes, improvements, and dependency updates that I also contributed to Seklfreak's original project._

> I've been a user of Seklfreak's project since ~2018 and it's been great for my uses, but there were certain aspects I wanted to expand upon, one of those being customization of channel configuration, and other features like message reactions upon success, differently formatted statuses, etc. If some aspects are rudimentary or messy, please make a pull request, as this is my first project using Go and I've learned everything from observation & Stack Overflow.

---

## 📚 Guide: Downloading History (Old Messages)

> This guide is to show you how to make the bot go through all old messages in a channel and catalog them as though they were being sent right now, in order to download them all.

### Command Arguments

If no channel IDs are specified, it will try and use the channel ID for the channel you're using the command in.

Argument / Flag         | Details
---                     | ---
**channel ID(s)**       | One or more channel IDs, separated by commas if multiple.
`all`                   | Use all available registered channels.
`cancel` or `stop`      | Stop downloading history for specified channel(s).
`list` or `status`      | Output running history jobs in Discord & program.
`--since=YYYY-MM-DD`    | Will process messages sent after this date.
`--since=message_id`    | Will process messages sent after this message.
`--before=YYYY-MM-DD`   | Will process messages sent before this date.
`--before=message_id`   | Will process messages sent before this message.

**_Order of arguments does not matter_**

#### Examples

- `ddg history`
- `ddg history cancel`
- `ddg history all`
- `ddg history stop all`
- `ddg history 000111000111000`
- `ddg history 000111000111000, 000222000222000`
- `ddg history 000111000111000,000222000222000,000333000333000`
- `ddg history 000111000111000, 000333000333000 cancel`
- `ddg history 000111000111000 --before=000555000555000`
- `ddg history 000111000111000 --since=2020-01-02`
- `ddg history 000111000111000 --since=2020-10-12 --before=2021-05-06`
- `ddg history 000111000111000 --since=000555000555000 --before=2021-05-06`
- `ddg history status`
- `ddg history list`

---

## 🔨 Guide: Settings / Configuration

> I tried to make the configuration as user friendly as possible, though you still need to follow proper JSON syntax **(watch those commas)**. Most settings are optional and will use default values or be unused if missing from your settings file.

When initially launching the bot it will create a default settings file if you do not create your own `settings.json` manually. All JSON settings follow camelCase format.

**If you have a ``config.ini`` from _Seklfreak's discord-image-downloader-go_, it will import settings if it's in the same folder as the program.**

The bot accepts `.json` or `.jsonc` for comment-friendly json.

---

## 🛠 List of Settings

TODO: UNDER CONSTRUCTION

THIS IS THE MAIN SETTINGS GROUP, ALL OF THIS WOULD GO INSIDE THE MAIN `{ }` FILE BRACKETS.

- [_constants](#_constants) `{ "key": "value", "key2": "value2" }`
- [credentials](#credentials)
  - [token](#token) `string ; genuine bot or 2FA user`
  - [email](#email) `string ; non-2FA user`
  - [password](#password) `string ; non-2FA user`
  - [twitterAccessToken](#twitteraccesstoken) `string`
  - [twitterAccessTokenSecret](#twitteraccesstokensecret) `string`
  - [twitterConsumerKey](#twitterconsumerkey) `string`
  - [twitterConsumerSecret](#twitterconsumersecret) `string`
  - [instagramUsername](#instagramusername) `string`
  - [instagramPassword](#instagrampassword) `string`
  - [flickrApiKey](#flickrapikey) `string`
- [admins](#admins) `[ ] of strings`
- [adminChannels](#adminchannels)  `[ ] of adminChannels`
- [processLimit](#processlimit) `int ; 3`
- debug `bool ; false`
- settingsOutput `bool ; true`
- messageOutput `bool ; true`
- messageOutputHistory `bool ; false`
- discordLogLevel `int ; 0`
- discordTimeout `int ; 180`
- downloadTimeout `int ; 60`
- downloadRetryMax `int ; 3`
- exitOnBadConnection `bool ; false`
- githubUpdateChecking `bool ; true`
- commandPrefix `string ; "ddg "`
- scanOwnMessages `bool ; false`
- allowGeneralCommands `bool ; true`
- inflateDownloadCount `int`
- europeanNumbers `bool ; false`
- checkupRate `int ; 30`
- connectionCheckRate `int ; 5`
- presenceRefreshRate `int ; 3`
- save `bool ; true`
- allowCommands `bool ; true`
- scanEdits `bool ; true`
- ignoreBots `bool ; true`
- sendErrorMessages `bool ; true`
- sendFileToChannel `string`
- sendFileToChannels `[ ] of strings`
- sendFileDirectly `bool`
- sendFileCaption `string`
- filenameDateFormat `string ; "2006-01-02_15-04-05"`
- filenameFormat `string ; "{{date}} {{shortID}} {{file}}"`
- presenceEnabled `bool ; true`
- presenceStatus `string ; "idle"`
- presenceType `int ; 0`
- presenceLabel `string`
- presenceDetails `string`
- presenceState `string`
- reactWhenDownloaded `bool ; true`
- reactWhenDownloadedEmoji `string`
- reactWhenDownloadedHistory `bool ; false`
- historyTyping `bool ; true`
- embedColor `string`
- historyMaxJobs `int ; 3`
- autoHistory `bool ; false`
- autoHistoryBefore `string`
- autoHistorySince `string`
- sendAutoHistoryStatus `bool ; false`
- sendHistoryStatus `bool ; true`
- divideByYear `bool ; false`
- divideByMonth `bool ; false`
- divideByServer `bool ; false`
- divideByChannel `bool ; false`
- divideByUser `bool ; false`
- divideByType `bool ; true`
- divideFoldersUseID `bool ; false`
- saveImages `bool ; true`
- saveVideos `bool ; true`
- saveAudioFiles `bool ; true`
- saveTextFiles `bool ; false`
- saveOtherFiles `bool ; false`
- savePossibleDuplicates `bool ; false`
- filters
  - blockedPhrases `[ ] of strings`
  - allowedPhrases `[ ] of strings`
  - blockedUsers `[ ] of strings`
  - allowedUsers `[ ] of strings`
  - blockedRoles `[ ] of strings`
  - allowedRoles `[ ] of strings`
  - blockedExtensions `[ ] of strings ; common misc files by default`
  - allowedExtensions `[ ] of strings`
  - blockedDomains `[ ] of strings`
  - allowedDomains `[ ] of strings`
- **all**
  - _... follow **channels** below ..._
- allBlacklistUsers `[ ] of strings`
- allBlacklistServers `[ ] of strings`
- allBlacklistCategories `[ ] of strings`
- allBlacklistChannels `[ ] of strings`
- **users**
  - user `string`
  - users `[ ] of strings`
  - _... all other options follow **channels** below ..._
- **servers**
  - server `string`
  - servers `[ ] of strings`
  - serverBlacklist `[ ] of strings`
  - _... all other options follow **channels** below ..._
- **categories**
  - category `string`
  - categories `[ ] of strings`
  - categoryBlacklist `[ ] of strings`
  - _... all other options follow **channels** below ..._
- **channels**
  - channel `string`
  - channels `[ ] of strings`
  - destination `string`
  - enabled `bool ; true`
  - save `bool ; true`
  - allowCommands `bool ; true`
  - scanEdits `bool ; true`
  - sendErrorMessages `bool ; true`
  - sendFileToChannel `string`
  - sendFileToChannels `[ ] of strings`
  - sendFileDirectly `bool ; true`
  - sendFileCaption `string`
  - filenameDateFormat `string ; "2006-01-02_15-04-05"`
  - filenameFormat `string ; "{{date}} {{shortID}} {{file}}"`
  - presenceEnabled `bool ; true`
  - reactWhenDownloaded `bool ; true`
  - reactWhenDownloadedEmoji `string`
  - reactWhenDownloadedHistory `bool ; false`
  - blacklistReactEmojis `[ ] of strings`
  - historyTyping `bool ; true`
  - embedColor `string`
  - autoHistory `bool ; false`
  - autoHistoryBefore `string`
  - autoHistorySince `string`
  - sendAutoHistoryStatus `bool ; false`
  - sendHistoryStatus `bool ; true`
  - divideByYear `bool ; false`
  - divideByMonth `bool ; false`
  - divideByServer `bool ; false`
  - divideByChannel `bool ; false`
  - divideByUser `bool ; false`
  - divideByType `bool ; true`
  - divideFoldersUseID `bool ; false`
  - saveImages `bool ; true`
  - saveVideos `bool ; true`
  - saveAudioFiles `bool ; true`
  - saveTextFiles `bool ; false`
  - saveOtherFiles `bool ; false`
  - savePossibleDuplicates `bool ; false`
  - filters
    - blockedPhrases `[ ] of strings`
    - allowedPhrases `[ ] of strings`
    - blockedUsers `[ ] of strings`
    - allowedUsers `[ ] of strings`
    - blockedRoles `[ ] of strings`
    - allowedRoles `[ ] of strings`
    - blockedExtensions `[ ] of strings ; common misc files by default`
    - allowedExtensions `[ ] of strings`
    - blockedDomains `[ ] of strings`
    - allowedDomains `[ ] of strings`
  - logLinks
  - logMessages

##### _constants

<sup>LOREM IPSUM</sup>

##### credentials

<sup>LOREM IPSUM</sup>

###### token

<sup>LOREM IPSUM</sup>

###### email

<sup>LOREM IPSUM</sup>

###### password

<sup>LOREM IPSUM</sup>

###### twitterAccessToken

<sup>LOREM IPSUM</sup>

###### twitterAccessTokenSecret

<sup>LOREM IPSUM</sup>

###### twitterConsumerKey

<sup>LOREM IPSUM</sup>

###### twitterConsumerSecret

<sup>LOREM IPSUM</sup>

###### instagramUsername

<sup>LOREM IPSUM</sup>

###### instagramPassword

<sup>LOREM IPSUM</sup>

###### flickrApiKey

<sup>LOREM IPSUM</sup>

##### admins

<sup>LOREM IPSUM</sup>

##### adminChannels

<sup>LOREM IPSUM</sup>

##### processLimit

<sup>LOREM IPSUM</sup>

##### debug

<sup>LOREM IPSUM</sup>

##### settingsOutput

<sup>LOREM IPSUM</sup>

##### messageOutput

<sup>LOREM IPSUM</sup>

##### messageOutputHistory

<sup>LOREM IPSUM</sup>

##### discordLogLevel

<sup>LOREM IPSUM</sup>

##### discordTimeout

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

##### _constants

<sup>LOREM IPSUM</sup>

### settings - main

| SETTING KEY         | TYPE                                    | DEFAULT    | DESCRIPTION                                                          | EXAMPLE                           |
| :-----------------: | --------------------------------------- | :--------: | -------------------------------------------------------------------- | --------------------------------- |
| credentials         | `credentials group`                     |            | See `credentials group` below.                                       |                                   |
| admins              | array of <br/>strings                   | None       | Discord IDs of users<br/> to use admin commands.                     | `"admins": [ "0", "0" ],`         |
| adminChannels       | array of <br/>`adminChannel groups`     | None       | See `adminChannel group` below.                                      |                                   |
| discordLogLevel     | int (whole number)                      | 0 (errors) | 0 = Errors, <br/>1 = Warning, <br/>2 = Informational, <br/>3 = Debug | `"discordLogLevel": 2,`           |
| debug         | boolean <br/>(true or false)            | false      | Enables extra output for narrowing down problems.                    | `"debug": true,`            |
| messageOutput       | boolean <br/>(true or false)            | true       | Enables discord message output.                                      | `"messageOutput": true,`          |

### settings - credentials group

### settings - adminChannels group

### settings - source group

### settings - source / filters group

### settings - source / log group

---

## 🤖 Settings Examples

TODO: UNDER CONSTRUCTION

### example - minimum bot app

### example - minimum user account without 2FA

### example - minimum user account with 2FA

### example - server with friends

### example - scraping public servers (user)

### example - scraping public server (bot app, as admin)

---

## ❔ FAQ

- **_Q: How do I install?_**
- **A: [SEE #getting-started](#%EF%B8%8F-getting-started)**

---

- **_Q: How do I convert from Seklfreak's discord-image-downloader-go?_**
- **A: Place your config.ini from that program in the same directory as this program and delete any settings.json file if present. The program will import your settings from the old project and make a new settings.json. It will still re-download files that DIDG already downloaded, as the database layout is different and the old database is not imported.**

---

## ⚙️ Development

- I'm a complete amateur with Golang. If anything's bad please make a pull request.
- Follows Semantic Versioning: `[MAJOR].[MINOR].[PATCH]` <https://semver.org/>
- [github.com/Seklfreak/discord-image-downloader-go - the original project this was founded on](https://github.com/Seklfreak/discord-image-downloader-go)
