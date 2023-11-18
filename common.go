package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
)

func uptime() time.Duration {
	return time.Since(startTime) //.Truncate(time.Second)
}

func properExit() {
	// Not formatting string because I only want the exit message to be red.
	log.Println(lg("Main", "", color.HiRedString, "[EXIT IN 15 SECONDS] Uptime was %s...", timeSince(startTime)))
	log.Println(color.HiCyanString("----------------------------------------------------"))
	time.Sleep(15 * time.Second)
	os.Exit(1)
}

func getJSON(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getJSONwithHeaders(url string, target interface{}, headers map[string]string) error {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

//#region Github

type githubReleaseApiObject struct {
	TagName string `json:"tag_name"`
}

var latestGithubRelease string

func getLatestGithubRelease() string {
	githubReleaseApiObject := new(githubReleaseApiObject)
	err := getJSON("https://api.github.com/repos/"+projectRepoBase+"/releases/latest", githubReleaseApiObject)
	if err != nil {
		log.Println(lg("API", "Github", color.RedString, "Error fetching current Release JSON: %s", err))
		return ""
	}
	return githubReleaseApiObject.TagName
}

func isLatestGithubRelease() bool {
	latestGithubRelease = getLatestGithubRelease()
	if latestGithubRelease == "" {
		return true
	}

	thisVersion, err := version.NewVersion(projectVersion)
	if err != nil {
		log.Println(lg("API", "Github", color.RedString, "Error parsing current version: %s", err))
		return true
	}

	latestVersion, err := version.NewVersion(latestGithubRelease)
	if err != nil {
		log.Println(lg("API", "Github", color.RedString, "Error parsing latest version: %s", err))
		return true
	}

	if latestVersion.GreaterThan(thisVersion) {
		return false
	}

	return true
}

//#endregion

//#region Logging

func lg(group string, subgroup string, colorFunc func(string, ...interface{}) string, line string, p ...interface{}) string {
	colorPrefix := group
	switch strings.ToLower(group) {

	case "main":
		if subgroup == "" {
			colorPrefix = ""
		} else {
			colorPrefix = ""
		}

	case "verbose":
		if subgroup == "" {
			colorPrefix = color.HiBlueString("[VERBOSE]")
		} else {
			colorPrefix = color.HiBlueString("[VERBOSE | %s]", subgroup)
		}

	case "debug":
		if subgroup == "" {
			colorPrefix = color.HiYellowString("[DEBUG]")
		} else {
			colorPrefix = color.HiYellowString("[DEBUG | %s]", subgroup)
		}

	case "debug2":
		if subgroup == "" {
			colorPrefix = color.YellowString("[DEBUG2]")
		} else {
			colorPrefix = color.YellowString("[DEBUG2 | %s]", subgroup)
		}

	case "test":
		if subgroup == "" {
			colorPrefix = color.HiYellowString("[TEST]")
		} else {
			colorPrefix = color.HiYellowString("[TEST | %s]", subgroup)
		}

	case "info":
		if subgroup == "" {
			colorPrefix = color.CyanString("[Info]")
		} else {
			colorPrefix = color.CyanString("[Info | %s]", subgroup)
		}

	case "version":
		if subgroup == "" {
			colorPrefix = color.HiGreenString("[Version]")
		} else {
			colorPrefix = color.HiGreenString("[Version | %s]", subgroup)
		}

	case "settings":
		if subgroup == "" {
			colorPrefix = color.GreenString("[Settings]")
		} else {
			colorPrefix = color.GreenString("[Settings | %s]", subgroup)
		}

	case "database":
		if subgroup == "" {
			colorPrefix = color.HiYellowString("[Database]")
		} else {
			colorPrefix = color.HiYellowString("[Database | %s]", subgroup)
		}

	case "setup":
		if subgroup == "" {
			colorPrefix = color.HiGreenString("[Setup]")
		} else {
			colorPrefix = color.HiGreenString("[Setup | %s]", subgroup)
		}

	case "checkup":
		if subgroup == "" {
			colorPrefix = color.HiGreenString("[Checkup]")
		} else {
			colorPrefix = color.HiGreenString("[Checkup | %s]", subgroup)
		}

	case "discord":
		if subgroup == "" {
			colorPrefix = color.HiBlueString("[Discord]")
		} else {
			colorPrefix = color.HiBlueString("[Discord | %s]", subgroup)
		}

	case "history":
		if subgroup == "" {
			colorPrefix = color.HiCyanString("[History]")
		} else {
			colorPrefix = color.HiCyanString("[History | %s]", subgroup)
		}

	case "command":
		if subgroup == "" {
			colorPrefix = color.HiGreenString("[Commands]")
		} else {
			colorPrefix = color.HiGreenString("[Command : %s]", subgroup)
		}

	case "download":
		if subgroup == "" {
			colorPrefix = color.GreenString("[Downloads]")
		} else {
			colorPrefix = color.GreenString("[Downloads | %s]", subgroup)
		}

	case "message":
		if subgroup == "" {
			colorPrefix = color.CyanString("[Messages]")
		} else {
			colorPrefix = color.CyanString("[Messages | %s]", subgroup)
		}

	case "regex":
		if subgroup == "" {
			colorPrefix = color.YellowString("[Regex]")
		} else {
			colorPrefix = color.YellowString("[Regex | %s]", subgroup)
		}

	case "api":
		if subgroup == "" {
			colorPrefix = color.HiMagentaString("[APIs]")
		} else {
			colorPrefix = color.HiMagentaString("[API | %s]", subgroup)
		}
	}

	if bot != nil && botReady {
		simplePrefix := group
		if subgroup != "" {
			simplePrefix += ":" + subgroup
		}
		for _, adminChannel := range config.AdminChannels {
			if *adminChannel.LogProgram {
				outputToChannel := func(channel string) {
					if channel != "" {
						if hasPerms(channel, discordgo.PermissionSendMessages) {
							if _, err := bot.ChannelMessageSend(channel,
								fmt.Sprintf("```%s | [%s] %s```",
									time.Now().Format(time.RFC3339), simplePrefix, fmt.Sprintf(line, p...)),
							); err != nil {
								log.Println(color.HiRedString("Failed to send message...\t%s", err))
							}
						}
					}
				}
				outputToChannel(adminChannel.ChannelID)
				if adminChannel.ChannelIDs != nil {
					for _, ch := range *adminChannel.ChannelIDs {
						outputToChannel(ch)
					}
				}
			}
		}
	}

	pp := "> " // prefix prefix :)
	if strings.ToLower(group) == "debug" || strings.ToLower(subgroup) == "debug" ||
		strings.ToLower(group) == "debug2" || strings.ToLower(subgroup) == "debug2" {
		pp = color.YellowString("? ")
	}
	if strings.ToLower(group) == "verbose" || strings.ToLower(subgroup) == "verbose" {
		pp = color.HiBlueString("? ")
	}

	if colorPrefix != "" {
		colorPrefix += " "
	}
	tabPrefix := ""
	if config.LogIndent {
		tabPrefix = "\t"
	}
	return tabPrefix + pp + colorPrefix + colorFunc(line, p...)
}

//#endregion
