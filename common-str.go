package main

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/unicode/norm"
	godiacritics "gopkg.in/Regis24GmbH/go-diacritics.v2"
)

var (
	pathBlacklist = []string{"/", "\\", "<", ">", ":", "\"", "|", "?", "*"}
)

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
	if config.EuropeanNumbers {
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
	if config.EuropeanNumbers {
		numberSeparator = "."
	}
	var decimalSeparator string = "."
	if config.EuropeanNumbers {
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
	return wrapHyphens(i, 90)
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

func clearPathIllegalChars(p string) string {
	r := p
	for _, key := range pathBlacklist {
		r = strings.ReplaceAll(r, key, "")
	}
	return r
}

func clearDiacritics(p string) string {
	ret := p
	ret = godiacritics.Normalize(ret)
	ret = norm.NFKC.String(ret)
	return ret
}

func clearNonAscii(p string) string {
	re := regexp.MustCompile("[[:^ascii:]]")
	return re.ReplaceAllLiteralString(p, "")
}

func clearDoubleSpaces(p string) string {
	ret := p
	for {
		ret = strings.ReplaceAll(ret, "  ", " ")
		if !strings.Contains(ret, "  ") {
			break
		}
	}
	if ret == "" {
		return p
	}
	return ret
}

func clearSourceLogField(p string, cfg configurationSourceLog) string {
	ret := p

	if cfg.FilepathNormalizeText != nil {
		if *cfg.FilepathNormalizeText {
			ret = clearDiacritics(ret)
		}
	}

	if cfg.FilepathStripSymbols != nil {
		if *cfg.FilepathStripSymbols {
			ret = clearNonAscii(ret)
		}
	}

	ret = clearDoubleSpaces(ret)

	return ret
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

func getDomain(URL string) (string, error) {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return parsedURL.Hostname(), nil
	}
	return "UNKNOWN", err
}
