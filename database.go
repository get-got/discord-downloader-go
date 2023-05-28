package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/fatih/color"
)

func backupDatabase() error {
	if err := os.MkdirAll(pathDatabaseBackups, 0755); err != nil {
		return err
	}
	file, err := os.Create(pathDatabaseBackups + string(os.PathSeparator) + time.Now().Format("2006-01-02_15-04-05.000000000") + ".zip")
	if err != nil {
		return err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	err = filepath.Walk(pathDatabaseBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Ensure that `path` is not absolute; it should not start with "/".
		// This snippet happens to work because I don't use
		// absolute paths, but ensure your real-world code
		// transforms path into a zip-root relative path.
		f, err := w.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

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
