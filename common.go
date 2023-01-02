package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/hashicorp/go-version"
)

//#region Instance

func uptime() time.Duration {
	return time.Since(startTime)
}

func properExit() {
	// Not formatting string because I only want the exit message to be red.
	log.Println(lg("Main", "", color.HiRedString, "[EXIT IN 15 SECONDS] Uptime was %s...", durafmt.Parse(time.Since(startTime)).String()))
	log.Println(color.HiCyanString("--------------------------------------------------------------------------------"))
	time.Sleep(15 * time.Second)
	os.Exit(1)
}

//#endregion

//#region Files

var (
	pathBlacklist = []string{"/", "\\", "<", ">", ":", "\"", "|", "?", "*"}
)

func clearPath(p string) string {
	r := p
	for _, key := range pathBlacklist {
		r = strings.ReplaceAll(r, key, "")
	}
	return r
}

func filenameFromURL(inputURL string) string {
	base := path.Base(inputURL)
	parts := strings.Split(base, "?")
	return path.Clean(parts[0])
}

func filepathExtension(filepath string) string {
	if strings.Contains(filepath, "?") {
		filepath = strings.Split(filepath, "?")[0]
	}
	filepath = path.Ext(filepath)
	return filepath
}

//#endregion

//#region Text Formatting & Querying

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.EqualFold(a, b) {
			return true
		}
	}
	return false
}

func formatNumber(n int64) string {
	var numberSeparator byte = ','
	if config.NumberFormatEuropean {
		numberSeparator = '.'
	}

	in := strconv.FormatInt(n, 10)
	out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)
	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = numberSeparator
		}
	}
}

func formatNumberShort(x int64) string {
	var numberSeparator string = ","
	if config.NumberFormatEuropean {
		numberSeparator = "."
	}
	var decimalSeparator string = "."
	if config.NumberFormatEuropean {
		decimalSeparator = ","
	}

	if x > 1000 {
		formattedNumber := formatNumber(x)
		splitSlice := strings.Split(formattedNumber, numberSeparator)
		suffixes := [4]string{"k", "m", "b", "t"}
		partCount := len(splitSlice) - 1
		var output string
		if splitSlice[1][:1] != "0" {
			output = fmt.Sprintf("%s%s%s%s", splitSlice[0], decimalSeparator, splitSlice[1][:1], suffixes[partCount-1])
		} else {
			output = fmt.Sprintf("%s%s", splitSlice[0], suffixes[partCount-1])
		}
		return output
	}
	return fmt.Sprint(x)
}

func pluralS(num int) string {
	if num == 1 {
		return ""
	}
	return "s"
}

func wrapHyphens(i string, l int) string {
	n := i
	if len(n) < l {
		n = "- " + n + " -"
		for len(n) < l {
			n = "-" + n + "-"
		}
	}
	return n
}

func wrapHyphensW(i string) string {
	return wrapHyphens(i, 80)
}

func stripSymbols(i string) string {
	re, err := regexp.Compile(`[^\w]`)
	if err != nil {
		log.Fatal(err)
	}
	return re.ReplaceAllString(i, " ")
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func dateLocalToUTC(s string) string {
	if s == "" || !isDate(s) {
		return ""
	}
	rawDate, _ := time.Parse("2006-01-02", s)
	localDate := time.Date(rawDate.Year(), rawDate.Month(), rawDate.Day(), 0, 0, 0, 0, time.Local)
	return fmt.Sprintf("%04d-%02d-%02d", localDate.In(time.UTC).Year(), localDate.In(time.UTC).Month(), localDate.In(time.UTC).Day())
}

func condenseString(input string, length int) string {
	filler := "....."
	ret := input
	if len(input) > length+len(filler) {
		half := int((length / 2) - len(filler))
		ret = input[0:half] + filler + input[len(input)-half:]
	}
	return ret
}

//#endregion

//#region Github Release Checking

type githubReleaseApiObject struct {
	TagName string `json:"tag_name"`
}

func isLatestGithubRelease() bool {
	githubReleaseApiObject := new(githubReleaseApiObject)
	err := getJSON(projectReleaseApiURL, githubReleaseApiObject)
	if err != nil {
		log.Println(lg("API", "Github", color.RedString, "Error fetching current Release JSON: %s", err))
		return true
	}

	thisVersion, err := version.NewVersion(projectVersion)
	if err != nil {
		log.Println(lg("API", "Github", color.RedString, "Error parsing current version: %s", err))
		return true
	}

	latestVersion, err := version.NewVersion(githubReleaseApiObject.TagName)
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

//#region Requests

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

//#endregion

//#region Log

/*const (
	logLevelOff       = -1
	logLevelEssential = iota
	logLevelFatal
	logLevelError
	logLevelWarning
	logLevelInfo
	logLevelDebug
	logLevelVerbose
	logLevelAll
)*/

func lg(group string, subgroup string, colorFunc func(string, ...interface{}) string, line string, p ...interface{}) string {
	colorPrefix := group
	switch strings.ToLower(group) {

	case "main":
		if subgroup == "" {
			colorPrefix = color.CyanString("[~]")
		} else {
			colorPrefix = color.CyanString("[~:%s]", subgroup)
		}
	case "debug":
		if subgroup == "" {
			colorPrefix = color.HiYellowString("[DEBUG]")
		} else {
			colorPrefix = color.HiYellowString("[DEBUG | %s]", subgroup)
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
		colorPrefix = color.HiMagentaString("[Version]")

	case "settings":
		colorPrefix = color.GreenString("[Settings]")

	case "database":
		colorPrefix = color.HiYellowString("[Database]")

	case "setup":
		colorPrefix = color.HiGreenString("[Setup]")

	case "checkup":
		colorPrefix = color.HiGreenString("[Checkup]")

	case "discord":
		colorPrefix = color.HiBlueString("[Discord]")

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
	if strings.ToLower(group) == "debug" || strings.ToLower(subgroup) == "debug" {
		pp = color.YellowString("? ")
	}

	return "\t" + pp + colorPrefix + " " + colorFunc(line, p...)
}

//#endregion
