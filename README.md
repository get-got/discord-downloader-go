<div align="center">
    <img src="https://i.imgur.com/8KSripJ.png" alt="Discord Downloader Go"/>
    <p>Maintained since late 2020</p>
</div>
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
</h2><br>

<div align="center">

This project is a cross-platform cli executable program to interact with a Discord Bot (genuine bot application or user account, limitations apply to both respectively) to locally download files posted from Discord in real-time as well as a full archive of old messages. It can download any directly sent Discord attachments or linked files and supports fetching highest possible quality files from specific sources _([see list below](#supported-download-sources))._

It also supports **very extensive** settings configurations and customization, applicable globally or per-server/category/channel/user. Tailor the bot to your exact needs and runtime environment.

<h3 align="center">
    <b>Originally a fork of <a href="https://github.com/Seklfreak/discord-image-downloader-go">Seklfreak's <i>discord-image-downloader-go</i></a></b>
</h3>
<h4 align="center">
    The original project was abandoned, for a list of differences and why I made an independent project, <a href="#differences-from-seklfreaks-discord-image-downloader-go--why-i-made-this"><b>see below</b></a>
</h4>

---

<h3><a href="https://github.com/get-got/discord-downloader-go/wiki/Home">SEE THE WIKI FOR MORE</a></h3>

<h3>Wiki - <a href="https://github.com/get-got/discord-downloader-go/wiki/Getting-Started">Getting Started</a> </h3>

<h3>Wiki - <a href="https://github.com/get-got/discord-downloader-go/wiki/Settings">Settings</a></h3>

<h3>Wiki - <a href="https://github.com/get-got/discord-downloader-go/wiki/Settings-Examples">Settings Examples</a></h3>

<h3>Wiki - <a href="https://github.com/get-got/discord-downloader-go/wiki/Guide-%E2%80%90-Commands">Guide - Commands</a></h3>

<h3>Wiki - <a href="https://github.com/get-got/discord-downloader-go/wiki/Guide-%E2%80%90-History-(Old-Messages)">Guide - History (Old Messages)</a></h3>

| Operating System  | Architectures _( ? = available but untested )_    |
| -----------------:|:----------------------------------------------- |
| Windows           | **amd64**, arm64 _(?)_, armv7/6/5 _(?)_, 386 _(?)_
| Linux             | **amd64**, **arm64**, **armv7/6/5**,<br/>risc-v64 _(?)_, mips64/64le _(?)_, s390x _(?)_, 386 _(?)_
| Darwin (Mac)      | **amd64**, arm64 _(?)_
| FreeBSD           | amd64 _(?)_, arm64 _(?)_, armv7/6/5 _(?)_, 386 _(?)_
| OpenBSD           | amd64 _(?)_, arm64 _(?)_, armv7/6/5 _(?)_, 386 _(?)_
| NetBSD            | amd64 _(?)_, arm64 _(?)_, armv7/6/5 _(?)_, 386 _(?)_

</div>

---

## ⚠️ **WARNING!** Discord does not allow Automated User Accounts (Self-Bots/User-Bots)

[Read more in Discord Trust & Safety Team's Official Statement...](https://support.discordapp.com/hc/en-us/articles/115002192352-Automated-user-accounts-self-bots-)

While this project works for user logins, I do not recommend it as you risk account termination. If you can, [use a proper Discord Bot user for this program.](https://discord.com/developers/applications)

> _NOTE: This only applies to real User Accounts, not Bot users. This program currently works for either._

---

## 🤖 Features

### Supported Download Sources

- Direct Links to Files
- Discord File Attachments
- Twitter / X _(requires account login, see config section)_
- Instagram _(requires account login, see config section)_
- Reddit
- Imgur
- Streamable
- Gfycat
- Tistory
- Flickr _(requires API key, see config section)_
- _I'll always welcome requests but some sources can be tricky to parse..._

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

## ❔ FAQ

- **_Q: How do I install?_**
- **A: [SEE #getting-started](https://github.com/get-got/discord-downloader-go/wiki/Getting-Started)**

---

- **_Q: How do I convert from Seklfreak's discord-image-downloader-go?_**
- **A: Place your config.ini from that program in the same directory as this program and delete any settings.json file if present. The program will import your settings from the old project and make a new settings.json. It will still re-download files that DIDG already downloaded, as the database layout is different and the old database is not imported.**

---

## ⚙️ Development

- I'm a complete amateur with Golang. If anything's bad please make a pull request.
- Follows Semantic Versioning: `[MAJOR].[MINOR].[PATCH]` <https://semver.org/>
- [github.com/Seklfreak/discord-image-downloader-go - the original project this was founded on](https://github.com/Seklfreak/discord-image-downloader-go)
