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
	"github.com/rivo/duplo"
)

func openDatabase() {
	var openT time.Time
	var createT time.Time
	// Database
	log.Println(lg("Database", "", color.YellowString, "Opening database...\t(this can take a bit...)"))
	openT = time.Now()
	myDB, err = db.OpenDB(pathDatabaseBase)
	if err != nil {
		log.Println(lg("Database", "", color.HiRedString, "Unable to open database: %s", err))
		return
	}
	if myDB.Use("Downloads") == nil {
		log.Println(lg("Database", "Setup", color.YellowString, "Creating database, please wait..."))
		createT = time.Now()
		if err := myDB.Create("Downloads"); err != nil {
			log.Println(lg("Database", "Setup", color.HiRedString, "Error while trying to create database: %s", err))
			return
		}
		log.Println(lg("Database", "Setup", color.HiYellowString, "Created new database...\t(took %s)", timeSinceShort(createT)))
		//
		log.Println(lg("Database", "Setup", color.YellowString, "Structuring database, please wait..."))
		createT = time.Now()
		indexColumn := func(col string) {
			if err := myDB.Use("Downloads").Index([]string{col}); err != nil {
				log.Println(lg("Database", "Setup", color.HiRedString, "Unable to create index for %s: %s", col, err))
				return
			}
		}
		indexColumn("URL")
		indexColumn("ChannelID")
		indexColumn("UserID")
		log.Println(lg("Database", "Setup", color.HiYellowString, "Created database structure...\t(took %s)", timeSinceShort(createT)))
	}
	// Cache download tally
	cachedDownloadID = dbDownloadCount()
	log.Println(lg("Database", "", color.HiYellowString, "Database opened, contains %d entries...\t(took %s)", cachedDownloadID, timeSinceShort(openT)))

	// Duplo
	if config.Duplo || sourceHasDuplo {
		log.Println(lg("Duplo", "", color.HiRedString, "!!! Duplo is barely supported and may cause issues, use at your own risk..."))
		duploCatalog = duplo.New()
		if _, err := os.Stat(pathCacheDuplo); err == nil {
			log.Println(lg("Duplo", "", color.YellowString, "Opening duplo image catalog..."))
			openT = time.Now()
			storeFile, err := os.ReadFile(pathCacheDuplo)
			if err != nil {
				log.Println(lg("Duplo", "", color.HiRedString, "Error opening duplo catalog:\t%s", err))
			} else {
				err = duploCatalog.GobDecode(storeFile)
				if err != nil {
					log.Println(lg("Duplo", "", color.HiRedString, "Error decoding duplo catalog:\t%s", err))
				}
				if duploCatalog != nil {
					log.Println(lg("Duplo", "", color.HiYellowString, "Duplo catalog opened (%d)\t(took %s)", duploCatalog.Size(), timeSinceShort(openT)))
				}
			}
		}
	}
}

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

//#region Database Utility

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

//#endregion

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
