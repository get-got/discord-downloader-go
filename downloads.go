package main

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/rivo/duplo"
	"mvdan.cc/xurls/v2"
)

type download struct {
	URL         string
	Time        time.Time
	Destination string
	Filename    string
	ChannelID   string
	UserID      string
}

type downloadStatus int

const (
	downloadSuccess downloadStatus = iota

	downloadIgnored

	downloadSkipped
	downloadSkippedDuplicate
	downloadSkippedUnpermittedDomain
	downloadSkippedUnpermittedType
	downloadSkippedUnpermittedExtension
	downloadSkippedDetectedDuplicate

	downloadFailed
	downloadFailedInvalidSource
	downloadFailedInvalidPath
	downloadFailedCreatingFolder
	downloadFailedRequesting
	downloadFailedDownloadingResponse
	downloadFailedReadResponse
	downloadFailedCreatingSubfolder
	downloadFailedWritingFile
	downloadFailedWritingDatabase
)

type downloadStatusStruct struct {
	Status downloadStatus
	Error  error
}

func mDownloadStatus(status downloadStatus, _error ...error) downloadStatusStruct {
	if len(_error) == 0 {
		return downloadStatusStruct{
			Status: status,
			Error:  nil,
		}
	}
	return downloadStatusStruct{
		Status: status,
		Error:  _error[0],
	}
}

func getDownloadStatusString(status downloadStatus) string {
	switch status {
	case downloadSuccess:
		return "Download Succeeded"
	//
	case downloadIgnored:
		return "Download Ignored"
	//
	case downloadSkipped:
		return "Download Skipped"
	case downloadSkippedDuplicate:
		return "Download Skipped - Duplicate"
	case downloadSkippedUnpermittedDomain:
		return "Download Skipped - Unpermitted Domain"
	case downloadSkippedUnpermittedType:
		return "Download Skipped - Unpermitted File Type"
	case downloadSkippedUnpermittedExtension:
		return "Download Skipped - Unpermitted File Extension"
	case downloadSkippedDetectedDuplicate:
		return "Download Skipped - Detected Duplicate"
	//
	case downloadFailed:
		return "Download Failed"
	case downloadFailedInvalidSource:
		return "Download Failed - Invalid Source"
	case downloadFailedInvalidPath:
		return "Download Failed - Invalid Path"
	case downloadFailedCreatingFolder:
		return "Download Failed - Error Creating Folder"
	case downloadFailedRequesting:
		return "Download Failed - Error Requesting URL Data"
	case downloadFailedDownloadingResponse:
		return "Download Failed - Error Downloading URL Response"
	case downloadFailedReadResponse:
		return "Download Failed - Error Reading URL Response"
	case downloadFailedCreatingSubfolder:
		return "Download Failed - Error Creating Subfolder for Type"
	case downloadFailedWritingFile:
		return "Download Failed - Error Writing File"
	case downloadFailedWritingDatabase:
		return "Download Failed - Error Writing to Database"
	}
	return "Unknown Error"
}

// Trim duplicate links in link list
func trimDuplicateLinks(fileItems []*fileItem) []*fileItem {
	var result []*fileItem
	seen := map[string]bool{}

	for _, item := range fileItems {
		if seen[item.Link] {
			continue
		}

		seen[item.Link] = true
		result = append(result, item)
	}

	return result
}

func getRawLinks(m *discordgo.Message) []*fileItem {
	var links []*fileItem

	if m.Author == nil {
		m.Author = new(discordgo.User)
	}

	for _, attachment := range m.Attachments {
		links = append(links, &fileItem{
			Link:     attachment.URL,
			Filename: attachment.Filename,
		})
	}

	foundLinks := xurls.Strict().FindAllString(m.Content, -1)
	for _, foundLink := range foundLinks {
		links = append(links, &fileItem{
			Link: foundLink,
		})
	}

	for _, embed := range m.Embeds {
		if embed.URL != "" {
			links = append(links, &fileItem{
				Link: embed.URL,
			})
		}

		// Removing for now as this causes it to try and pull shit from things like YouTube descriptions
		/*if embed.Description != "" {
			foundLinks = xurls.Strict().FindAllString(embed.Description, -1)
			for _, foundLink := range foundLinks {
				links = append(links, &fileItem{
					Link: foundLink,
				})
			}
		}*/

		if embed.Image != nil && embed.Image.URL != "" {
			links = append(links, &fileItem{
				Link: embed.Image.URL,
			})
		}

		if embed.Video != nil && embed.Video.URL != "" {
			links = append(links, &fileItem{
				Link: embed.Video.URL,
			})
		}
	}

	return links
}

func getDownloadLinks(inputURL string, channelID string) map[string]string {
	logPrefixErrorHere := color.HiRedString("[getDownloadLinks]")
	/*
		Throw out photo/[0-9]+ and mobile.
		Throw out trailing slash.

		These are specifically for Twitter. Why are they not in their respective if cases?
		Sometimes MatchString(inputURL) is called multiple (3 or so) times with the same url, somehow gets around else { return nil }
		and goes down the execution chain, causing unsupported type text.
	*/

	inputURL = strings.ReplaceAll(inputURL, "mobile.twitter", "twitter")
	photo := regexp.MustCompile(`(\/)?photo(\/)?([0-9]+)?(\/)?$`)
	trailingslash := regexp.MustCompile(`(\/)$`)

	inputURL = photo.ReplaceAllString(inputURL, "")
	inputURL = trailingslash.ReplaceAllString(inputURL, "")

	/* TODO: Download Support...
	- TikTok: Tried, once the connection is closed the cdn URL is rendered invalid
	- Facebook Photos: Tried, it doesn't preload image data, it's loaded in after. Would have to keep connection open, find alternative way to grab, or use api.
	- Facebook Videos: Previously supported but they split mp4 into separate audio and video streams
	*/

	if regexUrlTwitter.MatchString(inputURL) {
		links, err := getTwitterUrls(inputURL)
		if err != nil {
			if !strings.Contains(err.Error(), "suspended") {
				log.Println(logPrefixErrorHere, color.RedString("Twitter Media fetch failed for %s -- %s", inputURL, err))
			}
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		} else {
			//stop looking for link matches here. fixes unsupported type on twitter
			return nil
		}
	}
	if regexUrlTwitterStatus.MatchString(inputURL) {
		links, err := getTwitterStatusUrls(inputURL, channelID)
		if err != nil {
			if !strings.Contains(err.Error(), "suspended") && !strings.Contains(err.Error(), "No status found") {
				log.Println(logPrefixErrorHere, color.RedString("Twitter Status fetch failed for %s -- %s", inputURL, err))
			}
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		} else {
			//stop looking for link matches here. fixes unsupported type on twitter
			return nil
		}
	}

	if regexUrlInstagram.MatchString(inputURL) {
		links, err := getInstagramUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Instagram fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if regexUrlImgurSingle.MatchString(inputURL) {
		links, err := getImgurSingleUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Imgur Media fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}
	if regexUrlImgurAlbum.MatchString(inputURL) {
		links, err := getImgurAlbumUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Imgur Album fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if regexUrlStreamable.MatchString(inputURL) {
		links, err := getStreamableUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Streamable fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if regexUrlGfycat.MatchString(inputURL) {
		links, err := getGfycatUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Gfycat fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if regexUrlFlickrPhoto.MatchString(inputURL) {
		links, err := getFlickrPhotoUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Flickr Photo fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}
	if regexUrlFlickrAlbum.MatchString(inputURL) {
		links, err := getFlickrAlbumUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Flickr Album fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}
	if regexUrlFlickrAlbumShort.MatchString(inputURL) {
		links, err := getFlickrAlbumShortUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Flickr Album (short) fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if config.Credentials.GoogleDriveCredentialsJSON != "" {
		if regexUrlGoogleDrive.MatchString(inputURL) {
			links, err := getGoogleDriveUrls(inputURL)
			if err != nil {
				log.Println(logPrefixErrorHere, color.RedString("Google Drive Album URL for %s -- %s", inputURL, err))
			} else if len(links) > 0 {
				return trimDownloadedLinks(links, channelID)
			}
		}
		if regexUrlGoogleDriveFolder.MatchString(inputURL) {
			links, err := getGoogleDriveFolderUrls(inputURL)
			if err != nil {
				log.Println(logPrefixErrorHere, color.RedString("Google Drive Folder URL for %s -- %s", inputURL, err))
			} else if len(links) > 0 {
				return trimDownloadedLinks(links, channelID)
			}
		}
	}

	if regexUrlTistory.MatchString(inputURL) {
		links, err := getTistoryUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Tistory URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}
	if regexUrlTistoryLegacy.MatchString(inputURL) {
		links, err := getLegacyTistoryUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Legacy Tistory URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if regexUrlRedditPost.MatchString(inputURL) {
		links, err := getRedditPostUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Reddit Post URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if regexUrlMastodonPost1.MatchString(inputURL) || regexUrlMastodonPost2.MatchString(inputURL) {
		links, err := getMastodonPostUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Mastodon Post URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	// The original project has this as an option,
	if regexUrlPossibleTistorySite.MatchString(inputURL) {
		links, err := getPossibleTistorySiteUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Checking for Tistory site failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if strings.HasPrefix(inputURL, "https://cdn.discordapp.com/emojis/") {
		return nil
	}

	// Try without queries
	parsedURL, err := url.Parse(inputURL)
	if err == nil {
		parsedURL.RawQuery = ""
		inputURLWithoutQueries := parsedURL.String()
		if inputURLWithoutQueries != inputURL {
			return trimDownloadedLinks(getDownloadLinks(inputURLWithoutQueries, channelID), channelID)
		}
	}

	return trimDownloadedLinks(map[string]string{inputURL: ""}, channelID)
}

func getFileLinks(m *discordgo.Message) []*fileItem {
	var fileItems []*fileItem

	linkTime, err := m.Timestamp.Parse()
	if err != nil {
		linkTime = time.Now()
	}

	rawLinks := getRawLinks(m)
	for _, rawLink := range rawLinks {
		downloadLinks := getDownloadLinks(
			rawLink.Link,
			m.ChannelID,
		)
		for link, filename := range downloadLinks {
			if rawLink.Filename != "" {
				filename = rawLink.Filename
			}

			fileItems = append(fileItems, &fileItem{
				Link:     link,
				Filename: filename,
				Time:     linkTime,
			})
		}
	}

	fileItems = trimDuplicateLinks(fileItems)

	return fileItems
}

func startDownload(inputURL string, filename string, path string, message *discordgo.Message, fileTime time.Time, historyCmd bool, emojiCmd bool) downloadStatusStruct {
	status := mDownloadStatus(downloadFailed)
	logPrefixErrorHere := color.HiRedString("[startDownload]")

	for i := 0; i < config.DownloadRetryMax; i++ {
		status = tryDownload(inputURL, filename, path, message, fileTime, historyCmd, emojiCmd)
		if status.Status < downloadFailed { // Success or Skip
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	if status.Status >= downloadFailed && !historyCmd { // Any kind of failure
		log.Println(logPrefixErrorHere, color.RedString("Gave up on downloading %s after %d failed attempts...\t%s", inputURL, config.DownloadRetryMax, getDownloadStatusString(status.Status)))
		if isChannelRegistered(message.ChannelID) {
			channelConfig := getChannelConfig(message.ChannelID)
			if !historyCmd && *channelConfig.ErrorMessages {
				content := fmt.Sprintf(
					"Gave up trying to download\n<%s>\nafter %d failed attempts...\n\n``%s``",
					inputURL, config.DownloadRetryMax, getDownloadStatusString(status.Status))
				if status.Error != nil {
					content += fmt.Sprintf("\n```ERROR: %s```", status.Error)
				}
				// Failure Notice
				if hasPerms(message.ChannelID, discordgo.PermissionSendMessages) {
					_, err := bot.ChannelMessageSendComplex(message.ChannelID,
						&discordgo.MessageSend{
							Content: fmt.Sprintf("<@!%s>", message.Author.ID),
							Embed:   buildEmbed(message.ChannelID, "Download Failure", content),
						})
					if err != nil {
						log.Println(logPrefixErrorHere, color.HiRedString("Failed to send failure message to %s: %s", message.ChannelID, err))
					}
				} else {
					log.Println(logPrefixErrorHere, color.HiRedString(fmtBotSendPerm, message.ChannelID))
				}
			}
			if status.Error != nil {
				logErrorMessage(fmt.Sprintf("**%s**\n\n%s", getDownloadStatusString(status.Status), status.Error))
			}
		}
	}

	return status
}

func tryDownload(inputURL string, filename string, path string, message *discordgo.Message, fileTime time.Time, historyCmd bool, emojiCmd bool) downloadStatusStruct {
	cachedDownloadID++
	thisDownloadID := cachedDownloadID

	startTime := time.Now()

	logPrefixErrorHere := color.HiRedString("[tryDownload]")
	logPrefix := ""
	if historyCmd {
		logPrefix = logPrefixHistory + " "
	}

	if stringInSlice(message.ChannelID, getAllChannels()) || emojiCmd {
		var channelConfig configurationChannel
		if isChannelRegistered(message.ChannelID) {
			channelConfig = getChannelConfig(message.ChannelID)
		} else {
			channelDefault(&channelConfig)
		}

		var err error

		// Source validation
		_, err = url.ParseRequestURI(inputURL)
		if err != nil {
			return mDownloadStatus(downloadFailedInvalidSource, err)
		}

		// Clean/fix path
		if path == "" || path == string(os.PathSeparator) {
			log.Println(logPrefixErrorHere, color.HiRedString("Destination cannot be empty path..."))
			return mDownloadStatus(downloadFailedInvalidPath, err)
		}
		if !strings.HasSuffix(path, string(os.PathSeparator)) {
			path = path + string(os.PathSeparator)
		}

		// Create folder
		err = os.MkdirAll(path, 0755)
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Error while creating destination folder \"%s\": %s", path, err))
			return mDownloadStatus(downloadFailedCreatingFolder, err)
		}

		// Request
		timeout := time.Duration(time.Duration(config.DownloadTimeout) * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		request, err := http.NewRequest("GET", inputURL, nil)
		request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.139 Safari/537.36")
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Error while requesting \"%s\": %s", inputURL, err))
			return mDownloadStatus(downloadFailedRequesting, err)
		}
		request.Header.Add("Accept-Encoding", "identity")
		response, err := client.Do(request)
		if err != nil {
			if !strings.Contains(err.Error(), "no such host") && !strings.Contains(err.Error(), "connection refused") {
				log.Println(logPrefixErrorHere, color.HiRedString("Error while receiving response from \"%s\": %s", inputURL, err))
			}
			return mDownloadStatus(downloadFailedDownloadingResponse, err)
		}
		defer response.Body.Close()

		// Download duration
		if config.DebugOutput && !historyCmd {
			log.Println(logPrefixDebug, color.YellowString("#%d - %s to download.", thisDownloadID, durafmt.ParseShort(time.Since(startTime)).String()))
		}
		downloadTime := time.Now()

		// Read
		bodyOfResp, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Could not read response from \"%s\": %s", inputURL, err))
			return mDownloadStatus(downloadFailedReadResponse, err)
		}

		// Filename
		if filename == "" {
			filename = filenameFromURL(response.Request.URL.String())
			for key, iHeader := range response.Header {
				if key == "Content-Disposition" {
					_, params, err := mime.ParseMediaType(iHeader[0])
					if err == nil {
						newFilename, err := url.QueryUnescape(params["filename"])
						if err != nil {
							newFilename = params["filename"]
						}
						if newFilename != "" {
							filename = newFilename
						}
					}
				}
			}
		}

		extension := strings.ToLower(filepath.Ext(filename))

		contentType := http.DetectContentType(bodyOfResp)
		contentTypeParts := strings.Split(contentType, "/")
		contentTypeFound := contentTypeParts[0]

		parsedURL, err := url.Parse(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Error while parsing url:\t%s", err))
		}

		// Check extension
		if channelConfig.Filters.AllowedExtensions != nil || channelConfig.Filters.BlockedExtensions != nil {
			shouldAbort := false
			if channelConfig.Filters.AllowedExtensions != nil {
				shouldAbort = true
			}

			if channelConfig.Filters.BlockedExtensions != nil {
				if stringInSlice(extension, *channelConfig.Filters.BlockedExtensions) {
					shouldAbort = true
				}
			}
			if channelConfig.Filters.AllowedExtensions != nil {
				if stringInSlice(extension, *channelConfig.Filters.AllowedExtensions) {
					shouldAbort = false
				}
			}

			// Abort
			if shouldAbort {
				if !historyCmd {
					log.Println(logPrefixFileSkip, color.GreenString("Unpermitted extension (%s) found at %s", extension, inputURL))
				}

				file, ferr := os.OpenFile(message.ChannelID+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Println(ferr)
				}
				_, werr := file.WriteString(inputURL + " (Unpermitted extension)\n")
				if werr != nil {
					fmt.Println(werr)
				}
				file.Close()

				return mDownloadStatus(downloadSkippedUnpermittedExtension)
			}
		}

		// Fix content type
		if stringInSlice(extension, []string{".mov"}) ||
			stringInSlice(extension, []string{".mp4"}) ||
			stringInSlice(extension, []string{".webm"}) {
			contentTypeFound = "video"
		} else if stringInSlice(extension, []string{".psd"}) ||
			stringInSlice(extension, []string{".nef"}) ||
			stringInSlice(extension, []string{".dng"}) ||
			stringInSlice(extension, []string{".tif"}) ||
			stringInSlice(extension, []string{".tiff"}) {
			contentTypeFound = "image"
		}

		// Filename extension fix
		if filepath.Ext(filename) == "" {
			possibleExtension, _ := mime.ExtensionsByType(contentType)
			if len(possibleExtension) > 0 {
				filename += possibleExtension[0]
			}
		}
		// Filename validation
		if !regexFilename.MatchString(filename) {
			filename = "InvalidFilename"
			possibleExtension, _ := mime.ExtensionsByType(contentType)
			if len(possibleExtension) > 0 {
				filename += possibleExtension[0]
			}
		}

		// Check Domain
		if channelConfig.Filters.AllowedDomains != nil || channelConfig.Filters.BlockedDomains != nil {
			shouldAbort := false
			if channelConfig.Filters.AllowedDomains != nil {
				shouldAbort = true
			}

			if channelConfig.Filters.BlockedDomains != nil {
				if stringInSlice(parsedURL.Hostname(), *channelConfig.Filters.BlockedDomains) {
					shouldAbort = true
				}
			}
			if channelConfig.Filters.AllowedDomains != nil {
				if stringInSlice(parsedURL.Hostname(), *channelConfig.Filters.AllowedDomains) {
					shouldAbort = false
				}
			}

			// Abort
			if shouldAbort {
				if !historyCmd {
					log.Println(logPrefixFileSkip, color.GreenString("Unpermitted domain (%s) found at %s", parsedURL.Hostname(), inputURL))
				}
				return mDownloadStatus(downloadSkippedUnpermittedDomain)
			}
		}

		// Check content type
		if !((*channelConfig.SaveImages && contentTypeFound == "image") ||
			(*channelConfig.SaveVideos && contentTypeFound == "video") ||
			(*channelConfig.SaveAudioFiles && contentTypeFound == "audio") ||
			(*channelConfig.SaveTextFiles && contentTypeFound == "text") ||
			(*channelConfig.SaveOtherFiles && contentTypeFound == "application")) {
			if !historyCmd {
				log.Println(logPrefixFileSkip, color.GreenString("Unpermitted filetype (%s) found at %s", contentTypeFound, inputURL))
			}

			file, ferr := os.OpenFile(message.ChannelID+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println(ferr)
			}
			_, werr := file.WriteString(inputURL + " (Unpermitted filetype)\n")
			if werr != nil {
				fmt.Println(werr)
			}
			file.Close()

			return mDownloadStatus(downloadSkippedUnpermittedType)
		}

		// Duplicate Image Filter
		if config.FilterDuplicateImages && contentTypeFound == "image" && extension != ".gif" && extension != ".webp" {
			img, _, err := image.Decode(bytes.NewReader(bodyOfResp))
			if err != nil {
				log.Println(color.HiRedString("Error converting buffer to image for hashing:\t%s", err))
			} else {
				hash, _ := duplo.CreateHash(img)
				matches := imgStore.Query(hash)
				sort.Sort(matches)
				for _, match := range matches {
					/*if config.DebugOutput {
						log.Println(color.YellowString("Similarity Score: %f", match.Score))
					}*/
					if match.Score < config.FilterDuplicateImagesThreshold {
						log.Println(logPrefixFileSkip, color.GreenString("Duplicate detected (Score of %f) found at %s", match.Score, inputURL))
						return mDownloadStatus(downloadSkippedDetectedDuplicate)
					}
				}
				imgStore.Add(cachedDownloadID, hash)
			}
		}

		// Names
		sourceChannelName := message.ChannelID
		sourceName := "UNKNOWN"
		sourceChannel, err := bot.State.Channel(message.ChannelID)
		if sourceChannel != nil {
			if sourceChannel.Name != "" {
				sourceChannelName = sourceChannel.Name
			}
			switch sourceChannel.Type {
			case discordgo.ChannelTypeGuildText:
				if sourceChannel.GuildID != "" {
					sourceGuild, _ := bot.State.Guild(sourceChannel.GuildID)
					if sourceGuild != nil && sourceGuild.Name != "" {
						sourceName = "\"" + sourceGuild.Name + "\""
					}
				}
			case discordgo.ChannelTypeDM:
				sourceName = "Direct Messages"
			case discordgo.ChannelTypeGroupDM:
				sourceName = "Group Messages"
			}
		}

		subfolder := ""
		if message.Author != nil {
			// Subfolder Division - Server Nesting
			if *channelConfig.DivideFoldersByServer {
				subfolderSuffix := ""
				if sourceName != "" && sourceName != "UNKNOWN" {
					subfolderSuffix = sourceName
					for _, key := range pathBlacklist {
						subfolderSuffix = strings.ReplaceAll(subfolderSuffix, key, "")
					}
				}
				if subfolderSuffix != "" {
					subfolderSuffix = subfolderSuffix + string(os.PathSeparator)
					subfolder = subfolder + subfolderSuffix
					// Create folder.
					err := os.MkdirAll(path+subfolder, 0755)
					if err != nil {
						log.Println(logPrefixErrorHere, color.HiRedString("Error while creating server subfolder \"%s\": %s", path, err))
						return mDownloadStatus(downloadFailedCreatingSubfolder, err)
					}
				}
			}
			// Subfolder Division - Channel Nesting
			if *channelConfig.DivideFoldersByChannel {
				subfolderSuffix := ""
				if sourceChannelName != "" {
					subfolderSuffix = sourceChannelName
					for _, key := range pathBlacklist {
						subfolderSuffix = strings.ReplaceAll(subfolderSuffix, key, "")
					}
				}
				if subfolderSuffix != "" {
					subfolder = subfolder + subfolderSuffix + string(os.PathSeparator)
					// Create folder.
					err := os.MkdirAll(path+subfolder, 0755)
					if err != nil {
						log.Println(logPrefixErrorHere, color.HiRedString("Error while creating channel subfolder \"%s\": %s", path, err))
						return mDownloadStatus(downloadFailedCreatingSubfolder, err)
					}
				}
			}

			// Subfolder Division - User Nesting
			if *channelConfig.DivideFoldersByUser {
				subfolderSuffix := message.Author.ID
				if message.Author.Username != "" {
					subfolderSuffix = message.Author.Username + "#" + message.Author.Discriminator
					for _, key := range pathBlacklist {
						subfolderSuffix = strings.ReplaceAll(subfolderSuffix, key, "")
					}
				}
				if subfolderSuffix != "" {
					subfolder = subfolder + subfolderSuffix + string(os.PathSeparator)
					// Create folder.
					err := os.MkdirAll(path+subfolder, 0755)
					if err != nil {
						log.Println(logPrefixErrorHere, color.HiRedString("Error while creating user subfolder \"%s\": %s", path, err))
						return mDownloadStatus(downloadFailedCreatingSubfolder, err)
					}
				}
			}
		}

		// Subfolder Division - Content Type
		if *channelConfig.DivideFoldersByType && message.Author != nil {
			subfolderSuffix := ""
			switch contentTypeFound {
			case "image":
				subfolderSuffix = "images"
			case "video":
				subfolderSuffix = "videos"
			case "audio":
				subfolderSuffix = "audio"
			case "text":
				subfolderSuffix = "text"
			case "application":
				subfolderSuffix = "applications"
			}
			if subfolderSuffix != "" {
				subfolder = subfolder + subfolderSuffix + string(os.PathSeparator)
				// Create folder.
				err := os.MkdirAll(path+subfolder, 0755)
				if err != nil {
					log.Println(logPrefixErrorHere, color.HiRedString("Error while creating type subfolder \"%s\": %s", path+subfolder, err))
					return mDownloadStatus(downloadFailedCreatingSubfolder, err)
				}
			}
		}

		// Format filename/path

		completePath := path + subfolder + filename

		// Check if exists
		if _, err := os.Stat(completePath); err == nil {
			if *channelConfig.SavePossibleDuplicates {
				tmpPath := completePath
				i := 1
				for {
					// Append number to name
					completePath = tmpPath[0:len(tmpPath)-len(filepathExtension(tmpPath))] +
						"-" + strconv.Itoa(i) + filepathExtension(tmpPath)
					if _, err := os.Stat(completePath); os.IsNotExist(err) {
						break
					}
					i = i + 1
				}
				if !historyCmd {
					log.Println(color.GreenString("Matching filenames, possible duplicate? Saving \"%s\" as \"%s\" instead", tmpPath, completePath))
				}
			} else {
				if !historyCmd {
					log.Println(logPrefixFileSkip, color.GreenString("Matching filenames, possible duplicate..."))
				}
				return mDownloadStatus(downloadSkippedDuplicate)
			}
		}

		// Write
		err = ioutil.WriteFile(completePath, bodyOfResp, 0644)
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Error while writing file to disk \"%s\": %s", inputURL, err))
			return mDownloadStatus(downloadFailedWritingFile, err)
		}

		// Change file time
		err = os.Chtimes(completePath, fileTime, fileTime)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Error while changing metadata date \"%s\": %s", inputURL, err))
		}

		// Write duration
		if config.DebugOutput && !historyCmd {
			log.Println(logPrefixDebug, color.YellowString("#%d - %s to save.", thisDownloadID, durafmt.ParseShort(time.Since(downloadTime)).String()))
		}
		writeTime := time.Now()

		// Output
		log.Println(logPrefix + color.HiGreenString("SAVED %s sent in %s#%s to \"%s\"", strings.ToUpper(contentTypeFound), sourceName, sourceChannelName, completePath))

		userID := user.ID
		if message.Author != nil {
			userID = message.Author.ID
		}
		// Store in db
		err = dbInsertDownload(&download{
			URL:         inputURL,
			Time:        time.Now(),
			Destination: completePath,
			Filename:    filename,
			ChannelID:   message.ChannelID,
			UserID:      userID,
		})
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Error writing to database: %s", err))
			return mDownloadStatus(downloadFailedWritingDatabase, err)
		}

		// Storage & output duration
		if config.DebugOutput && !historyCmd {
			log.Println(logPrefixDebug, color.YellowString("#%d - %s to update database.", thisDownloadID, durafmt.ParseShort(time.Since(writeTime)).String()))
		}
		finishTime := time.Now()

		// React
		if !historyCmd && *channelConfig.ReactWhenDownloaded && message.Author != nil {
			reaction := ""
			if *channelConfig.ReactWhenDownloadedEmoji == "" {
				if message.GuildID != "" {
					guild, err := bot.State.Guild(message.GuildID)
					if err != nil {
						log.Println(logPrefixErrorHere, color.RedString("Error fetching guild state for emojis from %s: %s", message.GuildID, err))
					} else {
						emojis := guild.Emojis
						if len(emojis) > 1 {
							for {
								rand.Seed(time.Now().UnixNano())
								chosenEmoji := emojis[rand.Intn(len(emojis))]
								formattedEmoji := chosenEmoji.APIName()
								if !chosenEmoji.Animated && !stringInSlice(formattedEmoji, *channelConfig.BlacklistReactEmojis) {
									reaction = formattedEmoji
									break
								}
							}
						} else {
							reaction = defaultReact
						}
					}
				} else {
					reaction = defaultReact
				}
			} else {
				reaction = *channelConfig.ReactWhenDownloadedEmoji
			}
			// Add Reaction
			if hasPerms(message.ChannelID, discordgo.PermissionAddReactions) {
				err = bot.MessageReactionAdd(message.ChannelID, message.ID, reaction)
				if err != nil {
					log.Println(logPrefixErrorHere, color.RedString("Error adding reaction to message: %s", err))
				}
			} else {
				log.Println(logPrefixErrorHere, color.RedString("Bot does not have permission to add reactions in %s", message.ChannelID))
			}
			// React duration
			if config.DebugOutput {
				log.Println(logPrefixDebug, color.YellowString("#%d - %s to react with \"%s\".", thisDownloadID, durafmt.ParseShort(time.Since(finishTime)).String(), reaction))
			}
		}

		if !historyCmd {
			timeLastUpdated = time.Now()
			if *channelConfig.UpdatePresence {
				updateDiscordPresence()
			}
		}

		if config.DebugOutput && !historyCmd {
			log.Println(logPrefixDebug, color.YellowString("#%d - %s total.", thisDownloadID, time.Since(startTime)))
		}

		// Log Links to File
		if channelConfig.LogLinks != nil {
			if channelConfig.LogLinks.Destination != "" {
				logPath := channelConfig.LogLinks.Destination
				if *channelConfig.LogLinks.DestinationIsFolder == true {
					if !strings.HasSuffix(logPath, string(os.PathSeparator)) {
						logPath += string(os.PathSeparator)
					}
					err := os.MkdirAll(logPath, 0755)
					if err == nil {
						logPath += "Log_Links"
						if *channelConfig.LogLinks.DivideLogsByServer == true {
							if message.GuildID == "" {
								ch, err := bot.State.Channel(message.ChannelID)
								if err == nil {
									if ch.Type == discordgo.ChannelTypeDM {
										logPath += " DM"
									} else if ch.Type == discordgo.ChannelTypeGroupDM {
										logPath += " GroupDM"
									} else {
										logPath += " Unknown"
									}
								} else {
									logPath += " Unknown"
								}
							} else {
								logPath += " SID_" + message.GuildID
							}
						}
						if *channelConfig.LogLinks.DivideLogsByChannel == true {
							logPath += " CID_" + message.ChannelID
						}
						if *channelConfig.LogLinks.DivideLogsByUser == true {
							logPath += " UID_" + message.Author.ID
						}
					}
					logPath += ".txt"
				}
				// Read
				currentLog, err := ioutil.ReadFile(logPath)
				currentLogS := ""
				if err == nil {
					currentLogS = string(currentLog)
				}
				// Writer
				f, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
				if err != nil {
					log.Println(color.RedString("[channelConfig.LogLinks] Failed to open log file:\t%s", err))
					f.Close()
				}
				defer f.Close()

				var newLine string
				rawLinks := getRawLinks(message)
				// Each Link
				for _, rawLink := range rawLinks {
					// Filter Duplicates
					if channelConfig.LogLinks.FilterDuplicates != nil {
						if *channelConfig.LogLinks.FilterDuplicates {
							if strings.Contains(currentLogS, rawLink.Link) {
								continue
							}
						}
					}
					// Prepend
					prefix := ""
					if channelConfig.LogLinks.Prefix != nil {
						prefix = *channelConfig.LogLinks.Prefix
					}
					// More Data
					additionalInfo := ""
					if channelConfig.LogLinks.UserData != nil {
						if *channelConfig.LogLinks.UserData == true {
							additionalInfo = fmt.Sprintf("[%s/%s] \"%s\"#%s (%s) @ %s: ", message.GuildID, message.ChannelID, message.Author.Username, message.Author.Discriminator, message.Author.ID, message.Timestamp)
						}
					}
					// Append
					suffix := ""
					if channelConfig.LogLinks.Suffix != nil {
						suffix = *channelConfig.LogLinks.Suffix
					}
					// New Line
					newLine += "\n" + prefix + additionalInfo + rawLink.Link + suffix
				}

				if _, err = f.WriteString(newLine); err != nil {
					log.Println(color.RedString("[channelConfig.LogLinks] Failed to append file:\t%s", err))
				}
			}
		}

		if thisDownloadID > 0 {
			// Filter Duplicate Images
			if config.FilterDuplicateImages {
				encodedStore, err := imgStore.GobEncode()
				if err != nil {
					log.Println(color.HiRedString("Failed to encode imgStore:\t%s"))
				} else {
					f, err := os.OpenFile(imgStorePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
					if err != nil {
						log.Println(color.HiRedString("Failed to open imgStore file:\t%s"))
					}
					_, err = f.Write(encodedStore)
					if err != nil {
						log.Println(color.HiRedString("Failed to update imgStore file:\t%s"))
					}
					err = f.Close()
					if err != nil {
						log.Println(color.HiRedString("Failed to close imgStore file:\t%s"))
					}
				}
			}
		}

		return mDownloadStatus(downloadSuccess)
	}

	return mDownloadStatus(downloadIgnored)
}
