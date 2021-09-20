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

	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/hashicorp/go-version"
)

var (
	pathBlacklist = []string{"/", "\\", "<", ">", ":", "\"", "|", "?", "*"}
)

func uptime() time.Duration {
	return time.Since(startTime)
}

func properExit() {
	// Not formatting string because I only want the exit message to be red.
	log.Println(color.HiRedString("[EXIT IN 15 SECONDS]"), " Uptime was", durafmt.Parse(time.Since(startTime)).String(), "...")
	log.Println(color.HiCyanString("--------------------------------------------------------------------------------"))
	time.Sleep(15 * time.Second)
	os.Exit(1)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToLower(b) == strings.ToLower(a) {
			return true
		}
	}
	return false
}

//#region Formatting

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

func boolS(val bool) string {
	if val {
		return "ON"
	}
	return "OFF"
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

//#endregion

//#region Requests

type githubReleaseApiObject struct {
	TagName string `json:"tag_name"`
}

func isLatestGithubRelease() bool {
	prefixHere := color.HiMagentaString("[Github Update Check]")

	githubReleaseApiObject := new(githubReleaseApiObject)
	err := getJSON(projectReleaseApiURL, githubReleaseApiObject)
	if err != nil {
		log.Println(prefixHere, color.RedString("Error fetching current Release JSON: %s", err))
		return true
	}

	thisVersion, err := version.NewVersion(projectVersion)
	if err != nil {
		log.Println(prefixHere, color.RedString("Error parsing current version: %s", err))
		return true
	}

	latestVersion, err := version.NewVersion(githubReleaseApiObject.TagName)
	if err != nil {
		log.Println(prefixHere, color.RedString("Error parsing latest version: %s", err))
		return true
	}

	if latestVersion.GreaterThan(thisVersion) {
		return false
	}

	return true
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

//#endregion

//#region Parsing

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

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

//#endregion
