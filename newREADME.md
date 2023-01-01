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
    <a href="https://github.com/get-got/discord-downloader-go/releases/latest">
        <b>DOWNLOAD LATEST RELEASE</b>
    </a>
    <br/><br/>
    <a href="https://discord.com/invite/6Z6FJZVaDV">
        <b>Need help? Have suggestions? Join the Discord server!</b>
    </a>
</h2><br>

This project is a cross-platform command-line program to interact with a Discord Bot (genuine bot application or user account) to locally download files posted from Discord in real-time as well as a full archive of old messages. It can download any directly sent Discord attachments or linked files and supports fetching highest possible quality files from specific sources _(see list below)._ It also supports **very extensive** global, server-specific, category-specific, channel-specific, and user-specific settings configuration and customization. Tailor the bot to your exact needs. See the [Features](#Features) list below for the full list. Any operating system supported by the latest version of Golang is compatible.

<h3 align="center">
    <b>This project is a fork of <a href="https://github.com/Seklfreak/discord-image-downloader-go">Seklfreak's <i>discord-image-downloader-go</i></a></b>
</h3>
<h4 align="center">
    The original project was abandoned, for a list of differences and why I made an independent project, <a href="#differences-from-seklfreaks-discord-image-downloader-go--why-i-made-this"><b>see below</b></a>
</h4>

---

<h2 align="center">Table of Contents</h2>

<h3 align="center">
    <a href="#Features"><b>- List of Features -</b></a><br>
    <a href="#getting-started"><b>- Getting Started -</b></a><br>
    <a href="#guide-downloading-history-old-messages"><b>- Guide: Downloading History <i>(Old Messages)</i> -</b></a><br>
    <a href="#guide-settings--configuration"><b>- Guide: Settings / Configuration -</b></a><br>
    <a href="#list-of-settings"><b>- List of Settings -</b></a><br>
    <a href="#faq"><b>- FAQ (Frequently Asked Questions) -</b></a><br>
    <a href="#development"><b>- Development, Credits, Dependencies -</b></a><br>
</h3>

---

## Features

<details>
<summary><b><i>(COLLAPSABLE SECTION)</i> LIST OF FEATURES & COMMANDS</b></summary>

### Supported Download Sources
* Discord File Attachments
* Direct Links to Files
* Twitter _(requires API key, see config section)_
* Instagram [BROKEN, WORKING ON IT]
* Reddit [BROKEN, WORKING ON IT]
* Imgur _(Single Posts & Albums)_
* Streamable
* Gfycat
* Tistory
* Mastodon [BROKEN, WORKING ON IT]
* Flickr _(requires API key, see config section)_
* Google Drive _(requires API Credentials, see config section)_
* _I'll always welcome requests but some sources can be tricky to parse..._
  
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
* If using a **User Account (Self-Bot) with 2FA (Two-Factor Authentication),** enter the token into the `"token"` setting. Remove the lines for `"username"` and `"password"` or leave blank (`""`). Token can be found from `Developer Tools` in browser under `localStorage.token` or in the Discord client `Ctrl+Shift+I (Windows)`/`Cmd+Option+I (Mac)` under `Application → Local Storage → https://discordapp.com → "token"`.

### Bot Permissions in Discord...
* In order to perform basic downloading functions, the bot will need `Read Message` permissions in the server(s) of your designated channel(s).
* In order to respond to commands, the bot will need `Send Message` permissions in the server(s) of your designated channel(s). If executing commands via an Admin Channel, the bot will only need `Send Message` permissions for that channel, and that permission will not be required for the source channel.
* In order to process history commands, the bot will need `Read Message History` permissions in the server(s) of your designated channel(s).

#### NOTE: GENUINE DISCORD BOTS REQUIRE PERMISSIONS ENABLED!
* Go to the Discord Application management page, choose your application, go to the `Bot` category, and ensure `Message Content Intent` is enabled.

<img src="https://i.imgur.com/2GcyA2B.png"/>

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
