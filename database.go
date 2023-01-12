package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/fatih/color"
)

func dbInsertDownload(download *downloadItem) error {
	_, err := myDB.Use("Downloads").Insert(map[string]interface{}{
		"URL":         download.URL,
		"Time":        download.Time.String(),
		"Destination": download.Destination,
		"Filename":    download.Filename,
		"ChannelID":   download.ChannelID,
		"UserID":      download.UserID,
	})
	return err
}

func dbFindDownloadByID(id int) *downloadItem {
	downloads := myDB.Use("Downloads")
	readBack, err := downloads.Read(id)
	if err != nil {
		log.Println(lg("Database", "Downloads", color.HiRedString, "Failed to read database:\t%s", err))
	}
	timeT, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", readBack["Time"].(string))
	return &downloadItem{
		URL:         readBack["URL"].(string),
		Time:        timeT,
		Destination: readBack["Destination"].(string),
		Filename:    readBack["Filename"].(string),
		ChannelID:   readBack["ChannelID"].(string),
		UserID:      readBack["UserID"].(string),
	}
}

func dbFindDownloadByURL(inputURL string) []*downloadItem {
	var query interface{}
	json.Unmarshal([]byte(fmt.Sprintf(`[{"eq": "%s", "in": ["URL"]}]`, inputURL)), &query)
	queryResult := make(map[int]struct{})
	db.EvalQuery(query, myDB.Use("Downloads"), &queryResult)

	downloadedImages := make([]*downloadItem, 0)
	for id := range queryResult {
		downloadedImages = append(downloadedImages, dbFindDownloadByID(id))
	}
	return downloadedImages
}

func dbDeleteByChannelID(channelID string) {
	var query interface{}
	json.Unmarshal([]byte(fmt.Sprintf(`[{"eq": "%s", "in": ["ChannelID"]}]`, channelID)), &query)
	queryResult := make(map[int]struct{})
	db.EvalQuery(query, myDB.Use("Downloads"), &queryResult)
	for id := range queryResult {
		myDB.Use("Downloads").Delete(id)
	}
}

//#region Statistics

func dbDownloadCount() int {
	i := 0
	myDB.Use("Downloads").ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		i++
		return true
	})
	return i
}

func dbDownloadCountByChannel(channelID string) int {
	var query interface{}
	json.Unmarshal([]byte(fmt.Sprintf(`[{"eq": "%s", "in": ["ChannelID"]}]`, channelID)), &query)
	queryResult := make(map[int]struct{})
	db.EvalQuery(query, myDB.Use("Downloads"), &queryResult)

	downloadedImages := make([]*downloadItem, 0)
	for id := range queryResult {
		downloadedImages = append(downloadedImages, dbFindDownloadByID(id))
	}
	return len(downloadedImages)
}

//#endregion
