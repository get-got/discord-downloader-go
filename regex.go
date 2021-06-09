package main

import (
	"regexp"
)

//TODO: Reddit short url ... https://redd.it/post_code

const (
	regexpFilename                = `^^[^/\\:*?"<>|]{1,150}\.[A-Za-z0-9]{1,9}$$`
	regexpUrlTwitter              = `^http(s?):\/\/pbs(-[0-9]+)?\.twimg\.com\/media\/[^\./]+\.(jpg|png)((\:[a-z]+)?)$`
	regexpUrlTwitterStatus        = `^http(s?):\/\/(www\.)?twitter\.com\/([A-Za-z0-9-_\.]+\/status\/|statuses\/|i\/web\/status\/)([0-9]+)(\/)?$`
	regexpUrlInstagram            = `^http(s?):\/\/(www\.)?instagram\.com\/p\/[^/]+\/(\?[^/]+)?$`
	regexpUrlImgurSingle          = `^http(s?):\/\/(i\.)?imgur\.com\/[A-Za-z0-9]+(\.gifv)?$`
	regexpUrlImgurAlbum           = `^http(s?):\/\/imgur\.com\/(a\/|gallery\/|r\/[^\/]+\/)[A-Za-z0-9]+(#[A-Za-z0-9]+)?$`
	regexpUrlStreamable           = `^http(s?):\/\/(www\.)?streamable\.com\/([0-9a-z]+)$`
	regexpUrlGfycat               = `^http(s?):\/\/gfycat\.com\/(gifs\/detail\/)?[A-Za-z]+$`
	regexpUrlFlickrPhoto          = `^http(s)?:\/\/(www\.)?flickr\.com\/photos\/([0-9]+)@([A-Z0-9]+)\/([0-9]+)(\/)?(\/in\/album-([0-9]+)(\/)?)?$`
	regexpUrlFlickrAlbum          = `^http(s)?:\/\/(www\.)?flickr\.com\/photos\/(([0-9]+)@([A-Z0-9]+)|[A-Za-z0-9]+)\/(albums\/(with\/)?|(sets\/)?)([0-9]+)(\/)?$`
	regexpUrlFlickrAlbumShort     = `^http(s)?:\/\/((www\.)?flickr\.com\/gp\/[0-9]+@[A-Z0-9]+\/[A-Za-z0-9]+|flic\.kr\/s\/[a-zA-Z0-9]+)$`
	regexpUrlGoogleDrive          = `^http(s?):\/\/drive\.google\.com\/file\/d\/[^/]+\/view$`
	regexpUrlGoogleDriveFolder    = `^http(s?):\/\/drive\.google\.com\/(drive\/folders\/|open\?id=)([^/]+)$`
	regexpUrlTistory              = `^http(s?):\/\/t[0-9]+\.daumcdn\.net\/cfile\/tistory\/([A-Z0-9]+?)(\?original)?$`
	regexpUrlTistoryLegacy        = `^http(s?):\/\/[a-z0-9]+\.uf\.tistory\.com\/(image|original)\/[A-Z0-9]+$`
	regexpUrlTistoryLegacyWithCDN = `^http(s)?:\/\/[0-9a-z]+.daumcdn.net\/[a-z]+\/[a-zA-Z0-9\.]+\/\?scode=mtistory&fname=http(s?)%3A%2F%2F[a-z0-9]+\.uf\.tistory\.com%2F(image|original)%2F[A-Z0-9]+$`
	regexpUrlPossibleTistorySite  = `^http(s)?:\/\/[0-9a-zA-Z\.-]+\/(m\/)?(photo\/)?[0-9]+$`
	regexpUrlRedditPost           = `^http(s?):\/\/(www\.)?reddit\.com\/r\/([0-9a-zA-Z'_]+)?\/comments\/([0-9a-zA-Z'_]+)\/?([0-9a-zA-Z'_]+)?(.*)?$`
	regexpUrlMastodonPost1        = `^http(s)?:\/\/([0-9a-zA-Z\.-]+)?\/@([0-9a-zA-Z'_]+)?\/([0-9]+)?$`
	regexpUrlMastodonPost2        = `^http(s)?:\/\/([0-9a-zA-Z\.-]+)?\/web\/statuses\/([0-9]+)?$`
)

var (
	regexFilename                *regexp.Regexp
	regexUrlTwitter              *regexp.Regexp
	regexUrlTwitterStatus        *regexp.Regexp
	regexUrlInstagram            *regexp.Regexp
	regexUrlImgurSingle          *regexp.Regexp
	regexUrlImgurAlbum           *regexp.Regexp
	regexUrlStreamable           *regexp.Regexp
	regexUrlGfycat               *regexp.Regexp
	regexUrlFlickrPhoto          *regexp.Regexp
	regexUrlFlickrAlbum          *regexp.Regexp
	regexUrlFlickrAlbumShort     *regexp.Regexp
	regexUrlGoogleDrive          *regexp.Regexp
	regexUrlGoogleDriveFolder    *regexp.Regexp
	regexUrlTistory              *regexp.Regexp
	regexUrlTistoryLegacy        *regexp.Regexp
	regexUrlTistoryLegacyWithCDN *regexp.Regexp
	regexUrlPossibleTistorySite  *regexp.Regexp
	regexUrlRedditPost           *regexp.Regexp
	regexUrlMastodonPost1        *regexp.Regexp
	regexUrlMastodonPost2        *regexp.Regexp
)

func compileRegex() error {
	var err error

	regexFilename, err = regexp.Compile(regexpFilename)
	if err != nil {
		return err
	}
	regexUrlTwitter, err = regexp.Compile(regexpUrlTwitter)
	if err != nil {
		return err
	}
	regexUrlTwitterStatus, err = regexp.Compile(regexpUrlTwitterStatus)
	if err != nil {
		return err
	}
	regexUrlInstagram, err = regexp.Compile(regexpUrlInstagram)
	if err != nil {
		return err
	}
	regexUrlImgurSingle, err = regexp.Compile(regexpUrlImgurSingle)
	if err != nil {
		return err
	}
	regexUrlImgurAlbum, err = regexp.Compile(regexpUrlImgurAlbum)
	if err != nil {
		return err
	}
	regexUrlStreamable, err = regexp.Compile(regexpUrlStreamable)
	if err != nil {
		return err
	}
	regexUrlGfycat, err = regexp.Compile(regexpUrlGfycat)
	if err != nil {
		return err
	}
	regexUrlFlickrPhoto, err = regexp.Compile(regexpUrlFlickrPhoto)
	if err != nil {
		return err
	}
	regexUrlFlickrAlbum, err = regexp.Compile(regexpUrlFlickrAlbum)
	if err != nil {
		return err
	}
	regexUrlFlickrAlbumShort, err = regexp.Compile(regexpUrlFlickrAlbumShort)
	if err != nil {
		return err
	}
	regexUrlGoogleDrive, err = regexp.Compile(regexpUrlGoogleDrive)
	if err != nil {
		return err
	}
	regexUrlGoogleDriveFolder, err = regexp.Compile(regexpUrlGoogleDriveFolder)
	if err != nil {
		return err
	}
	regexUrlTistory, err = regexp.Compile(regexpUrlTistory)
	if err != nil {
		return err
	}
	regexUrlTistoryLegacy, err = regexp.Compile(regexpUrlTistoryLegacy)
	if err != nil {
		return err
	}
	regexUrlTistoryLegacyWithCDN, err = regexp.Compile(regexpUrlTistoryLegacyWithCDN)
	if err != nil {
		return err
	}
	regexUrlPossibleTistorySite, err = regexp.Compile(regexpUrlPossibleTistorySite)
	if err != nil {
		return err
	}
	regexUrlRedditPost, err = regexp.Compile(regexpUrlRedditPost)
	if err != nil {
		return err
	}
	regexUrlMastodonPost1, err = regexp.Compile(regexpUrlMastodonPost1)
	if err != nil {
		return err
	}
	regexUrlMastodonPost2, err = regexp.Compile(regexpUrlMastodonPost2)
	if err != nil {
		return err
	}

	return nil
}
