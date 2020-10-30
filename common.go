package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

var startTime time.Time

func uptime() time.Duration {
	return time.Since(startTime)
}

func delayedExit() {
	// Not formatting string because I only want the exit message to be red.
	log.Println(color.HiRedString("Exiting in 15 seconds"), "- uptime was ", durafmt.Parse(time.Since(startTime)).String(), "...")
	time.Sleep(15 * time.Second)
	os.Exit(1)
}

func filenameFromUrl(inputURL string) string {
	base := path.Base(inputURL)
	parts := strings.Split(base, "?")
	return parts[0]
}

func filepathExtension(filepath string) string {
	if strings.Contains(filepath, "?") {
		filepath = strings.Split(filepath, "?")[0]
	}
	filepath = path.Ext(filepath)
	return filepath
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func formatNumber(n int64) string {
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
			out[j] = ','
		}
	}
}

func formatNumberShort(x int64) string {
	if x > 1000 {
		x_number_format := formatNumber(x)
		x_array := strings.Split(x_number_format, ",")
		x_parts := [4]string{"k", "m", "b", "t"}
		x_count_parts := len(x_array) - 1
		var x_display string
		if x_array[1][:1] != "0" {
			x_display = fmt.Sprintf("%s.%s%s", x_array[0], x_array[1][:1], x_parts[x_count_parts-1])
		} else {
			x_display = fmt.Sprintf("%s%s", x_array[0], x_parts[x_count_parts-1])
		}
		return x_display
	}
	return fmt.Sprint(x)
}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getJsonWithHeaders(url string, target interface{}, headers map[string]string) error {
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
