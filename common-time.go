package main

import (
	"strings"
	"time"

	"github.com/hako/durafmt"
)

/*const (
	timeFmtAMPML       = "pm"
	timeFmtAMPMU       = "PM"
	timeFmtH12         = "3"
	timeFmtH24         = "15"
	timeFmtMM          = "04" // H:MM:SS
	timeFmtSS          = "05" // H:MM:SS
	timeFmtHMM12       = timeFmtH12 + ":" + timeFmtMM
	timeFmtHMMSS12     = timeFmtHMM12 + ":" + timeFmtSS
	timeFmtHMM12AMPM   = timeFmtHMM12 + timeFmtAMPML
	timeFmtHMMSS12AMPM = timeFmtHMMSS12 + timeFmtAMPML
	timeFmtHMM24       = timeFmtH24 + ":" + timeFmtMM
	timeFmtHMMSS24     = timeFmtHMM24 + ":" + timeFmtSS
	timeFmtTZ          = "MST"

	dateFmtDOM      = "2" // Day of month, 2nd
	dateFmtMonthNum = "1"
	dateFmtMonth    = "January"
	dateFmtYear     = "2006"
	dateFmtMDYSlash = dateFmtMonthNum + "/" + dateFmtDOM + "/" + dateFmtYear
	dateFmtMDYDash  = dateFmtMonthNum + "-" + dateFmtDOM + "-" + dateFmtYear
	dateFmtDMYSlash = dateFmtDOM + "/" + dateFmtMonthNum + "/" + dateFmtYear
	dateFmtDMYDash  = dateFmtDOM + "-" + dateFmtMonthNum + "-" + dateFmtYear
	dateFmtDateMD   = dateFmtMonth + " " + dateFmtDOM + ", " + dateFmtYear
	dateFmtDateDM   = dateFmtDOM + " " + dateFmtMonth + ", " + dateFmtYear
)*/

func isDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func shortenTime(input string) string {
	input = strings.ReplaceAll(input, " nanoseconds", "ns")
	input = strings.ReplaceAll(input, " nanosecond", "ns")
	input = strings.ReplaceAll(input, " microseconds", "μs")
	input = strings.ReplaceAll(input, " microsecond", "μs")
	input = strings.ReplaceAll(input, " milliseconds", "ms")
	input = strings.ReplaceAll(input, " millisecond", "ms")
	input = strings.ReplaceAll(input, " seconds", "s")
	input = strings.ReplaceAll(input, " second", "s")
	input = strings.ReplaceAll(input, " minutes", "m")
	input = strings.ReplaceAll(input, " minute", "m")
	input = strings.ReplaceAll(input, " hours", "h")
	input = strings.ReplaceAll(input, " hour", "h")
	input = strings.ReplaceAll(input, " days", "d")
	input = strings.ReplaceAll(input, " day", "d")
	input = strings.ReplaceAll(input, " weeks", "w")
	input = strings.ReplaceAll(input, " week", "w")
	input = strings.ReplaceAll(input, " months", "mo")
	input = strings.ReplaceAll(input, " month", "mo")
	return input
}

func timeSince(input time.Time) string {
	return durafmt.Parse(time.Since(input)).String()
}

func timeSinceShort(input time.Time) string {
	return shortenTime(durafmt.ParseShort(time.Since(input)).String())
}
