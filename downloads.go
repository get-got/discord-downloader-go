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

type downloadItem struct {
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
	downloadFailedCode
	downloadFailedCode404
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
	case downloadFailedCode:
		return "Download Failed - BAD CONNECTION"
	case downloadFailedCode404:
		return "Download Failed - 404 NOT FOUND"
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

func getDownloadLinks(inputURL string, m *discordgo.Message) map[string]string {
	/* TODO: Download Support...
	- TikTok: Tried, once the connection is closed the cdn URL is rendered invalid
	- Facebook Photos: Tried, it doesn't preload image data, it's loaded in after. Would have to keep connection open, find alternative way to grab, or use api.
	- Facebook Videos: Previously supported but they split mp4 into separate audio and video streams
	*/

	inputURL = strings.ReplaceAll(inputURL, "mobile.twitter", "twitter")
	inputURL = strings.ReplaceAll(inputURL, "fxtwitter.com", "twitter.com")
	inputURL = strings.ReplaceAll(inputURL, "c.vxtwitter.com", "twitter.com")
	inputURL = strings.ReplaceAll(inputURL, "vxtwitter.com", "twitter.com")

	if regexUrlTwitter.MatchString(inputURL) {
		links, err := getTwitterUrls(inputURL)
		if err != nil {
			if !strings.Contains(err.Error(), "suspended") {
				log.Println(lg("Download", "", color.RedString, "Twitter Media fetch failed for %s -- %s", inputURL, err))
			}
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}
	if regexUrlTwitterStatus.MatchString(inputURL) {
		links, err := getTwitterStatusUrls(inputURL, m)
		if err != nil {
			if !strings.Contains(err.Error(), "suspended") && !strings.Contains(err.Error(), "No status found") {
				log.Println(lg("Download", "", color.RedString, "Twitter Status fetch failed for %s -- %s", inputURL, err))
			}
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	if regexUrlInstagram.MatchString(inputURL) {
		links, err := getInstagramUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Instagram fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	if regexUrlImgurSingle.MatchString(inputURL) {
		links, err := getImgurSingleUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Imgur Media fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}
	if regexUrlImgurAlbum.MatchString(inputURL) {
		links, err := getImgurAlbumUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Imgur Album fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	if regexUrlStreamable.MatchString(inputURL) {
		links, err := getStreamableUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Streamable fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	if regexUrlGfycat.MatchString(inputURL) {
		links, err := getGfycatUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Gfycat fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	if regexUrlFlickrPhoto.MatchString(inputURL) {
		links, err := getFlickrPhotoUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Flickr Photo fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}
	if regexUrlFlickrAlbum.MatchString(inputURL) {
		links, err := getFlickrAlbumUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Flickr Album fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}
	if regexUrlFlickrAlbumShort.MatchString(inputURL) {
		links, err := getFlickrAlbumShortUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Flickr Album (short) fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	if config.Credentials.GoogleDriveCredentialsJSON != "" {
		if regexUrlGoogleDrive.MatchString(inputURL) {
			links, err := getGoogleDriveUrls(inputURL)
			if err != nil {
				log.Println(lg("Download", "", color.RedString, "Google Drive Album URL for %s -- %s", inputURL, err))
			} else if len(links) > 0 {
				return trimDownloadedLinks(links, m)
			}
		}
		if regexUrlGoogleDriveFolder.MatchString(inputURL) {
			links, err := getGoogleDriveFolderUrls(inputURL)
			if err != nil {
				log.Println(lg("Download", "", color.RedString, "Google Drive Folder URL for %s -- %s", inputURL, err))
			} else if len(links) > 0 {
				return trimDownloadedLinks(links, m)
			}
		}
	}

	if regexUrlTistory.MatchString(inputURL) {
		links, err := getTistoryUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Tistory URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}
	if regexUrlTistoryLegacy.MatchString(inputURL) {
		links, err := getLegacyTistoryUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Legacy Tistory URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	if regexUrlRedditPost.MatchString(inputURL) {
		links, err := getRedditPostUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Reddit Post URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	if regexUrlMastodonPost1.MatchString(inputURL) || regexUrlMastodonPost2.MatchString(inputURL) {
		links, err := getMastodonPostUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Mastodon Post URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	// The original project has this as an option,
	if regexUrlPossibleTistorySite.MatchString(inputURL) {
		links, err := getPossibleTistorySiteUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Checking for Tistory site failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, m)
		}
	}

	if strings.HasPrefix(inputURL, "https://cdn.discordapp.com/emojis/") {
		return nil
	}

	// Try without queries
	parsedURL, err := url.Parse(inputURL)
	if err == nil {
		if strings.Contains(parsedURL.String(), "format=") {
			parsedURL.RawQuery = "format=" + parsedURL.Query().Get("format")
		} else {
			parsedURL.RawQuery = ""
		}
		inputURLWithoutQueries := parsedURL.String()
		if inputURLWithoutQueries != inputURL {
			return trimDownloadedLinks(getDownloadLinks(inputURLWithoutQueries, m), m)
		}
	}

	return trimDownloadedLinks(map[string]string{inputURL: ""}, m)
}

func getFileLinks(m *discordgo.Message) []*fileItem {
	var fileItems []*fileItem

	linkTime := m.Timestamp

	rawLinks := getRawLinks(m)
	for _, rawLink := range rawLinks {
		downloadLinks := getDownloadLinks(rawLink.Link, m)
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

type downloadRequestStruct struct {
	InputURL       string
	Filename       string
	FileExtension  string
	Path           string
	Message        *discordgo.Message
	FileTime       time.Time
	HistoryCmd     bool
	EmojiCmd       bool
	ManualDownload bool
	StartTime      time.Time
}

func handleDownload(download downloadRequestStruct) downloadStatusStruct {
	status := mDownloadStatus(downloadFailed)
	for i := 0; i < config.DownloadRetryMax; i++ {
		status = tryDownload(download)
		if status.Status < downloadFailed || status.Status == downloadFailedCode404 { // Success or Skip
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	// Any kind of failure
	if status.Status >= downloadFailed && !download.HistoryCmd && !download.EmojiCmd {
		log.Println(lg("Download", "", color.RedString,
			"Gave up on downloading %s after %d failed attempts...\t%s",
			download.InputURL, config.DownloadRetryMax, getDownloadStatusString(status.Status)))
		if channelConfig := getSource(download.Message); channelConfig != emptyConfig {
			if !download.HistoryCmd && *channelConfig.SendErrorMessages {
				content := fmt.Sprintf(
					"Gave up trying to download\n<%s>\nafter %d failed attempts...\n\n``%s``",
					download.InputURL, config.DownloadRetryMax, getDownloadStatusString(status.Status))
				if status.Error != nil {
					content += fmt.Sprintf("\n```ERROR: %s```", status.Error)
				}
				// Failure Notice
				if !hasPerms(download.Message.ChannelID, discordgo.PermissionSendMessages) {
					log.Println(lg("Download", "", color.HiRedString, fmtBotSendPerm, download.Message.ChannelID))
				} else {
					if selfbot {
						_, err := bot.ChannelMessageSend(download.Message.ChannelID,
							fmt.Sprintf("%s **Download Failure**\n\n%s", download.Message.Author.Mention(), content))
						if err != nil {
							log.Println(lg("Download", "", color.HiRedString,
								"Failed to send failure message to %s: %s", download.Message.ChannelID, err))
						}
					} else {
						if _, err := bot.ChannelMessageSendComplex(download.Message.ChannelID,
							&discordgo.MessageSend{
								Content: fmt.Sprintf("<@!%s>", download.Message.Author.ID),
								Embed:   buildEmbed(download.Message.ChannelID, "Download Failure", content),
							}); err != nil {
							log.Println(lg("Download", "", color.HiRedString,
								"Failed to send failure message to %s: %s",
								download.Message.ChannelID, err))
						}
					}
				}
			}
			if status.Error != nil {
				sendErrorMessage(fmt.Sprintf("**%s**\n\n%s", getDownloadStatusString(status.Status), status.Error))
			}
		}
	}

	// Log Links to File
	if channelConfig := getSource(download.Message); channelConfig != emptyConfig {
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
							if download.Message.GuildID == "" {
								ch, err := bot.State.Channel(download.Message.ChannelID)
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
								logPath += " SID_" + download.Message.GuildID
							}
						}
						if *channelConfig.LogLinks.DivideLogsByChannel == true {
							logPath += " CID_" + download.Message.ChannelID
						}
						if *channelConfig.LogLinks.DivideLogsByUser == true {
							logPath += " UID_" + download.Message.Author.ID
						}
						if *channelConfig.LogLinks.DivideLogsByStatus == true {
							if status.Status >= downloadFailed {
								logPath += " - FAILED"
							} else if status.Status >= downloadSkipped {
								logPath += " - SKIPPED"
							} else if status.Status == downloadIgnored {
								logPath += " - IGNORED"
							} else if status.Status == downloadSuccess {
								logPath += " - DOWNLOADED"
							}
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
					log.Println(lg("Download", "", color.RedString,
						"[channelConfig.LogLinks] Failed to open log file:\t%s", err))
					f.Close()
				}
				defer f.Close()

				var newLine string
				shouldLog := true

				// Log Failures
				if status.Status > downloadSuccess {
					shouldLog = *channelConfig.LogLinks.LogFailures // will not log if LogFailures is false
				} else if *channelConfig.LogLinks.LogDownloads { // Log Downloads
					shouldLog = true
				}
				// Filter Duplicates
				if channelConfig.LogLinks.FilterDuplicates != nil {
					if *channelConfig.LogLinks.FilterDuplicates {
						if strings.Contains(currentLogS, download.InputURL) {
							shouldLog = false
						}
					}
				}
				if shouldLog {
					// Prepend
					prefix := ""
					if channelConfig.LogLinks.Prefix != nil {
						prefix = *channelConfig.LogLinks.Prefix
					}
					// More Data
					additionalInfo := ""
					if channelConfig.LogLinks.UserData != nil {
						if *channelConfig.LogLinks.UserData == true {
							additionalInfo = fmt.Sprintf("[%s/%s] \"%s\"#%s (%s) @ %s: ",
								download.Message.GuildID, download.Message.ChannelID,
								download.Message.Author.Username, download.Message.Author.Discriminator,
								download.Message.Author.ID, download.Message.Timestamp)
						}
					}
					// Append
					suffix := ""
					if channelConfig.LogLinks.Suffix != nil {
						suffix = *channelConfig.LogLinks.Suffix
					}
					// New Line
					newLine += "\n" + prefix + additionalInfo + download.InputURL + suffix

					if _, err = f.WriteString(newLine); err != nil {
						log.Println(lg("Download", "", color.RedString,
							"[channelConfig.LogLinks] Failed to append file:\t%s", err))
					}
				}
			}
		}
	}

	return status
}

func tryDownload(download downloadRequestStruct) downloadStatusStruct {
	var err error

	cachedDownloadID++
	thisDownloadID := cachedDownloadID

	logPrefix := ""
	if download.HistoryCmd {
		logPrefix = "HISTORY "
	}

	var channelConfig configurationSource
	channelDefault(&channelConfig)
	_channelConfig := getSource(download.Message)
	if _channelConfig != emptyConfig {
		channelConfig = _channelConfig
	}
	if _channelConfig != emptyConfig || download.EmojiCmd || download.ManualDownload {

		// Source validation
		if _, err = url.ParseRequestURI(download.InputURL); err != nil {
			return mDownloadStatus(downloadFailedInvalidSource, err)
		}

		// Clean/fix path
		if download.Path == "" || download.Path == string(os.PathSeparator) {
			log.Println(lg("Download", "", color.HiRedString, "Destination cannot be empty path..."))
			return mDownloadStatus(downloadFailedInvalidPath, err)
		}
		if !strings.HasSuffix(download.Path, string(os.PathSeparator)) {
			download.Path = download.Path + string(os.PathSeparator)
		}

		// Create folder
		if err = os.MkdirAll(download.Path, 0755); err != nil {
			log.Println(lg("Download", "", color.HiRedString,
				"Error while creating destination folder \"%s\": %s",
				download.Path, err))
			return mDownloadStatus(downloadFailedCreatingFolder, err)
		}

		// Request
		timeout := time.Duration(time.Duration(config.DownloadTimeout) * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		request, err := http.NewRequest("GET", download.InputURL, nil)
		request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.139 Safari/537.36")
		if err != nil {
			log.Println(lg("Download", "", color.HiRedString, "Error while requesting \"%s\": %s", download.InputURL, err))
			return mDownloadStatus(downloadFailedRequesting, err)
		}
		request.Header.Add("Accept-Encoding", "identity")
		response, err := client.Do(request)
		if err != nil {
			if !strings.Contains(err.Error(), "no such host") && !strings.Contains(err.Error(), "connection refused") {
				log.Println(lg("Download", "", color.HiRedString,
					"Error while receiving response from \"%s\": %s",
					download.InputURL, err))
			}
			return mDownloadStatus(downloadFailedDownloadingResponse, err)
		}
		defer response.Body.Close()

		// Read
		bodyOfResp, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(lg("Download", "", color.HiRedString,
				"Could not read response from \"%s\": %s",
				download.InputURL, err))
			return mDownloadStatus(downloadFailedReadResponse, err)
		}

		// Errors
		if response.StatusCode >= 400 {
			log.Println(lg("Download", "", color.HiRedString, logPrefix+"DOWNLOAD FAILED, %d %s: %s",
				response.StatusCode, http.StatusText(response.StatusCode), download.InputURL))
			if response.StatusCode == 404 {
				return mDownloadStatus(downloadFailedCode404, err)
			} else {
				return mDownloadStatus(downloadFailedCode, err)
			}
		}

		// Filename
		if download.Filename == "" {
			download.Filename = filenameFromURL(response.Request.URL.String())
			for key, iHeader := range response.Header {
				if key == "Content-Disposition" {
					if _, params, err := mime.ParseMediaType(iHeader[0]); err == nil {
						newFilename, err := url.QueryUnescape(params["filename"])
						if err != nil {
							newFilename = params["filename"]
						}
						if newFilename != "" {
							download.Filename = newFilename
						}
					}
				}
			}
		}

		if len(download.Filename) >= 260 {
			download.Filename = download.Filename[:255]
		}

		extension := strings.ToLower(filepath.Ext(download.Filename))
		download.FileExtension = extension

		contentType := http.DetectContentType(bodyOfResp)
		contentTypeParts := strings.Split(contentType, "/")
		contentTypeFound := contentTypeParts[0]

		parsedURL, err := url.Parse(download.InputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Error while parsing url:\t%s", err))
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
				if !download.HistoryCmd {
					log.Println(lg("Download", "Skip", color.GreenString, "Unpermitted extension (%s) found at %s", extension, download.InputURL))
				}
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
		if filepath.Ext(download.Filename) == "" {
			if possibleExtension, _ := mime.ExtensionsByType(contentType); len(possibleExtension) > 0 {
				download.Filename += possibleExtension[0]
				download.FileExtension = possibleExtension[0]
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
				if !download.HistoryCmd {
					log.Println(lg("Download", "Skip", color.GreenString,
						"Unpermitted domain (%s) found at %s", parsedURL.Hostname(), download.InputURL))
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
			if !download.HistoryCmd {
				log.Println(lg("Download", "Skip", color.GreenString,
					"Unpermitted filetype (%s) found at %s", contentTypeFound, download.InputURL))
			}
			return mDownloadStatus(downloadSkippedUnpermittedType)
		}

		// Duplicate Image Filter
		if config.FilterDuplicateImages && contentTypeFound == "image" && extension != ".gif" && extension != ".webp" {
			img, _, err := image.Decode(bytes.NewReader(bodyOfResp))
			if err != nil {
				log.Println(lg("Download", "", color.HiRedString,
					"[FilterDuplicateImages] Error converting buffer to image for hashing:\t%s", err))
			} else {
				hash, _ := duplo.CreateHash(img)
				matches := imgStore.Query(hash)
				sort.Sort(matches)
				for _, match := range matches {
					if match.Score < config.FilterDuplicateImagesThreshold {
						log.Println(lg("Download", "Skip", color.GreenString,
							"Duplicate detected (Score of %f) found at %s", match.Score, download.InputURL))
						return mDownloadStatus(downloadSkippedDetectedDuplicate)
					}
				}
				imgStore.Add(cachedDownloadID, hash)
			}
		}

		// Names
		sourceChannelName := download.Message.ChannelID
		sourceName := "UNKNOWN"
		sourceChannel, _ := bot.State.Channel(download.Message.ChannelID)
		if sourceChannel != nil {
			// Channel Naming
			if sourceChannel.Name != "" {
				sourceChannelName = "#" + sourceChannel.Name
			}
			switch sourceChannel.Type {
			case discordgo.ChannelTypeGuildText:
				// Server Naming
				if sourceChannel.GuildID != "" {
					sourceGuild, _ := bot.State.Guild(sourceChannel.GuildID)
					if sourceGuild != nil && sourceGuild.Name != "" {
						sourceName = sourceGuild.Name
					}
				}
				// Category Naming
				if sourceChannel.ParentID != "" {
					sourceParent, _ := bot.State.Channel(sourceChannel.ParentID)
					if sourceParent != nil {
						if sourceParent.Name != "" {
							sourceChannelName = sourceParent.Name + " / " + sourceChannelName
						}
					}
				}
			case discordgo.ChannelTypeDM:
				sourceName = "Direct Messages"
			case discordgo.ChannelTypeGroupDM:
				sourceName = "Group Messages"
			}
		}

		subfolder := ""
		// Subfolder Division - Year Nesting
		if *channelConfig.DivideFoldersByYear {
			year := fmt.Sprint(time.Now().Year())
			if download.Message.Author != nil {
				year = fmt.Sprint(download.Message.Timestamp.Year())
			}
			subfolderSuffix := year + string(os.PathSeparator)
			subfolder = subfolder + subfolderSuffix
			// Create folder
			if err := os.MkdirAll(download.Path+subfolder, 0755); err != nil {
				log.Println(lg("Download", "", color.HiRedString,
					"Error while creating server subfolder \"%s\": %s", download.Path, err))
				return mDownloadStatus(downloadFailedCreatingSubfolder, err)
			}
		}
		// Subfolder Division - Month Nesting
		if *channelConfig.DivideFoldersByMonth {
			year := fmt.Sprintf("%02d", time.Now().Month())
			if download.Message.Author != nil {
				year = fmt.Sprintf("%02d", download.Message.Timestamp.Month())
			}
			subfolderSuffix := year + string(os.PathSeparator)
			subfolder = subfolder + subfolderSuffix
			// Create folder
			if err := os.MkdirAll(download.Path+subfolder, 0755); err != nil {
				log.Println(lg("Download", "", color.HiRedString,
					"Error while creating server subfolder \"%s\": %s", download.Path, err))
				return mDownloadStatus(downloadFailedCreatingSubfolder, err)
			}
		}
		// Subfolder Division - Server Nesting
		if *channelConfig.DivideFoldersByServer {
			subfolderSuffix := download.Message.GuildID
			if !*channelConfig.DivideFoldersUseID && sourceName != "" && sourceName != "UNKNOWN" {
				subfolderSuffix = clearPath(sourceName)
			}
			if subfolderSuffix != "" {
				subfolderSuffix = subfolderSuffix + string(os.PathSeparator)
				subfolder = subfolder + subfolderSuffix
				// Create folder
				if err := os.MkdirAll(download.Path+subfolder, 0755); err != nil {
					log.Println(lg("Download", "", color.HiRedString,
						"Error while creating server subfolder \"%s\": %s", download.Path, err))
					return mDownloadStatus(downloadFailedCreatingSubfolder, err)
				}
			}
		}
		// Subfolder Division - Channel Nesting
		if *channelConfig.DivideFoldersByChannel {
			subfolderSuffix := download.Message.ChannelID
			if !*channelConfig.DivideFoldersUseID && sourceChannelName != "" {
				subfolderSuffix = clearPath(sourceChannelName)
			}
			if subfolderSuffix != "" {
				subfolder = subfolder + subfolderSuffix + string(os.PathSeparator)
				// Create folder
				if err := os.MkdirAll(download.Path+subfolder, 0755); err != nil {
					log.Println(lg("Download", "", color.HiRedString,
						"Error while creating channel subfolder \"%s\": %s", download.Path, err))
					return mDownloadStatus(downloadFailedCreatingSubfolder, err)
				}
			}
		}
		// Subfolder Division - User Nesting
		if *channelConfig.DivideFoldersByUser && download.Message.Author != nil {
			subfolderSuffix := download.Message.Author.ID
			if !*channelConfig.DivideFoldersUseID && download.Message.Author.Username != "" {
				subfolderSuffix = clearPath(download.Message.Author.Username + "#" +
					download.Message.Author.Discriminator)
			}
			if subfolderSuffix != "" {
				subfolder = subfolder + subfolderSuffix + string(os.PathSeparator)
				// Create folder
				if err := os.MkdirAll(download.Path+subfolder, 0755); err != nil {
					log.Println(lg("Download", "", color.HiRedString,
						"Error while creating user subfolder \"%s\": %s", download.Path, err))
					return mDownloadStatus(downloadFailedCreatingSubfolder, err)
				}
			}
		}

		// Subfolder Division - Content Type
		if *channelConfig.DivideFoldersByType {
			subfolderSuffix := contentTypeFound
			switch contentTypeFound {
			case "image":
				subfolderSuffix = "images"
			case "video":
				subfolderSuffix = "videos"
			case "application":
				subfolderSuffix = "applications"
			}
			if subfolderSuffix != "" {
				subfolder = subfolder + subfolderSuffix + string(os.PathSeparator)
				// Create folder.
				if err := os.MkdirAll(download.Path+subfolder, 0755); err != nil {
					log.Println(lg("Download", "", color.HiRedString,
						"Error while creating type subfolder \"%s\": %s", download.Path+subfolder, err))
					return mDownloadStatus(downloadFailedCreatingSubfolder, err)
				}
			}
		}

		// Format Filename
		completePath := filepath.Clean(download.Path + subfolder + dynamicKeyReplacement(channelConfig, download))

		// Check if filepath exists
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
				if !download.HistoryCmd {
					log.Println(lg("Download", "Skip", color.GreenString,
						"Matching filenames, possible duplicate? Saving \"%s\" as \"%s\" instead",
						tmpPath, completePath))
				}
			} else {
				if !download.HistoryCmd {
					log.Println(lg("Download", "Skip", color.GreenString,
						"Matching filenames, possible duplicate..."))
				}
				return mDownloadStatus(downloadSkippedDuplicate)
			}
		}

		// Write
		if *channelConfig.Save {
			if err = ioutil.WriteFile(completePath, bodyOfResp, 0644); err != nil {
				log.Println(lg("Download", "", color.HiRedString,
					"Error while writing file to disk \"%s\": %s", download.InputURL, err))
				return mDownloadStatus(downloadFailedWritingFile, err)
			}

			// Change file time
			if err = os.Chtimes(completePath, download.FileTime, download.FileTime); err != nil {
				log.Println(lg("Download", "", color.RedString,
					logPrefix+"Error while changing metadata date \"%s\": %s", download.InputURL, err))
			}

			dlColor := color.HiGreenString
			msgTimestamp := ""
			if download.HistoryCmd {
				dlColor = color.HiCyanString
				msgTimestamp = "on " + download.Message.Timestamp.Format("2006/01/02 @ 15:04:05") + " "
			}
			log.Println(lg("Download", "", dlColor,
				logPrefix+"SAVED %s sent %sin \"%s / %s\" from \"%s\" to %s (%s)",
				strings.ToUpper(contentTypeFound), msgTimestamp, sourceName, sourceChannelName,
				condenseString(download.InputURL, 50), download.Path+subfolder,
				durafmt.ParseShort(time.Since(download.StartTime)).String()))
		} else {
			log.Println(lg("Download", "", color.GreenString,
				logPrefix+"Did not save %s sent in %s#%s --- file saving disabled...",
				contentTypeFound, sourceName, sourceChannelName))
		}

		userID := botUser.ID
		if download.Message.Author != nil {
			userID = download.Message.Author.ID
		}
		// Store in db
		err = dbInsertDownload(&downloadItem{
			URL:         download.InputURL,
			Time:        time.Now(),
			Destination: completePath,
			Filename:    download.Filename,
			ChannelID:   download.Message.ChannelID,
			UserID:      userID,
		})
		if err != nil {
			log.Println(lg("Download", "", color.HiRedString, "Error writing to database: %s", err))
			return mDownloadStatus(downloadFailedWritingDatabase, err)
		}

		// React
		{
			shouldReact := config.ReactWhenDownloaded
			if channelConfig.ReactWhenDownloaded != nil {
				shouldReact = *channelConfig.ReactWhenDownloaded
			}
			if download.HistoryCmd {
				if !config.ReactWhenDownloadedHistory {
					shouldReact = false
				}
				if channelConfig.ReactWhenDownloadedHistory != nil {
					if *channelConfig.ReactWhenDownloadedHistory {
						shouldReact = true
					}
				}
			}
			if download.Message.Author != nil && shouldReact {
				reaction := ""
				if *channelConfig.ReactWhenDownloadedEmoji == "" {
					if download.Message.GuildID != "" {
						guild, err := bot.State.Guild(download.Message.GuildID)
						if err != nil {
							log.Println(lg("Download", "", color.RedString,
								"Error fetching guild state for emojis from %s: %s",
								download.Message.GuildID, err))
						} else {
							emojis := guild.Emojis
							if len(emojis) > 1 {
								for {
									rand.Seed(time.Now().UnixNano())
									chosenEmoji := emojis[rand.Intn(len(emojis))]
									formattedEmoji := chosenEmoji.APIName()
									if !chosenEmoji.Animated && !stringInSlice(formattedEmoji,
										*channelConfig.BlacklistReactEmojis) {
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
				if hasPerms(download.Message.ChannelID, discordgo.PermissionAddReactions) {
					if err = bot.MessageReactionAdd(download.Message.ChannelID, download.Message.ID, reaction); err != nil {
						log.Println(lg("Download", "", color.RedString,
							"Error adding reaction to message: %s", err))
					}
				} else {
					log.Println(lg("Download", "", color.RedString,
						"Bot does not have permission to add reactions in %s", download.Message.ChannelID))
				}
			}
		}

		// Log Media To Channel(s)
		{
			var logMediaChannels []string
			if channelConfig.SendFileToChannel != nil {
				if *channelConfig.SendFileToChannel != "" {
					logMediaChannels = append(logMediaChannels, *channelConfig.SendFileToChannel)
				}
			}
			if channelConfig.SendFileToChannels != nil {
				for _, logChannel := range *channelConfig.SendFileToChannels {
					logMediaChannels = append(logMediaChannels, logChannel)
				}
			}
			for _, logChannel := range logMediaChannels {
				if logChannel != "" {
					if hasPerms(logChannel, discordgo.PermissionSendMessages) {
						actualFile := false
						if channelConfig.SendFileDirectly != nil {
							actualFile = *channelConfig.SendFileDirectly
						}
						msg := ""
						if channelConfig.SendFileCaption != nil {
							msg = *channelConfig.SendFileCaption
							msg = channelKeyReplacement(msg, download.Message.ChannelID)
						}
						// File
						if actualFile {
							_, err := bot.ChannelMessageSendComplex(logChannel,
								&discordgo.MessageSend{
									Content: msg,
									File:    &discordgo.File{Name: download.Filename, Reader: bytes.NewReader(bodyOfResp)},
								},
							)
							if err != nil {
								log.Println(lg("Download", "", color.HiRedString,
									"File log message failed to send:\t%s", err))
							}
						} else { // Embed
							embed := &discordgo.MessageEmbed{
								Title: fmt.Sprintf("Downloaded: %s", download.Filename),
								Color: getEmbedColor(logChannel),
								Footer: &discordgo.MessageEmbedFooter{
									IconURL: projectIcon,
									Text:    fmt.Sprintf("%s v%s", projectName, projectVersion),
								},
							}
							if contentTypeFound == "image" {
								embed.Image = &discordgo.MessageEmbedImage{URL: download.InputURL}
							} else if contentTypeFound == "video" {
								embed.Video = &discordgo.MessageEmbedVideo{URL: download.InputURL}
							} else {
								embed.Description = fmt.Sprintf("Unsupported filetype: %s\n%s",
									contentTypeFound, download.InputURL)
							}
							_, err := bot.ChannelMessageSendComplex(logChannel,
								&discordgo.MessageSend{
									Content: msg,
									Embed:   embed,
								},
							)
							if err != nil {
								log.Println(lg("Download", "", color.HiRedString,
									"File log message failed to send:\t%s", err))
							}
						}
					}
				}
			}
		}

		// Update Presence
		if !download.HistoryCmd {
			timeLastUpdated = time.Now()
			if *channelConfig.UpdatePresence {
				updateDiscordPresence()
			}
		}

		// Filter Duplicate Images
		if config.FilterDuplicateImages && thisDownloadID > 0 {
			encodedStore, err := imgStore.GobEncode()
			if err != nil {
				log.Println(lg("Download", "", color.HiRedString, "Failed to encode imgStore:\t%s"))
			} else {
				f, err := os.OpenFile(imgStorePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				if err != nil {
					log.Println(lg("Download", "", color.HiRedString, "Failed to open imgStore file:\t%s"))
				}
				_, err = f.Write(encodedStore)
				if err != nil {
					log.Println(lg("Download", "", color.HiRedString, "Failed to update imgStore file:\t%s"))
				}
				err = f.Close()
				if err != nil {
					log.Println(lg("Download", "", color.HiRedString, "Failed to close imgStore file:\t%s"))
				}
			}
		}

		timeLastDownload = time.Now()
		return mDownloadStatus(downloadSuccess)
	}

	return mDownloadStatus(downloadIgnored)
}
