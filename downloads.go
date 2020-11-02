package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"mvdan.cc/xurls/v2"
)

type Download struct {
	URL         string
	Time        time.Time
	Destination string
	Filename    string
	ChannelID   string
	UserID      string
}

type DownloadStatus int

const (
	DownloadSuccess DownloadStatus = 0

	DownloadSkippedDuplicate            DownloadStatus = 1
	DownloadSkippedUnpermittedType      DownloadStatus = 2
	DownloadSkippedUnpermittedExtension DownloadStatus = 3

	DownloadFailed                    DownloadStatus = 4
	DownloadFailedCreatingFolder      DownloadStatus = 5
	DownloadFailedRequesting          DownloadStatus = 6
	DownloadFailedDownloadingResponse DownloadStatus = 7
	DownloadFailedReadResponse        DownloadStatus = 8
	DownloadFailedCreatingSubfolder   DownloadStatus = 9
	DownloadFailedWritingFile         DownloadStatus = 10
	DownloadFailedWritingDatabase     DownloadStatus = 11
)

type DownloadStatusStruct struct {
	Status DownloadStatus
	Error  error
}

func mDownloadStatus(status DownloadStatus, _error ...error) DownloadStatusStruct {
	if len(_error) == 0 {
		return DownloadStatusStruct{
			Status: status,
			Error:  nil,
		}
	} else {
		return DownloadStatusStruct{
			Status: status,
			Error:  _error[0],
		}
	}
}

func getDownloadStatusString(status DownloadStatus) string {
	switch status {
	case DownloadSuccess:
		return "Download Succeeded"
	//
	case DownloadSkippedDuplicate:
		return "Download Skipped - Duplicate"
	case DownloadSkippedUnpermittedType:
		return "Download Skipped - Unpermitted File Type"
	case DownloadSkippedUnpermittedExtension:
		return "Download Skipped - Unpermitted File Extension"
	//
	case DownloadFailed:
		return "Download Failed"
	case DownloadFailedCreatingFolder:
		return "Download Failed - Error Creating Folder"
	case DownloadFailedRequesting:
		return "Download Failed - Error Requesting URL Data"
	case DownloadFailedDownloadingResponse:
		return "Download Failed - Error Downloading URL Response"
	case DownloadFailedReadResponse:
		return "Download Failed - Error Reading URL Response"
	case DownloadFailedCreatingSubfolder:
		return "Download Failed - Error Creating Subfolder for Type"
	case DownloadFailedWritingFile:
		return "Download Failed - Error Writing File"
	case DownloadFailedWritingDatabase:
		return "Download Failed - Error Writing to Database"
	}
	return "Unknown Error"
}

func isDiscordEmoji(link string) bool {
	// always match discord emoji URLs, eg https://cdn.discordapp.com/emojis/340989430460317707.png
	if strings.HasPrefix(link, BASE_URL_DISCORD_EMOJI) {
		return true
	}
	return false
}

var (
	timeLastUpdated time.Time
)

// Trim duplicate links in link list
func trimDuplicateLinks(FileItems []*FileItem) []*FileItem {
	var result []*FileItem
	seen := map[string]bool{}

	for _, item := range FileItems {
		if seen[item.Link] {
			continue
		}

		seen[item.Link] = true
		result = append(result, item)
	}

	return result
}

func getRawLinks(m *discordgo.Message) []*FileItem {
	var links []*FileItem

	if m.Author == nil {
		m.Author = new(discordgo.User)
	}

	for _, attachment := range m.Attachments {
		links = append(links, &FileItem{
			Link:     attachment.URL,
			Filename: attachment.Filename,
		})
	}

	foundLinks := xurls.Strict().FindAllString(m.Content, -1)
	for _, foundLink := range foundLinks {
		links = append(links, &FileItem{
			Link: foundLink,
		})
	}

	for _, embed := range m.Embeds {
		if embed.URL != "" {
			links = append(links, &FileItem{
				Link: embed.URL,
			})
		}

		// Removing for now as this causes it to try and pull shit from things like YouTube descriptions
		/*if embed.Description != "" {
			foundLinks = xurls.Strict().FindAllString(embed.Description, -1)
			for _, foundLink := range foundLinks {
				links = append(links, &FileItem{
					Link: foundLink,
				})
			}
		}*/

		if embed.Image != nil && embed.Image.URL != "" {
			links = append(links, &FileItem{
				Link: embed.Image.URL,
			})
		}

		if embed.Video != nil && embed.Video.URL != "" {
			links = append(links, &FileItem{
				Link: embed.Video.URL,
			})
		}
	}

	return links
}

func getDownloadLinks(inputURL string, channelID string) map[string]string {
	logPrefixErrorHere := color.HiRedString("[getDownloadLinks]")

	/* TODO: Support?
	- TikTok: Tried, once the connection is closed the cdn URL is rendered invalid
	- Facebook Photos: Tried, it doesn't preload image data, it's loaded in after. Would have to keep connection open, find alternative way to grab, or use api.
	*/

	if RegexpUrlTwitter.MatchString(inputURL) {
		links, err := getTwitterUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Twitter Media fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}
	if RegexpUrlTwitterStatus.MatchString(inputURL) {
		links, err := getTwitterStatusUrls(inputURL, channelID)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Twitter Status fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if RegexpUrlInstagram.MatchString(inputURL) {
		links, err := getInstagramUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Instagram fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if RegexpUrlFacebookVideo.MatchString(inputURL) || RegexpUrlFacebookVideoWatch.MatchString(inputURL) {
		links, err := getFacebookVideoUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Facebook Video fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if RegexpUrlImgurSingle.MatchString(inputURL) {
		links, err := getImgurSingleUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Imgur Media fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}
	if RegexpUrlImgurAlbum.MatchString(inputURL) {
		links, err := getImgurAlbumUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Imgur Album fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if RegexpUrlStreamable.MatchString(inputURL) {
		links, err := getStreamableUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Streamable fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if RegexpUrlGfycat.MatchString(inputURL) {
		links, err := getGfycatUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Gfycat fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if RegexpUrlFlickrPhoto.MatchString(inputURL) {
		links, err := getFlickrPhotoUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Flickr Photo fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}
	if RegexpUrlFlickrAlbum.MatchString(inputURL) {
		links, err := getFlickrAlbumUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Flickr Album fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}
	if RegexpUrlFlickrAlbumShort.MatchString(inputURL) {
		links, err := getFlickrAlbumShortUrls(inputURL)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Flickr Album (short) fetch failed for %s -- %s", inputURL, err))
		} else if len(links) > 0 {
			return trimDownloadedLinks(links, channelID)
		}
	}

	if isDiscordEmoji(inputURL) {
		log.Println(logPrefixFileSkip, color.GreenString("Skipped %s as it is a Discord emoji", inputURL))
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

func getFileLinks(m *discordgo.Message) []*FileItem {
	var fileItems []*FileItem

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

			fileItems = append(fileItems, &FileItem{
				Link:     link,
				Filename: filename,
				Time:     linkTime,
			})
		}
	}

	fileItems = trimDuplicateLinks(fileItems)

	return fileItems
}

func startDownload(inputURL string, filename string, path string, messageID string, channelID string, guildID string, userID string, fileTime time.Time, historyCmd bool) DownloadStatusStruct {
	status := mDownloadStatus(DownloadFailed)
	logPrefixErrorHere := color.HiRedString("[startDownload]")

	for i := 0; i < config.DownloadRetryMax; i++ {
		status = tryDownload(inputURL, filename, path, messageID, channelID, guildID, userID, fileTime, historyCmd)
		if status.Status < DownloadFailed { // Success or Skip
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	if status.Status >= DownloadFailed { // Any kind of failure
		log.Println(logPrefixErrorHere, color.RedString("Gave up on downloading %s", inputURL))
		if isChannelRegistered(channelID) {
			channelConfig := getChannelConfig(channelID)
			if !historyCmd && *channelConfig.ErrorMessages {
				content := fmt.Sprintf(
					"Gave up trying to download\n<%s>\nafter %d failed attempts...\n\n``%s``",
					inputURL, config.DownloadRetryMax, getDownloadStatusString(status.Status))
				if status.Error != nil {
					content = content + fmt.Sprintf("\n```ERROR: %s```", status.Error)
				}
				_, err := bot.ChannelMessageSendComplex(channelID,
					&discordgo.MessageSend{
						Content: fmt.Sprintf("<@!%s>", userID),
						Embed:   buildEmbed(channelID, "Download Failure", content),
					})
				if err != nil {
					log.Println(logPrefixErrorHere, color.HiRedString("Failed to send failure message to %s: %s", channelID, err))
				}
			}
		}
	}

	return status
}

var (
	cachedDownloadID int
)

func tryDownload(inputURL string, filename string, path string, messageID string, channelID string, guildID string, userID string, fileTime time.Time, historyCmd bool) DownloadStatusStruct {
	cachedDownloadID++
	thisDownloadID := cachedDownloadID

	startTime := time.Now()

	logPrefixErrorHere := color.HiRedString("[tryDownload]")
	if isChannelRegistered(channelID) {
		channelConfig := getChannelConfig(channelID)

		// Clean/fix path
		if !strings.HasSuffix(path, "/") {
			path = path + "/"
		}

		// Create folder
		err := os.MkdirAll(path, 755)
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Error while creating destination folder \"%s\": %s", path, err))
			return mDownloadStatus(DownloadFailedCreatingFolder, err)
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
			return mDownloadStatus(DownloadFailedRequesting, err)
		}
		request.Header.Add("Accept-Encoding", "identity")
		response, err := client.Do(request)
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Error while receiving response from \"%s\": %s", inputURL, err))
			return mDownloadStatus(DownloadFailedDownloadingResponse, err)
		}
		defer response.Body.Close()

		// Download duration
		if config.DebugOutput {
			log.Println(logPrefixDebug, color.YellowString("#%d - %s to download.", thisDownloadID, durafmt.ParseShort(time.Since(startTime)).String()))
		}
		downloadTime := time.Now()

		// Filename
		if filename == "" {
			filename = filenameFromUrl(response.Request.URL.String())
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

		// Read
		bodyOfResp, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Could not read response from \"%s\": %s", inputURL, err))
			return mDownloadStatus(DownloadFailedReadResponse, err)
		}

		contentType := http.DetectContentType(bodyOfResp)
		contentTypeParts := strings.Split(contentType, "/")
		contentTypeFound := contentTypeParts[0]

		// Check for valid filename, if not, replace with generic filename
		if !RegexpFilename.MatchString(filename) {
			filename = "InvalidFilename"
			possibleExtension, _ := mime.ExtensionsByType(contentType)
			if len(possibleExtension) > 0 {
				filename += possibleExtension[0]
			}
		}

		// Check content type
		if !((*channelConfig.SaveImages && contentTypeFound == "image") ||
			(*channelConfig.SaveVideos && contentTypeFound == "video") ||
			(*channelConfig.SaveAudioFiles && contentTypeFound == "audio") ||
			(*channelConfig.SaveTextFiles && contentTypeFound == "text") ||
			(*channelConfig.SaveOtherFiles && contentTypeFound == "application")) {
			log.Println(logPrefixFileSkip, color.GreenString("Unpermitted filetype (%s) found at %s", contentTypeFound, inputURL))
			return mDownloadStatus(DownloadSkippedUnpermittedType)
		}

		// Check extension
		extension := filepath.Ext(filename)
		if stringInSlice(extension, *channelConfig.BlacklistedExtensions) || stringInSlice(extension, []string{".com", ".net", ".org"}) {
			log.Println(logPrefixFileSkip, color.GreenString("Unpermitted extension (%s) found at %s", extension, inputURL))
			return mDownloadStatus(DownloadSkippedUnpermittedExtension)
		}

		// Subfolder division
		subfolder := ""
		if *channelConfig.DivideFoldersByType {
			switch contentTypeFound {
			case "image":
				subfolder = "images/"
			case "video":
				subfolder = "videos/"
			case "audio":
				subfolder = "audio/"
			case "text":
				subfolder = "text/"
			case "application":
				if stringInSlice(extension, []string{".mov"}) {
					contentTypeFound = "video"
					subfolder = "videos/"
				} else {
					subfolder = "applications/"
				}
			}
			if subfolder != "" {
				// Create folder.
				err := os.MkdirAll(path+subfolder, 755)
				if err != nil {
					log.Println(logPrefixErrorHere, color.HiRedString("Error while creating subfolder \"%s\": %s", path, err))
					return mDownloadStatus(DownloadFailedCreatingSubfolder, err)
				}
			}
		}

		// Format filename/path
		filenameDateFormat := config.FilenameDateFormat
		if channelConfig.OverwriteFilenameDateFormat != nil {
			if *channelConfig.OverwriteFilenameDateFormat != "" {
				filenameDateFormat = *channelConfig.OverwriteFilenameDateFormat
			}
		}
		newFilename := time.Now().Format(filenameDateFormat) + " " + filename
		//TODO: Fix -- This was causing failure due to unnecessary separator (on windows at least), I have no idea what I'm doing but changing it seems to work fine on linux
		//completePath := path + string(os.PathSeparator) + subfolder + newFilename
		completePath := path + subfolder + newFilename

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
				log.Println(color.GreenString("Matching filenames, possible duplicate? Saving \"%s\" as \"%s\" instead", tmpPath, completePath))
			} else {
				log.Println(logPrefixFileSkip, color.GreenString("Matching filenames, possible duplicate..."))
				return mDownloadStatus(DownloadSkippedDuplicate)
			}
		}

		// Write
		err = ioutil.WriteFile(completePath, bodyOfResp, 0644)
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Error while writing file to disk \"%s\": %s", inputURL, err))
			return mDownloadStatus(DownloadFailedWritingFile, err)
		}

		// Change file time
		err = os.Chtimes(completePath, fileTime, fileTime)
		if err != nil {
			log.Println(logPrefixErrorHere, color.RedString("Error while changing metadata date \"%s\": %s", inputURL, err))
		}

		// Write duration
		if config.DebugOutput {
			log.Println(logPrefixDebug, color.YellowString("#%d - %s to save.", thisDownloadID, durafmt.ParseShort(time.Since(downloadTime)).String()))
		}
		writeTime := time.Now()

		// Output
		sourceChannelName := channelID
		sourceGuildName := "N/A"
		sourceChannel, err := bot.State.Channel(channelID)
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Error fetching channel state for %s: %s", channelID, err))
		}
		if sourceChannel != nil && sourceChannel.Name != "" {
			sourceChannelName = sourceChannel.Name
			sourceGuild, _ := bot.State.Guild(sourceChannel.GuildID)
			if sourceGuild != nil && sourceGuild.Name != "" {
				sourceGuildName = sourceGuild.Name
			}
		}
		log.Println(color.HiGreenString("SAVED FILE (%s) sent in \"%s\"#%s to \"%s\"", contentTypeFound, sourceGuildName, sourceChannelName, completePath))

		// Store in db
		err = dbInsertDownload(&Download{
			URL:         inputURL,
			Time:        time.Now(),
			Destination: completePath,
			Filename:    filename,
			ChannelID:   channelID,
			UserID:      userID,
		})
		if err != nil {
			log.Println(logPrefixErrorHere, color.HiRedString("Error writing to database: %s", err))
			return mDownloadStatus(DownloadFailedWritingDatabase, err)
		}

		// Storage & output duration
		if config.DebugOutput {
			log.Println(logPrefixDebug, color.YellowString("#%d - %s to update database.", thisDownloadID, durafmt.ParseShort(time.Since(writeTime)).String()))
		}
		finishTime := time.Now()

		// React
		if !historyCmd && *channelConfig.ReactWhenDownloaded {
			reaction := ""
			if *channelConfig.ReactWhenDownloadedEmoji == "" {
				guild, err := bot.State.Guild(guildID)
				if err != nil {
					log.Println(logPrefixErrorHere, color.RedString("Error fetching guild state for emojis from %s: %s", guildID, err))
				} else {
					emojis := guild.Emojis
					for {
						rand.Seed(time.Now().UnixNano())
						chosen_emoji := emojis[rand.Intn(len(emojis))]
						emoji_fmt := chosen_emoji.APIName()
						if !chosen_emoji.Animated && !stringInSlice(emoji_fmt, *channelConfig.BlacklistReactEmojis) {
							reaction = emoji_fmt
							break
						}
					}
				}
			} else {
				reaction = *channelConfig.ReactWhenDownloadedEmoji
			}
			bot.MessageReactionAdd(channelID, messageID, reaction)
			// React duration
			if config.DebugOutput {
				log.Println(logPrefixDebug, color.YellowString("#%d - %s to react with \"%s\".", thisDownloadID, durafmt.ParseShort(time.Since(finishTime)).String(), reaction))
			}
			finishTime = time.Now()
		}

		timeLastUpdated = time.Now()
		if *channelConfig.UpdatePresence {
			updateDiscordPresence()
			// Presence duration
			/*if config.DebugOutput {
				log.Println(logPrefixDebug, color.YellowString("#%d - %s to update presence.", thisDownloadID, durafmt.ParseShort(time.Since(finishTime)).String()))
			}*/
		}

		if config.DebugOutput {
			log.Println(logPrefixDebug, color.YellowString("#%d - %s total.", thisDownloadID, time.Since(startTime)))
		}

		return mDownloadStatus(DownloadSuccess)
	}
	return mDownloadStatus(DownloadFailed)
}
