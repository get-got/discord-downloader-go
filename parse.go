package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Davincible/goinsta/v3"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	twitterscraper "github.com/n0madic/twitter-scraper"
)

const (
	imgurClientID   = "08af502a9e70d65"
	sneakyUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36"
)

func botLoadAPIs() {
	// Twitter API
	if *config.Credentials.TwitterAuthEnabled {
		go func() {
			twitterScraper = twitterscraper.New()
			if config.Credentials.TwitterUsername != "" &&
				config.Credentials.TwitterPassword != "" {
				log.Println(lg("API", "Twitter", color.MagentaString, "Connecting..."))

				// Proxy
				twitterProxy := func(logIt bool) {
					if config.Credentials.TwitterProxy != "" {
						err := twitterScraper.SetProxy(config.Credentials.TwitterProxy)
						if logIt {
							if err != nil {
								log.Println(lg("API", "Twitter", color.HiRedString, "Error setting proxy: %s", err.Error()))
							} else {
								log.Println(lg("API", "Twitter", color.HiMagentaString, "Proxy set to "+config.Credentials.TwitterProxy))
							}
						}
					}
				}
				twitterProxy(true)

				twitterImport := func() error {
					f, err := os.Open(pathCacheTwitter)
					if err != nil {
						return err
					}
					var cookies []*http.Cookie
					err = json.NewDecoder(f).Decode(&cookies)
					if err != nil {
						return err
					}
					twitterScraper.SetCookies(cookies)
					twitterScraper.IsLoggedIn()
					_, err = twitterScraper.GetProfile("x")
					if err != nil {
						return err
					}
					return nil
				}

				twitterExport := func() error {
					cookies := twitterScraper.GetCookies()
					js, err := json.Marshal(cookies)
					if err != nil {
						return err
					}
					f, err := os.Create(pathCacheTwitter)
					if err != nil {
						return err
					}
					f.Write(js)
					return nil
				}

				// Login Loop
				twitterLoginCount := 0
			do_twitter_login:
				twitterLoginCount++
				if twitterLoginCount > 1 {
					time.Sleep(3 * time.Second)
				}

				if twitterImport() != nil {
					twitterScraper.ClearCookies()
					if err := twitterScraper.Login(config.Credentials.TwitterUsername, config.Credentials.TwitterPassword); err != nil {
						log.Println(lg("API", "Twitter", color.HiRedString, "Login Error: %s", err.Error()))
						if twitterLoginCount <= 3 {
							goto do_twitter_login
						} else {
							log.Println(lg("API", "Twitter", color.HiRedString,
								"Failed to login to Twitter (X), the bot will not fetch this media..."))
						}
					} else {
						twitterConnected = true
						defer twitterExport()
						if twitterScraper.IsLoggedIn() {
							log.Println(lg("API", "Twitter", color.HiMagentaString, fmt.Sprintf("Connected to @%s via new login", config.Credentials.TwitterUsername)))
						} else {
							log.Println(lg("API", "Twitter", color.HiRedString,
								"Scraper login seemed successful but bot is not logged in, Twitter (X) parsing may not work..."))
						}
					}
				} else {
					log.Println(lg("API", "Twitter", color.HiMagentaString,
						"Connected to @%s via cache", config.Credentials.TwitterUsername))
					twitterConnected = true
				}

				if twitterConnected {
					twitterProxy(false)
				}
			} else {
				log.Println(lg("API", "Twitter", color.MagentaString,
					"Twitter (X) login missing, the bot will not fetch this media..."))
			}
		}()
	} else {
		log.Println(lg("API", "Twitter", color.RedString,
			"TWITTER AUTHENTICATION IS DISABLED IN SETTINGS..."))
	}

	// Instagram API
	if *config.Credentials.InstagramAuthEnabled {
		go func() {
			if config.Credentials.InstagramUsername != "" &&
				config.Credentials.InstagramPassword != "" {
				log.Println(lg("API", "Instagram", color.MagentaString, "Connecting..."))

				// Proxy
				instagramProxy := func(logIt bool) {
					if config.Credentials.InstagramProxy != "" {
						insecure := false
						if config.Credentials.InstagramProxyInsecure != nil {
							insecure = *config.Credentials.InstagramProxyInsecure
						}
						forceHTTP2 := false
						if config.Credentials.InstagramProxyForceHTTP2 != nil {
							forceHTTP2 = *config.Credentials.InstagramProxyForceHTTP2
						}
						err := instagramClient.SetProxy(config.Credentials.InstagramProxy, insecure, forceHTTP2)
						if err != nil {
							log.Println(lg("API", "Instagram", color.HiRedString, "Error setting proxy: %s", err.Error()))
						} else {
							log.Println(lg("API", "Instagram", color.HiMagentaString, "Proxy set to "+config.Credentials.InstagramProxy))
						}
					}
				}
				instagramProxy(true)

				// Login Loop
				instagramLoginCount := 0
			do_instagram_login:
				instagramLoginCount++
				if instagramLoginCount > 1 {
					time.Sleep(3 * time.Second)
				}
				if instagramClient, err = goinsta.Import(pathCacheInstagram); err != nil {
					instagramClient = goinsta.New(config.Credentials.InstagramUsername, config.Credentials.InstagramPassword)
					if err := instagramClient.Login(); err != nil {
						log.Println(lg("API", "Instagram", color.HiRedString, "Login Error: %s", err.Error()))
						if instagramLoginCount <= 3 {
							goto do_instagram_login
						} else {
							log.Println(lg("API", "Instagram", color.HiRedString,
								"Failed to login to Instagram, the bot will not fetch this media..."))
						}
					} else {
						log.Println(lg("API", "Instagram", color.HiMagentaString,
							"Connected to @%s via new login", instagramClient.Account.Username))
						instagramConnected = true
						defer instagramClient.Export(pathCacheInstagram)
					}
				} else {
					log.Println(lg("API", "Instagram", color.HiMagentaString,
						"Connected to @%s via cache", instagramClient.Account.Username))
					instagramConnected = true
				}
				if instagramConnected {
					instagramProxy(false)
				}
			} else {
				log.Println(lg("API", "Instagram", color.MagentaString,
					"Instagram login missing, the bot will not fetch this media..."))
			}
		}()
	} else {
		log.Println(lg("API", "Instagram", color.RedString,
			"INSTAGRAM AUTHENTICATION IS DISABLED IN SETTINGS..."))
	}

	mainWg.Done()
}

//#region Twitter

func getTwitterUrls(inputURL string) (map[string]string, error) {
	parts := strings.Split(inputURL, ":")
	if len(parts) < 2 {
		return nil, errors.New("unable to parse Twitter URL")
	}
	return map[string]string{"https:" + parts[1] + ":orig": filenameFromURL(parts[1])}, nil
}

func getTwitterStatusUrls(inputURL string, m *discordgo.Message) (map[string]string, error) {
	if strings.Contains(inputURL, "/photo/") {
		inputURL = inputURL[:strings.Index(inputURL, "/photo/")]
	}
	if strings.Contains(inputURL, "/video/") {
		inputURL = inputURL[:strings.Index(inputURL, "/video/")]
	}

	matches := regexUrlTwitterStatus.FindStringSubmatch(inputURL)
	_, err := strconv.ParseInt(matches[4], 10, 64)
	if err != nil {
		return nil, err
	}
	tweet, err := twitterScraper.GetTweet(matches[4])
	if err != nil {
		return nil, err
	}

	links := make(map[string]string)
	for _, photo := range tweet.Photos {
		foundUrls := getDownloadLinks(photo.URL, m)
		for foundUrlKey, foundUrlValue := range foundUrls {
			links[foundUrlKey] = foundUrlValue
		}
	}
	for _, video := range tweet.Videos {
		foundUrls := getDownloadLinks(video.URL, m)
		for foundUrlKey, foundUrlValue := range foundUrls {
			links[foundUrlKey] = foundUrlValue
		}
	}
	for _, gif := range tweet.GIFs {
		foundUrls := getDownloadLinks(gif.URL, m)
		for foundUrlKey, foundUrlValue := range foundUrls {
			links[foundUrlKey] = foundUrlValue
		}
	}

	return links, nil
}

//#endregion

//#region Instagram

func getInstagramUrls(inputURL string, m *discordgo.Message) (map[string]string, error) {
	if instagramClient == nil {
		return nil, errors.New("invalid Instagram API credentials")
	}

	links := make(map[string]string)

	// fix
	shortcode := inputURL
	if strings.Contains(shortcode, ".com/p/") {
		shortcode = shortcode[strings.Index(shortcode, ".com/p/")+7:]
	}
	if strings.Contains(shortcode, ".com/reel/") {
		shortcode = shortcode[strings.Index(shortcode, ".com/reel/")+10:]
	}
	shortcode = strings.ReplaceAll(shortcode, "/", "")

	// fetch
	mediaID, err := goinsta.MediaIDFromShortID(shortcode)
	if err == nil {
		media, err := instagramClient.GetMedia(mediaID)
		if err != nil {
			return nil, err
		} else {
			postType := media.Items[0].MediaToString()
			if postType == "carousel" {
				for index, item := range media.Items[0].CarouselMedia {
					itemType := item.MediaToString()
					if itemType == "video" {
						url := item.Videos[0].URL
						links[url] = fmt.Sprintf("%s %d %s", shortcode, index, media.Items[0].User.Username)
					} else if itemType == "photo" {
						url := item.Images.GetBest()
						links[url] = fmt.Sprintf("%s %d %s", shortcode, index, media.Items[0].User.Username)
					}
				}
			} else if postType == "video" {
				url := media.Items[0].Videos[0].URL
				links[url] = fmt.Sprintf("%s %s", shortcode, media.Items[0].User.Username)
			} else if postType == "photo" {
				url := media.Items[0].Images.GetBest()
				links[url] = fmt.Sprintf("%s %s", shortcode, media.Items[0].User.Username)
			}
		}
	}

	return links, nil
}

//#endregion

//#region Imgur

func getImgurSingleUrls(url string) (map[string]string, error) {
	url = regexp.MustCompile(`(r\/[^\/]+\/)`).ReplaceAllString(url, "") // remove subreddit url
	url = strings.Replace(url, "imgur.com/", "imgur.com/download/", -1)
	url = strings.Replace(url, ".gifv", "", -1)
	return map[string]string{url: ""}, nil
}

type imgurAlbumObject struct {
	Data []struct {
		Link string
	}
}

func getImgurAlbumUrls(url string) (map[string]string, error) {
	url = regexp.MustCompile(`(#[A-Za-z0-9]+)?$`).ReplaceAllString(url, "") // remove anchor
	afterLastSlash := strings.LastIndex(url, "/")
	albumId := url[afterLastSlash+1:]
	headers := make(map[string]string)
	headers["Authorization"] = "Client-ID " + imgurClientID
	imgurAlbumObject := new(imgurAlbumObject)
	getJSONwithHeaders("https://api.imgur.com/3/album/"+albumId+"/images", imgurAlbumObject, headers)
	links := make(map[string]string)
	for _, v := range imgurAlbumObject.Data {
		links[v.Link] = ""
	}
	if len(links) <= 0 {
		return getImgurSingleUrls(url)
	}
	log.Printf("Found imgur album with %d images (url: %s)\n", len(links), url)
	return links, nil
}

//#endregion

//#region Streamable

type streamableObject struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Files  struct {
		Mp4 struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"mp4"`
		Mp4Mobile struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"mp4-mobile"`
	} `json:"files"`
	URL          string      `json:"url"`
	ThumbnailURL string      `json:"thumbnail_url"`
	Message      interface{} `json:"message"`
}

func getStreamableUrls(url string) (map[string]string, error) {
	matches := regexUrlStreamable.FindStringSubmatch(url)
	shortcode := matches[3]
	if shortcode == "" {
		return nil, errors.New("unable to get shortcode from URL")
	}
	reqUrl := fmt.Sprintf("https://api.streamable.com/videos/%s", shortcode)
	streamable := new(streamableObject)
	getJSON(reqUrl, streamable)
	if streamable.Status != 2 || streamable.Files.Mp4.URL == "" {
		return nil, errors.New("streamable object has no download candidate")
	}
	link := streamable.Files.Mp4.URL
	if !strings.HasPrefix(link, "http") {
		link = "https:" + link
	}
	links := make(map[string]string)
	links[link] = ""
	return links, nil
}

//#endregion

//#region Gfycat

type gfycatObject struct {
	GfyItem struct {
		Mp4URL string `json:"mp4Url"`
	} `json:"gfyItem"`
}

func getGfycatUrls(url string) (map[string]string, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 3 {
		return nil, errors.New("unable to parse Gfycat URL")
	}
	gfycatId := parts[len(parts)-1]
	gfycatObject := new(gfycatObject)
	getJSON("https://api.gfycat.com/v1/gfycats/"+gfycatId, gfycatObject)
	gfycatUrl := gfycatObject.GfyItem.Mp4URL
	if url == "" {
		return nil, errors.New("failed to read response from Gfycat")
	}
	return map[string]string{gfycatUrl: ""}, nil
}

//#endregion

//#region Flickr

type flickrPhotoSizeObject struct {
	Label  string `json:"label"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Source string `json:"source"`
	URL    string `json:"url"`
	Media  string `json:"media"`
}

type flickrPhotoObject struct {
	Sizes struct {
		Canblog     int                     `json:"canblog"`
		Canprint    int                     `json:"canprint"`
		Candownload int                     `json:"candownload"`
		Size        []flickrPhotoSizeObject `json:"size"`
	} `json:"sizes"`
	Stat string `json:"stat"`
}

func getFlickrUrlFromPhotoId(photoId string) string {
	reqUrl := fmt.Sprintf("https://www.flickr.com/services/rest/?format=json&nojsoncallback=1&method=%s&api_key=%s&photo_id=%s",
		"flickr.photos.getSizes", config.Credentials.FlickrApiKey, photoId)
	flickrPhoto := new(flickrPhotoObject)
	getJSON(reqUrl, flickrPhoto)
	var bestSize flickrPhotoSizeObject
	for _, size := range flickrPhoto.Sizes.Size {
		if bestSize.Label == "" {
			bestSize = size
		} else {
			if size.Width > bestSize.Width || size.Height > bestSize.Height {
				bestSize = size
			}
		}
	}
	return bestSize.Source
}

func getFlickrPhotoUrls(url string) (map[string]string, error) {
	if config.Credentials.FlickrApiKey == "" {
		return nil, errors.New("invalid Flickr API Key Set")
	}
	matches := regexUrlFlickrPhoto.FindStringSubmatch(url)
	photoId := matches[5]
	if photoId == "" {
		return nil, errors.New("unable to get Photo ID from URL")
	}
	return map[string]string{getFlickrUrlFromPhotoId(photoId): ""}, nil
}

type flickrAlbumObject struct {
	Photoset struct {
		ID        string `json:"id"`
		Primary   string `json:"primary"`
		Owner     string `json:"owner"`
		Ownername string `json:"ownername"`
		Photo     []struct {
			ID        string `json:"id"`
			Secret    string `json:"secret"`
			Server    string `json:"server"`
			Farm      int    `json:"farm"`
			Title     string `json:"title"`
			Isprimary string `json:"isprimary"`
			Ispublic  int    `json:"ispublic"`
			Isfriend  int    `json:"isfriend"`
			Isfamily  int    `json:"isfamily"`
		} `json:"photo"`
		Page    int    `json:"page"`
		PerPage int    `json:"per_page"`
		Perpage int    `json:"perpage"`
		Pages   int    `json:"pages"`
		Total   string `json:"total"`
		Title   string `json:"title"`
	} `json:"photoset"`
	Stat string `json:"stat"`
}

func getFlickrAlbumUrls(url string) (map[string]string, error) {
	if config.Credentials.FlickrApiKey == "" {
		return nil, errors.New("invalid Flickr API Key Set")
	}
	matches := regexUrlFlickrAlbum.FindStringSubmatch(url)
	if len(matches) < 10 || matches[9] == "" {
		return nil, errors.New("unable to find Flickr Album ID in URL")
	}
	albumId := matches[9]
	if albumId == "" {
		return nil, errors.New("unable to get Album ID from URL")
	}
	reqUrl := fmt.Sprintf("https://www.flickr.com/services/rest/?format=json&nojsoncallback=1&method=%s&api_key=%s&photoset_id=%s&per_page=500",
		"flickr.photosets.getPhotos", config.Credentials.FlickrApiKey, albumId)
	flickrAlbum := new(flickrAlbumObject)
	getJSON(reqUrl, flickrAlbum)
	links := make(map[string]string)
	for _, photo := range flickrAlbum.Photoset.Photo {
		links[getFlickrUrlFromPhotoId(photo.ID)] = ""
	}
	return links, nil
}

func getFlickrAlbumShortUrls(url string) (map[string]string, error) {
	result, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Error getting long URL from shortened Flickr Album URL: " + err.Error())
	}
	if regexUrlFlickrAlbum.MatchString(result.Request.URL.String()) {
		return getFlickrAlbumUrls(result.Request.URL.String())
	}
	return nil, errors.New("encountered invalid URL while trying to get long URL from short Flickr Album URL")
}

//#endregion

//#region Tistory

// getTistoryUrls downloads tistory URLs
// http://t1.daumcdn.net/cfile/tistory/[…] => http://t1.daumcdn.net/cfile/tistory/[…]
// http://t1.daumcdn.net/cfile/tistory/[…]?original => as is
func getTistoryUrls(link string) (map[string]string, error) {
	if !strings.HasSuffix(link, "?original") {
		link += "?original"
	}
	return map[string]string{link: ""}, nil
}

func getLegacyTistoryUrls(link string) (map[string]string, error) {
	link = strings.Replace(link, "/image/", "/original/", -1)
	return map[string]string{link: ""}, nil
}

func getTistoryWithCDNUrls(urlI string) (map[string]string, error) {
	parameters, _ := url.ParseQuery(urlI)
	if val, ok := parameters["fname"]; ok {
		if len(val) > 0 {
			if regexUrlTistoryLegacy.MatchString(val[0]) {
				return getLegacyTistoryUrls(val[0])
			}
		}
	}
	return nil, nil
}

func getPossibleTistorySiteUrls(url string) (map[string]string, error) {
	client := new(http.Client)
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept-Encoding", "identity")
	request.Header.Add("User-Agent", sneakyUserAgent)
	respHead, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	contentType := ""
	for headerKey, headerValue := range respHead.Header {
		if headerKey == "Content-Type" {
			contentType = headerValue[0]
		}
	}
	if !strings.Contains(contentType, "text/html") {
		return nil, nil
	}

	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept-Encoding", "identity")
	request.Header.Add("User-Agent", sneakyUserAgent)
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	var links = make(map[string]string)

	doc.Find(".article img, #content img, div[role=main] img, .section_blogview img").Each(func(i int, s *goquery.Selection) {
		foundUrl, exists := s.Attr("src")
		if exists {
			if regexUrlTistoryLegacyWithCDN.MatchString(foundUrl) {
				finalTistoryUrls, _ := getTistoryWithCDNUrls(foundUrl)
				if len(finalTistoryUrls) > 0 {
					for finalTistoryUrl := range finalTistoryUrls {
						foundFilename := s.AttrOr("filename", "")
						links[finalTistoryUrl] = foundFilename
					}
				}
			} else if regexUrlTistoryLegacy.MatchString(foundUrl) {
				finalTistoryUrls, _ := getLegacyTistoryUrls(foundUrl)
				if len(finalTistoryUrls) > 0 {
					for finalTistoryUrl := range finalTistoryUrls {
						foundFilename := s.AttrOr("filename", "")
						links[finalTistoryUrl] = foundFilename
					}
				}
			}
		}
	})

	if len(links) > 0 {
		log.Printf("[%s] Found tistory album with %d images (url: %s)\n", time.Now().Format(time.Stamp), len(links), url)
	}
	return links, nil
}

//#endregion

//#region Reddit

// This is very crude but works for now
type redditThreadObject []struct {
	Kind string `json:"kind"`
	Data struct {
		Children interface{} `json:"children"`
	} `json:"data"`
}

func getRedditPostUrls(link string) (map[string]string, error) {
	if strings.Contains(link, "?") {
		link = link[:strings.Index(link, "?")]
	}
	redditThread := new(redditThreadObject)
	headers := make(map[string]string)
	headers["Accept-Encoding"] = "identity"
	headers["User-Agent"] = sneakyUserAgent
	err := getJSONwithHeaders(link+".json", redditThread, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json from reddit post:\t%s", err)
	}

	redditPost := (*redditThread)[0].Data.Children.([]interface{})[0].(map[string]interface{})
	redditPostData := redditPost["data"].(map[string]interface{})
	if redditPostData["url_overridden_by_dest"] != nil {
		redditLink := redditPostData["url_overridden_by_dest"].(string)
		filename := fmt.Sprintf("Reddit-%s_%s %s", redditPostData["subreddit"].(string), redditPostData["id"].(string), filenameFromURL(redditLink))
		return map[string]string{redditLink: filename}, nil
	}
	return nil, nil
}

//#endregion
