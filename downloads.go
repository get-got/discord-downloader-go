package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"io/fs"
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
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/rivo/duplo"
	"mvdan.cc/xurls/v2"
)

type downloadedItem struct {
	MessageID   string
	URL         string
	Destination string
	Domain      string
	Filesize    int64
}

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
	downloadSkippedUnpermittedExtension
	downloadSkippedUnpermittedFilename
	downloadSkippedUnpermittedReaction
	downloadSkippedUnpermittedType
	downloadSkippedDetectedDuplicate

	downloadFailed
	downloadFailedCode
	downloadFailedCode403
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

type fileItem struct {
	Link         string
	Filename     string
	AttachmentID string
	Time         time.Time
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

func getDownloadStatus(status downloadStatus) string {
	switch status {
	case downloadSuccess:
		return "Succeeded"
	//
	case downloadIgnored:
		return "Ignored"
	//
	case downloadSkipped:
		return "Skipped"
	case downloadSkippedDuplicate:
		return "Skipped - Duplicate"
	case downloadSkippedUnpermittedDomain:
		return "Skipped - Unpermitted Domain"
	case downloadSkippedUnpermittedExtension:
		return "Skipped - Unpermitted File Extension"
	case downloadSkippedUnpermittedFilename:
		return "Skipped - Unpermitted Filename Content"
	case downloadSkippedUnpermittedReaction:
		return "Skipped - Unpermitted Message Reaction"
	case downloadSkippedUnpermittedType:
		return "Skipped - Unpermitted File Type"
	case downloadSkippedDetectedDuplicate:
		return "Skipped - Detected Duplicate"
	//
	case downloadFailed:
		return "Failed"
	case downloadFailedCode:
		return "Failed - BAD CONNECTION"
	case downloadFailedCode403:
		return "Failed - 403 UNAVAILABLE"
	case downloadFailedCode404:
		return "Failed - 404 NOT FOUND"
	case downloadFailedInvalidSource:
		return "Failed - Invalid Source"
	case downloadFailedInvalidPath:
		return "Failed - Invalid Path"
	case downloadFailedCreatingFolder:
		return "Failed - Error Creating Folder"
	case downloadFailedRequesting:
		return "Failed - Error Requesting"
	case downloadFailedDownloadingResponse:
		return "Failed - Error Downloading Data"
	case downloadFailedReadResponse:
		return "Failed - Error Reading Data"
	case downloadFailedCreatingSubfolder:
		return "Failed - Error Mapping Subfolder(s)"
	case downloadFailedWritingFile:
		return "Failed - Error Saving File"
	case downloadFailedWritingDatabase:
		return "Failed - Error Saving to Database"
	}
	return "Unknown Error"
}

func getDownloadStatusShort(status downloadStatus) string {
	if status >= downloadFailed {
		return "FAILED"
	} else if status >= downloadSkipped {
		return "SKIPPED"
	} else if status == downloadIgnored {
		return "IGNORED"
	} else if status == downloadSuccess {
		return "DOWNLOADED"
	}
	return "UNKNOWN"
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

// Trim files already downloaded and stored in database
func pruneCompletedLinks(linkList map[string]string, m *discordgo.Message) map[string]string {
	sourceConfig := getSource(m)

	newList := make(map[string]string, 0)
	for link, filename := range linkList {
		alreadyDownloaded := false
		testLink := link

		parsedURL, err := url.Parse(testLink)
		if err == nil {
			if parsedURL.Hostname() == "cdn.discordapp.com" {
				if strings.Contains(parsedURL.String(), "format=") {
					parsedURL.RawQuery = "format=" + parsedURL.Query().Get("format")
				} else {
					parsedURL.RawQuery = ""
				}
				testLink = parsedURL.String()
			}
		}

		for _, downloadedFile := range dbFindDownloadByURL(testLink) {
			if downloadedFile.ChannelID == m.ChannelID {
				alreadyDownloaded = true
			}
		}

		savePossibleDuplicates := false
		if sourceConfig.SavePossibleDuplicates != nil {
			savePossibleDuplicates = *sourceConfig.SavePossibleDuplicates
		}

		if !alreadyDownloaded || savePossibleDuplicates {
			newList[link] = filename
		} else if config.Debug {
			log.Println(lg("Download", "SKIP", color.GreenString, "Found URL has already been downloaded for this channel: %s", link))
		}
	}
	return newList
}

func getRawLinks(m *discordgo.Message) []*fileItem {
	var links []*fileItem

	// Fix source author if nil
	if m.Author == nil {
		m.Author = new(discordgo.User)
	}

	// Search Discord File Attachments
	for _, attachment := range m.Attachments {
		links = append(links, &fileItem{
			Link:         attachment.URL,
			Filename:     attachment.Filename,
			AttachmentID: attachment.ID,
		})
	}

	// Search Discord Embedded Content
	for _, embed := range m.Embeds {
		if embed.URL != "" {
			links = append(links, &fileItem{
				Link: embed.URL,
			})
		}

		// Description checking removed because it causes absolute chaos,
		// fetching every random link from the description of things like YouTube videos.
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

	// Search Detected Links
	foundLinks := xurls.Strict().FindAllString(m.Content, -1)
	for _, foundLink := range foundLinks {
		links = append(links, &fileItem{
			Link: foundLink,
		})
	}

	return links
}

func getParsedLinks(inputURL string, m *discordgo.Message) map[string]string {
	/* TODO: Download Support...
	- TikTok: Tried, once the connection is closed the cdn URL is rendered invalid
	- Facebook Photos: Tried, it doesn't preload image data, it's loaded in after. Would have to keep connection open, find alternative way to grab, or use api.
	- Facebook Videos: Previously supported but they split mp4 into separate audio and video streams
	*/

	// Twitter / X
	inputURL = strings.ReplaceAll(inputURL, "mobile.twitter", "twitter")
	inputURL = strings.ReplaceAll(inputURL, "fxtwitter.com", "twitter.com")
	inputURL = strings.ReplaceAll(inputURL, "c.vxtwitter.com", "twitter.com")
	inputURL = strings.ReplaceAll(inputURL, "vxtwitter.com", "twitter.com")
	inputURL = strings.ReplaceAll(inputURL, "//x.com", "//twitter.com")
	inputURL = strings.ReplaceAll(inputURL, ".x.com", ".twitter.com")
	if twitterConnected {
		if regexUrlTwitter.MatchString(inputURL) {
			links, err := getTwitterUrls(inputURL)
			if err != nil {
				if !strings.Contains(err.Error(), "suspended") {
					log.Println(lg("Download", "", color.RedString, "Twitter Media fetch failed for %s -- %s", inputURL, err))
				}
			} else if len(links) > 0 {
				return pruneCompletedLinks(links, m)
			}
		}
		if regexUrlTwitterStatus.MatchString(inputURL) {
			links, err := getTwitterStatusUrls(inputURL, m)
			if err != nil {
				if !strings.Contains(err.Error(), "suspended") && !strings.Contains(err.Error(), "No status found") {
					log.Println(lg("Download", "", color.RedString, "Twitter Status fetch failed for %s -- %s", inputURL, err))
				}
			} else if len(links) > 0 {
				return pruneCompletedLinks(links, m)
			}
		}
	} else if strings.Contains(inputURL, "twitter.com") {
		return pruneCompletedLinks(map[string]string{inputURL: ""}, m)
	}

	// Instagram
	if instagramConnected {
		if regexUrlInstagram.MatchString(inputURL) || regexUrlInstagramReel.MatchString(inputURL) {
			if strings.Contains(inputURL, "?") {
				inputURL = inputURL[:strings.Index(inputURL, "?")]
			}
			links, err := getInstagramUrls(inputURL, m)
			if err != nil {
				log.Println(lg("Download", "", color.RedString, "Instagram media fetch failed for %s -- %s", inputURL, err))
			} else if len(links) > 0 {
				return pruneCompletedLinks(links, m)
			}
		}
	}

	// Imgur (Legacy from DIDG)
	if regexUrlImgurSingle.MatchString(inputURL) {
		links, err := getImgurSingleUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Imgur Media fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}
	if regexUrlImgurAlbum.MatchString(inputURL) {
		links, err := getImgurAlbumUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Imgur Album fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}

	// Streamable (Legacy from DIDG)
	if regexUrlStreamable.MatchString(inputURL) {
		links, err := getStreamableUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Streamable fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}

	// Gfycat (Legacy from DIDG)
	if regexUrlGfycat.MatchString(inputURL) {
		links, err := getGfycatUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Gfycat fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}

	// Flickr (Legacy from DIDG)
	if regexUrlFlickrPhoto.MatchString(inputURL) {
		links, err := getFlickrPhotoUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Flickr Photo fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}
	if regexUrlFlickrAlbum.MatchString(inputURL) {
		links, err := getFlickrAlbumUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Flickr Album fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}
	if regexUrlFlickrAlbumShort.MatchString(inputURL) {
		links, err := getFlickrAlbumShortUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Flickr Album (short) fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}

	// Tistory (Legacy from DIDG)
	if regexUrlTistory.MatchString(inputURL) {
		links, err := getTistoryUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Tistory URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}
	if regexUrlTistoryLegacy.MatchString(inputURL) {
		links, err := getLegacyTistoryUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Legacy Tistory URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}
	// The original project has this as an option,
	if regexUrlPossibleTistorySite.MatchString(inputURL) {
		links, err := getPossibleTistorySiteUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Checking for Tistory site failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}

	// Reddit
	if regexUrlRedditPost.MatchString(inputURL) {
		links, err := getRedditPostUrls(inputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Reddit Post URL failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return pruneCompletedLinks(links, m)
		}
	}

	// Ignore Discord emojis / stickers
	if strings.HasPrefix(inputURL, "https://cdn.discordapp.com/emojis/") ||
		strings.HasPrefix(inputURL, "https://media.discordapp.net/stickers/") {
		return nil
	}

	// Try without queries
	/*parsedURL, err := url.Parse(inputURL)
	if err == nil {
		if strings.Contains(parsedURL.String(), "format=") {
			parsedURL.RawQuery = "format=" + parsedURL.Query().Get("format")
		} else {
			parsedURL.RawQuery = ""
		}
		inputURLWithoutQueries := parsedURL.String()
		if inputURLWithoutQueries != inputURL {
			return pruneCompletedLinks(getParsedLinks(inputURLWithoutQueries, m), m)
		}
	}*/

	return pruneCompletedLinks(map[string]string{inputURL: ""}, m)
}

func getLinksByMessage(m *discordgo.Message) []*fileItem {
	var fileItems []*fileItem

	linkTime := m.Timestamp

	rawLinks := getRawLinks(m)
	for _, rawLink := range rawLinks {
		downloadLinks := getParsedLinks(rawLink.Link, m)
		for link, filename := range downloadLinks {
			if rawLink.Filename != "" {
				filename = rawLink.Filename
			}

			fileItems = append(fileItems, &fileItem{
				Link:         link,
				Filename:     filename,
				Time:         linkTime,
				AttachmentID: rawLink.AttachmentID,
			})
		}
	}

	return trimDuplicateLinks(fileItems)
}

type downloadRequestStruct struct {
	InputURL       string
	Filename       string
	Extension      string
	Path           string
	Message        *discordgo.Message
	Channel        *discordgo.Channel
	FileTime       time.Time
	HistoryCmd     bool
	EmojiCmd       bool
	ManualDownload bool
	StartTime      time.Time
	AttachmentID   string
}

func (download downloadRequestStruct) handleDownload() (downloadStatusStruct, int64) {
	status := mDownloadStatus(downloadFailed)
	var tempfilesize int64 = -1
	for i := 0; i < config.DownloadRetryMax; i++ {
		status, tempfilesize = download.tryDownload()
		// Success or Skip
		if status.Status < downloadFailed || status.Status == downloadFailedCode403 || status.Status == downloadFailedCode404 {
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	// Any kind of failure
	if status.Status >= downloadFailed && !download.HistoryCmd && !download.EmojiCmd {
		log.Println(lg("Download", "", color.RedString,
			"Gave up on downloading %s after %d failed attempts...\t%s",
			download.InputURL, config.DownloadRetryMax, getDownloadStatus(status.Status)))
		if sourceConfig := getSource(download.Message); sourceConfig != emptySourceConfig {
			if !download.HistoryCmd && *sourceConfig.SendErrorMessages {
				content := fmt.Sprintf(
					"Gave up trying to download\n<%s>\nafter %d failed attempts...\n\n``%s``",
					download.InputURL, config.DownloadRetryMax, getDownloadStatus(status.Status))
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
				sendErrorMessage(fmt.Sprintf("**%s**\n\n%s", getDownloadStatus(status.Status), status.Error))
			}
		}
	}

	// Log Links to File
	if !download.EmojiCmd {
		if sourceConfig := getSource(download.Message); sourceConfig != emptySourceConfig {
			if sourceConfig.LogLinks != nil {
				if sourceConfig.LogLinks.Destination != "" {

					encounteredErrors := false
					savePath := sourceConfig.LogLinks.Destination + string(os.PathSeparator)

					// Subfolder Division - Format Subfolders
					if sourceConfig.LogLinks.Subfolders != nil {
						subfolders := []string{}
						for _, subfolder := range *sourceConfig.LogLinks.Subfolders {
							newSubfolder := dataKeys_DiscordMessage(
								dataKeys_DownloadStatus(subfolder, status, download),
								download.Message)

							// Scrub subfolder
							newSubfolder = clearSourceLogField(newSubfolder, *sourceConfig.LogLinks)

							subfolders = append(subfolders, newSubfolder)
						}

						// Subfolder Dividion - Handle Formatted Subfolders
						subpath := ""
						for _, subfolder := range subfolders {
							subpath = subpath + subfolder + string(os.PathSeparator)
							// Create folder
							if err := os.MkdirAll(filepath.Clean(savePath+subpath), 0755); err != nil {
								log.Println(lg("LogLinks", "", color.HiRedString,
									"Error while creating subfolder \"%s\": %s", savePath+subpath, err))
								encounteredErrors = true
							}
						}
						// Format Path
						savePath = filepath.Clean(savePath + subpath) // overwrite with new destination path
					}

					if !encounteredErrors {
						if _, err := os.Stat(savePath); err != nil {
							log.Println(lg("Download", "LogLinks", color.HiRedString,
								"Save path %s is invalid... %s", savePath, err))
						} else {
							// Format filename
							filename := download.Message.ChannelID + ".txt"
							if sourceConfig.LogLinks.FilenameFormat != nil {
								if *sourceConfig.LogLinks.FilenameFormat != "" {
									filename = dataKeys_DiscordMessage(
										dataKeys_DownloadStatus(*sourceConfig.LogLinks.FilenameFormat, status, download),
										download.Message)
									// if extension presumed missing
									if !strings.Contains(filename, ".") {
										filename += ".txt"
									}
								}
							}

							// Scrub filename
							filename = clearSourceLogField(filename, *sourceConfig.LogLinks)

							// Build path
							logPath := filepath.Clean(savePath + string(os.PathSeparator) + filename)

							// Format New Line
							var newLine string
							// Prepend
							prefix := ""
							if sourceConfig.LogLinks.LinePrefix != nil {
								prefix = *sourceConfig.LogLinks.LinePrefix
							}
							prefix = dataKeys_DiscordMessage(
								dataKeys_DownloadStatus(prefix, status, download),
								download.Message)

							// Append
							suffix := ""
							if sourceConfig.LogLinks.LineSuffix != nil {
								suffix = *sourceConfig.LogLinks.LineSuffix
							}
							suffix = dataKeys_DiscordMessage(
								dataKeys_DownloadStatus(suffix, status, download),
								download.Message)
							// New Line
							lineContent := download.InputURL
							if sourceConfig.LogLinks.LineContent != nil {
								lineContent = *sourceConfig.LogLinks.LineContent
							}
							// Message content
							msgContent := download.Message.Content
							if contentFmt, err := download.Message.ContentWithMoreMentionsReplaced(bot); err == nil {
								msgContent = contentFmt
							}
							keys := [][]string{
								{"{{link}}", download.InputURL},
								{"{{msgContent}}", msgContent},
							}
							if len(download.Message.Embeds) > 0 {
								keys = append(keys, [][]string{
									{"{{embedDesc}}", download.Message.Embeds[0].Description},
								}...)
							}
							for _, key := range keys {
								if strings.Contains(lineContent, key[0]) {
									lineContent = strings.ReplaceAll(lineContent, key[0], key[1])
								}
							}
							newLine += "\n" + prefix + lineContent + suffix

							// Read
							currentLog := ""
							if logfile, err := os.ReadFile(logPath); err == nil {
								currentLog = string(logfile)
							}
							canLog := true
							// Log Failures
							if status.Status > downloadSuccess {
								canLog = *sourceConfig.LogLinks.LogFailures // will not log if LogFailures is false
							} else if *sourceConfig.LogLinks.LogDownloads { // Log Downloads
								canLog = true
							}
							// Filter Duplicates
							if sourceConfig.LogLinks.FilterDuplicates != nil {
								if *sourceConfig.LogLinks.FilterDuplicates {
									if strings.Contains(currentLog, newLine) {
										canLog = false
									}
								}
							}

							if canLog {
								// Writer
								f, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
								if err != nil {
									log.Println(lg("Download", "LogLinks", color.RedString, "[sourceConfig.LogLinks] Failed to open log file:\t%s", err))
									f.Close()
								}
								defer f.Close()

								if _, err = f.WriteString(newLine); err != nil {
									log.Println(lg("Download", "LogLinks", color.RedString, "[sourceConfig.LogLinks] Failed to append file:\t%s", err))
								}
							}
						}
					}
				}
			}
		}
	}

	return status, tempfilesize
}

func (download downloadRequestStruct) tryDownload() (downloadStatusStruct, int64) {
	var err error

	cachedDownloadID++

	logPrefix := ""
	if download.HistoryCmd {
		logPrefix = "HISTORY "
	}

	var fileinfo fs.FileInfo

	var sourceConfig configurationSource
	sourceDefault(&sourceConfig)
	sourceConfigNew := emptySourceConfig
	if !download.EmojiCmd {
		sourceConfigNew = getSource(download.Message)
	}
	if sourceConfigNew != emptySourceConfig { // this looks stupid but it has a purpose.
		sourceConfig = sourceConfigNew
	}
	if sourceConfigNew != emptySourceConfig || download.EmojiCmd || download.ManualDownload {

		// Source validation
		if _, err = url.ParseRequestURI(download.InputURL); err != nil {
			return mDownloadStatus(downloadFailedInvalidSource, err), 0
		}

		// Check Domain
		parsedURL, err := url.Parse(download.InputURL)
		if err != nil {
			log.Println(lg("Download", "", color.RedString, "Error while parsing url:\t%s", err))
		}
		domain := parsedURL.Hostname()
		if sourceConfig.Filters.AllowedDomains != nil || sourceConfig.Filters.BlockedDomains != nil {
			shouldAbort := false
			if sourceConfig.Filters.AllowedDomains != nil {
				shouldAbort = true
			}

			if sourceConfig.Filters.BlockedDomains != nil {
				if stringInSlice(domain, *sourceConfig.Filters.BlockedDomains) {
					shouldAbort = true
				}
			}
			if sourceConfig.Filters.AllowedDomains != nil {
				if stringInSlice(domain, *sourceConfig.Filters.AllowedDomains) {
					shouldAbort = false
				}
			}

			// Abort
			if shouldAbort {
				if !download.HistoryCmd {
					log.Println(lg("Download", "Skip", color.GreenString,
						"Unpermitted domain (%s) found at %s", domain, download.InputURL))
				}
				return mDownloadStatus(downloadSkippedUnpermittedDomain), 0
			}
		}

		// Clean/fix path
		if download.Path == "" || download.Path == string(os.PathSeparator) {
			log.Println(lg("Download", "", color.HiRedString, "Destination cannot be empty path..."))
			return mDownloadStatus(downloadFailedInvalidPath, err), 0
		}
		if !strings.HasSuffix(download.Path, string(os.PathSeparator)) {
			download.Path = download.Path + string(os.PathSeparator)
		}

		// Create folder
		if err = os.MkdirAll(download.Path, 0755); err != nil {
			log.Println(lg("Download", "", color.HiRedString,
				"Error while creating destination folder \"%s\": %s",
				download.Path, err))
			return mDownloadStatus(downloadFailedCreatingFolder, err), 0
		}

		// Request
		timeout := time.Duration(time.Duration(config.DownloadTimeout) * time.Second)
		client := &http.Client{
			Timeout: timeout,
		}
		request, err := http.NewRequest("GET", download.InputURL, nil)
		request.Header.Set("User-Agent", sneakyUserAgent)
		if err != nil {
			log.Println(lg("Download", "", color.HiRedString, "Error while requesting \"%s\": %s", download.InputURL, err))
			return mDownloadStatus(downloadFailedRequesting, err), 0
		}
		request.Header.Add("Accept-Encoding", "identity")
		response, err := client.Do(request)
		if err != nil {
			if !strings.Contains(err.Error(), "no such host") && !strings.Contains(err.Error(), "connection refused") {
				log.Println(lg("Download", "", color.HiRedString,
					"Error while receiving response from \"%s\": %s",
					download.InputURL, err))
			}
			return mDownloadStatus(downloadFailedDownloadingResponse, err), 0
		}
		defer response.Body.Close()

		// Read
		bodyOfResp, err := io.ReadAll(response.Body)
		if err != nil {
			log.Println(lg("Download", "", color.HiRedString,
				"Could not read response from \"%s\": %s",
				download.InputURL, err))
			return mDownloadStatus(downloadFailedReadResponse, err), 0
		}

		// Errors
		if response.StatusCode >= 400 {
			// Output
			logHistoryErrors := true
			if sourceConfig.OutputHistoryStatus != nil {
				logHistoryErrors = *sourceConfig.OutputHistoryStatus
			}
			if logHistoryErrors {
				log.Println(lg("Download", "", color.HiRedString, logPrefix+"DOWNLOAD FAILED, %d %s: %s",
					response.StatusCode, http.StatusText(response.StatusCode), download.InputURL))
			}
			// Return
			if response.StatusCode == 403 {
				return mDownloadStatus(downloadFailedCode403, err), 0
			} else if response.StatusCode == 404 {
				return mDownloadStatus(downloadFailedCode404, err), 0
			} else {
				return mDownloadStatus(downloadFailedCode, err), 0
			}
		}

		// Content Type
		contentType := http.DetectContentType(bodyOfResp)
		contentTypeParts := strings.Split(contentType, "/")
		contentTypeBase := contentTypeParts[0]
		isHtml := strings.Contains(contentType, "text/html")

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

		// Check Filename
		if sourceConfig.Filters.AllowedFilenames != nil || sourceConfig.Filters.BlockedFilenames != nil {
			shouldAbort := false
			if sourceConfig.Filters.AllowedFilenames != nil {
				shouldAbort = true
			}

			if sourceConfig.Filters.BlockedFilenames != nil {
				for _, phrase := range *sourceConfig.Filters.BlockedFilenames {
					if phrase != "" && phrase != " " && strings.Contains(download.Filename, phrase) {
						shouldAbort = true
					}
				}
			}
			if sourceConfig.Filters.AllowedFilenames != nil {
				for _, phrase := range *sourceConfig.Filters.AllowedFilenames {
					if phrase != "" && phrase != " " && strings.Contains(download.Filename, phrase) {
						shouldAbort = false
					}
				}
			}

			// Abort
			if shouldAbort {
				if !download.HistoryCmd {
					log.Println(lg("Download", "Skip", color.GreenString,
						"Unpermitted filename content \"%s\"", download.Filename))
				}
				return mDownloadStatus(downloadSkippedUnpermittedFilename), 0
			}
		}

		// Check Reactions
		if sourceConfig.Filters.AllowedReactions != nil || sourceConfig.Filters.BlockedReactions != nil {
			shouldAbort := false
			if sourceConfig.Filters.AllowedReactions != nil {
				shouldAbort = true
			}

			if download.Message.Reactions != nil {
				for _, reaction := range download.Message.Reactions {
					if sourceConfig.Filters.BlockedReactions != nil {
						if stringInSlice(reaction.Emoji.ID, *sourceConfig.Filters.BlockedReactions) ||
							stringInSlice(reaction.Emoji.Name, *sourceConfig.Filters.BlockedReactions) {
							shouldAbort = true
						}
					}
					if sourceConfig.Filters.AllowedReactions != nil {
						if stringInSlice(reaction.Emoji.ID, *sourceConfig.Filters.AllowedReactions) ||
							stringInSlice(reaction.Emoji.Name, *sourceConfig.Filters.AllowedReactions) {
							shouldAbort = false
						}
					}
				}
			}

			// Abort
			if shouldAbort {
				if !download.HistoryCmd {
					log.Println(lg("Download", "Skip", color.GreenString,
						"Did not meet reaction filter criteria"))
				}
				return mDownloadStatus(downloadSkippedUnpermittedReaction), 0
			}
		}

		// Extension
		download.Extension = strings.ToLower(filepath.Ext(download.Filename))
		if filepath.Ext(download.Filename) == "" {
			if possibleExtension, _ := mime.ExtensionsByType(contentType); len(possibleExtension) > 0 {
				download.Filename += possibleExtension[0]
				download.Extension = possibleExtension[0]
			}
		}

		// Format Keys
		if !download.EmojiCmd {
			download.Filename = dataKeysDownload(sourceConfig, download)
		}

		// Scrub Filename
		download.Filename = clearSourceField(download.Filename, sourceConfig)

		// Fix filename length
		if len(download.Filename) >= 260 {
			download.Filename = download.Filename[:250]
			download.Filename += download.Extension
		}

		// Swap Extensions
		if download.Extension == ".jfif" {
			download.Extension = ".jpg"
			download.Filename = strings.ReplaceAll(download.Filename, ".jfif", ".jpg")
		}

		// Fix content type using extension
		if stringInSlice(download.Extension, []string{".mov"}) ||
			stringInSlice(download.Extension, []string{".mp4"}) ||
			stringInSlice(download.Extension, []string{".webm"}) {
			contentTypeBase = "video"
		} else if stringInSlice(download.Extension, []string{".psd"}) ||
			stringInSlice(download.Extension, []string{".nef"}) ||
			stringInSlice(download.Extension, []string{".dng"}) ||
			stringInSlice(download.Extension, []string{".tif"}) ||
			stringInSlice(download.Extension, []string{".tiff"}) {
			contentTypeBase = "image"
		}

		// Check extension
		if sourceConfig.Filters.AllowedExtensions != nil || sourceConfig.Filters.BlockedExtensions != nil {
			shouldAbort := false
			if sourceConfig.Filters.AllowedExtensions != nil {
				shouldAbort = true
			}

			if sourceConfig.Filters.BlockedExtensions != nil {
				if stringInSlice(download.Extension, *sourceConfig.Filters.BlockedExtensions) {
					shouldAbort = true
				}
			}
			if sourceConfig.Filters.AllowedExtensions != nil {
				if stringInSlice(download.Extension, *sourceConfig.Filters.AllowedExtensions) {
					shouldAbort = false
				}
			}

			// Abort
			if shouldAbort {
				if !download.HistoryCmd && !isHtml {
					log.Println(lg("Download", "Skip", color.GreenString, "Unpermitted extension (%s) found at %s",
						download.Extension, download.InputURL))
				}
				return mDownloadStatus(downloadSkippedUnpermittedExtension), 0
			}
		}

		// Check content type
		if !((*sourceConfig.SaveImages && contentTypeBase == "image") ||
			(*sourceConfig.SaveVideos && contentTypeBase == "video") ||
			(*sourceConfig.SaveAudioFiles && contentTypeBase == "audio") ||
			(*sourceConfig.SaveTextFiles && contentTypeBase == "text" && !isHtml) ||
			(*sourceConfig.SaveOtherFiles && contentTypeBase == "application")) {
			if !download.HistoryCmd && !isHtml {
				log.Println(lg("Download", "Skip", color.GreenString,
					"Unpermitted filetype (%s) found at %s", contentTypeBase, download.InputURL))
			}
			return mDownloadStatus(downloadSkippedUnpermittedType), 0
		}

		// Duplicate Image Filter
		if config.Duplo && contentTypeBase == "image" && download.Extension != ".gif" && download.Extension != ".webp" {
			img, _, err := image.Decode(bytes.NewReader(bodyOfResp))
			if err != nil {
				log.Println(lg("Duplo", "Download", color.HiRedString,
					"Error converting buffer to image for hashing:\t%s", err))
			} else {
				hash, _ := duplo.CreateHash(img)
				matches := duploCatalog.Query(hash)
				sort.Sort(matches)
				for _, match := range matches {
					if match.Score < config.DuploThreshold {
						log.Println(lg("Duplo", "Download", color.GreenString,
							"Duplicate detected (Score of %f) found at %s", match.Score, download.InputURL))
						return mDownloadStatus(downloadSkippedDetectedDuplicate), 0
					}
				}
				duploCatalog.Add(cachedDownloadID, hash)
			}
		}

		sourceName := "UNKNOWN"
		sourceChannelName := "UNKNOWN"
		if !download.EmojiCmd {
			// Names
			sourceChannel, err := bot.State.Channel(download.Message.ChannelID)
			if err != nil {
				sourceChannel, _ = bot.Channel(download.Message.ChannelID)
			}
			sourceChannelName = download.Message.ChannelID
			if sourceChannel != nil {
				// Channel Naming
				if sourceChannel.Name != "" {
					sourceChannelName = sourceChannel.Name
				}
				switch sourceChannel.Type {
				case discordgo.ChannelTypeDM:
					sourceName = "Direct Messages"
				case discordgo.ChannelTypeGroupDM:
					sourceName = "Group Messages"
				default:
					// Server Naming
					if sourceChannel.GuildID != "" {
						sourceGuild, err := bot.State.Guild(sourceChannel.GuildID)
						if err != nil {
							sourceGuild, _ = bot.Guild(sourceChannel.GuildID)
						}
						if sourceGuild != nil {
							if sourceGuild.Name != "" {
								sourceName = sourceGuild.Name
							} else {
								sourceName = sourceChannel.GuildID
							}
						}
					}
				}
			}

			// Subfolder Division - Format Subfolders
			if sourceConfig.Subfolders != nil {
				keys := [][]string{
					{"{{fileType}}",
						contentTypeBase + "s"},
				}
				subfolders := []string{}
				for _, subfolder := range *sourceConfig.Subfolders {
					fmtSubfolder := subfolder
					if strings.Contains(subfolder, "{{") && strings.Contains(subfolder, "}}") {
						for _, key := range keys {
							if strings.Contains(fmtSubfolder, key[0]) {
								fmtSubfolder = strings.ReplaceAll(fmtSubfolder, key[0], key[1])
							}
						}
						// all other keys ...
						fmtSubfolder = dataKeys_DiscordMessage(fmtSubfolder, download.Message)
					}

					// Scrub Subfolder
					fmtSubfolder = clearSourceField(fmtSubfolder, sourceConfig)

					subfolders = append(subfolders, fmtSubfolder)
				}

				// Subfolder Dividion - Handle Formatted Subfolders
				subpath := ""
				for _, subfolder := range subfolders {
					subpath = subpath + subfolder + string(os.PathSeparator)
					// Create folder
					if err := os.MkdirAll(download.Path+subpath, 0755); err != nil {
						log.Println(lg("Download", "", color.HiRedString,
							"Error while creating subfolder \"%s\": %s", download.Path+subpath, err))
						return mDownloadStatus(downloadFailedCreatingSubfolder, err), 0
					}
				}
				// Format Path
				download.Path = download.Path + subpath // overwrite with new destination path
			}
		}
		completePath := filepath.Clean(download.Path + download.Filename)

		// Check if filepath exists
		if _, err := os.Stat(completePath); err == nil {
			if *sourceConfig.SavePossibleDuplicates {
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
				if !download.HistoryCmd && !download.EmojiCmd {
					log.Println(lg("Download", "Skip", color.GreenString,
						"Matching filenames, possible duplicate? Saving \"%s\" as \"%s\" instead",
						tmpPath, completePath))
				}
			} else {
				if !download.HistoryCmd && !download.EmojiCmd {
					log.Println(lg("Download", "Skip", color.GreenString,
						"Matching filenames, possible duplicate..."))
				}
				return mDownloadStatus(downloadSkippedDuplicate), 0
			}
		}

		// Write
		if *sourceConfig.Save {
			if err = os.WriteFile(completePath, bodyOfResp, 0644); err != nil {
				log.Println(lg("Download", "", color.HiRedString,
					"Error while writing file to disk \"%s\": %s", download.InputURL, err))
				return mDownloadStatus(downloadFailedWritingFile, err), 0
			}

			// Change file time
			if err = os.Chtimes(completePath, download.FileTime, download.FileTime); err != nil {
				log.Println(lg("Download", "", color.RedString,
					logPrefix+"Error while changing metadata date \"%s\": %s", download.InputURL, err))
			}

			filesize := "unknown"
			speed := 0.0
			speedlabel := "kB/s"
			fileinfo, err = os.Stat(completePath)
			if err == nil {
				filesize = humanize.Bytes(uint64(fileinfo.Size()))
				speed = float64(fileinfo.Size() / humanize.KByte)
				if fileinfo.Size() >= humanize.MByte {
					speed = float64(fileinfo.Size() / humanize.MByte)
					speedlabel = "MB/s"
				}
			}

			dlColor := color.HiGreenString
			msgTimestamp := ""
			if download.HistoryCmd {
				dlColor = color.HiCyanString
				msgTimestamp = "on " + download.Message.Timestamp.Format("2006/01/02 @ 15:04:05") + " "
			}

			if download.EmojiCmd {
				log.Println(lg("Download", "", dlColor, "Saved emoji/sticker %s", download.Filename))
			} else {
				log.Println(lg("Download", "", dlColor,
					logPrefix+"SAVED %s sent %sin %s\n\t\t\t\t%s",
					strings.ToUpper(contentTypeBase), msgTimestamp,
					color.HiYellowString("\"%s / %s\" (%s, %s)", sourceName, sourceChannelName, download.Message.ChannelID, download.Message.ID),
					color.GreenString("> %s to \"%s%s\"\t\t%s", domain, download.Path, download.Filename,
						color.WhiteString("(%s, %s, %0.1f %s)",
							filesize, timeSinceShort(download.StartTime), speed/time.Since(download.StartTime).Seconds(), speedlabel))))
			}
		} else {
			if !download.EmojiCmd {
				log.Println(lg("Download", "", color.GreenString,
					logPrefix+"Did not save %s sent in %s#%s --- file saving disabled...",
					contentTypeBase, sourceName, sourceChannelName))
			}
		}

		userID := botUser.ID
		if !download.EmojiCmd {
			if download.Message.Author != nil {
				userID = download.Message.Author.ID
			}
		}
		// Store in db
		chID := "0"
		if !download.EmojiCmd {
			chID = download.Message.ChannelID
		}
		err = dbInsertDownload(&downloadItem{
			URL:         download.InputURL,
			Time:        time.Now(),
			Destination: completePath,
			Filename:    download.Filename,
			ChannelID:   chID,
			UserID:      userID,
		})
		if err != nil {
			log.Println(lg("Download", "", color.HiRedString, "Error writing to database: %s", err))
			return mDownloadStatus(downloadFailedWritingDatabase, err), 0
		}

		// React
		if !download.EmojiCmd {
			shouldReact := config.ReactWhenDownloaded
			if sourceConfig.ReactWhenDownloaded != nil {
				shouldReact = *sourceConfig.ReactWhenDownloaded
			}
			if download.HistoryCmd {
				if !config.ReactWhenDownloadedHistory {
					shouldReact = false
				}
				if sourceConfig.ReactWhenDownloadedHistory != nil {
					if *sourceConfig.ReactWhenDownloadedHistory {
						shouldReact = true
					}
				}
			}
			if download.Message.Author != nil && shouldReact {
				reaction := defaultReact
				if sourceConfig.ReactWhenDownloadedEmoji == nil {
					if download.Message.GuildID != "" {
						guild, err := bot.State.Guild(download.Message.GuildID)
						if err != nil {
							guild, err = bot.Guild(download.Message.GuildID)
						}
						if err != nil {
							log.Println(lg("Download", "", color.RedString,
								"Error fetching guild state for emojis from %s: %s",
								download.Message.GuildID, err))
						} else {
							emojis := guild.Emojis
							if len(emojis) > 1 {
								for {
									rand.New(rand.NewSource((time.Now().UnixNano())))
									chosenEmoji := emojis[rand.Intn(len(emojis))]
									formattedEmoji := chosenEmoji.APIName()
									if !chosenEmoji.Animated && !stringInSlice(formattedEmoji,
										*sourceConfig.BlacklistReactEmojis) {
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
					reaction = *sourceConfig.ReactWhenDownloadedEmoji
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
			if sourceConfig.SendFileToChannel != nil {
				if *sourceConfig.SendFileToChannel != "" {
					logMediaChannels = append(logMediaChannels, *sourceConfig.SendFileToChannel)
				}
			}
			if sourceConfig.SendFileToChannels != nil {
				logMediaChannels = append(logMediaChannels, *sourceConfig.SendFileToChannels...)
			}
			for _, logChannel := range logMediaChannels {
				if logChannel != "" {
					if hasPerms(logChannel, discordgo.PermissionSendMessages) {
						actualFile := false
						if sourceConfig.SendFileDirectly != nil {
							actualFile = *sourceConfig.SendFileDirectly
						}
						msg := ""
						if sourceConfig.SendFileCaption != nil {
							msg = *sourceConfig.SendFileCaption
							msg = dataKeysChannel(msg, download.Message.ChannelID)
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
							if contentTypeBase == "image" {
								embed.Image = &discordgo.MessageEmbedImage{URL: download.InputURL}
							} else if contentTypeBase == "video" {
								embed.Video = &discordgo.MessageEmbedVideo{URL: download.InputURL}
							} else {
								embed.Description = fmt.Sprintf("Unsupported filetype: %s\n%s",
									contentTypeBase, download.InputURL)
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
			if *sourceConfig.PresenceEnabled {
				go updateDiscordPresence()
			}
		}

		timeLastDownload = time.Now()
		if *sourceConfig.Save {
			return mDownloadStatus(downloadSuccess), fileinfo.Size()
		} else {
			return mDownloadStatus(downloadSuccess), 0
		}
	}

	return mDownloadStatus(downloadIgnored), 0
}
