package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/fatih/color"
)

// Trim files already downloaded and stored in database
func trimDownloadedLinks(linkList map[string]string, channelID string) map[string]string {
	newList := make(map[string]string, 0)
	for link, filename := range linkList {
		downloadedImages := dbFindDownloadByURL(link)
		isMatched := false
		for _, downloadedImage := range downloadedImages {
			if downloadedImage.ChannelID == channelID {
				isMatched = true
			}
		}
		if isMatched == false {
			newList[link] = filename
		} else {
			log.Println(logPrefixFileSkip, color.GreenString("Found URL has already been downloaded for this channel: %s", link))
		}
	}
	return newList
}

func dbInsertDownload(download *Download) error {
	_, err := myDB.Use("Downloads").Insert(map[string]interface{}{
		"URL":          download.URL,
		"Time":         download.Time.String(),
		"Destination":  download.Destination,
		"Filename":     download.Filename,
		"ChannelID":    download.ChannelID,
		"UserID":       download.UserID,
		"SourceDomain": download.SourceDomain,
	})
	return err
}

func dbFindDownloadByID(id int) *Download {
	downloads := myDB.Use("Downloads")
	readBack, err := downloads.Read(id)
	if err != nil {
		fmt.Println(err)
	}
	timeT, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", readBack["Time"].(string))
	return &Download{
		URL:          readBack["URL"].(string),
		Time:         timeT,
		Destination:  readBack["Destination"].(string),
		Filename:     readBack["Filename"].(string),
		ChannelID:    readBack["ChannelID"].(string),
		UserID:       readBack["UserID"].(string),
		SourceDomain: readBack["SourceDomain"].(string),
	}
}

func dbFindDownloadByURL(inputURL string) []*Download {
	var query interface{}
	json.Unmarshal([]byte(fmt.Sprintf(`[{"eq": "%s", "in": ["URL"]}]`, inputURL)), &query)
	queryResult := make(map[int]struct{})
	db.EvalQuery(query, myDB.Use("Downloads"), &queryResult)

	downloadedImages := make([]*Download, 0)
	for id := range queryResult {
		downloadedImages = append(downloadedImages, dbFindDownloadByID(id))
	}
	return downloadedImages
}

func dbDownloadCount() int {
	i := 0
	myDB.Use("Downloads").ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		i++
		return true
	})
	return i
}

func dbDownloadCountByUser(userID string) int {
	var query interface{}
	json.Unmarshal([]byte(fmt.Sprintf(`[{"eq": "%s", "in": ["UserID"]}]`, userID)), &query)
	queryResult := make(map[int]struct{})
	db.EvalQuery(query, myDB.Use("Downloads"), &queryResult)

	downloadedImages := make([]*Download, 0)
	for id := range queryResult {
		downloadedImages = append(downloadedImages, dbFindDownloadByID(id))
	}
	return len(downloadedImages)
}

func dbDownloadCountByChannel(channelID string) int {
	var query interface{}
	json.Unmarshal([]byte(fmt.Sprintf(`[{"eq": "%s", "in": ["ChannelID"]}]`, channelID)), &query)
	queryResult := make(map[int]struct{})
	db.EvalQuery(query, myDB.Use("Downloads"), &queryResult)

	downloadedImages := make([]*Download, 0)
	for id := range queryResult {
		downloadedImages = append(downloadedImages, dbFindDownloadByID(id))
	}
	return len(downloadedImages)
}

func dbDownloadCountBySource(domain string) int {
	var query interface{}
	json.Unmarshal([]byte(fmt.Sprintf(`[{"eq": "%s", "in": ["SourceDomain"]}]`, domain)), &query)
	queryResult := make(map[int]struct{})
	db.EvalQuery(query, myDB.Use("Downloads"), &queryResult)

	downloadedImages := make([]*Download, 0)
	for id := range queryResult {
		downloadedImages = append(downloadedImages, dbFindDownloadByID(id))
	}
	return len(downloadedImages)
}
