package main

import (
	"regexp"
)

//TODO: I'm trash at regex syntax so the facebook expressions are very bad
const (
	REGEXP_FILENAME                 = `^^[^/\\:*?"<>|]{1,150}\.[A-Za-z0-9]{2,4}$$`
	REGEXP_URL_TWITTER              = `^http(s?):\/\/pbs(-[0-9]+)?\.twimg\.com\/media\/[^\./]+\.(jpg|png)((\:[a-z]+)?)$`
	REGEXP_URL_TWITTER_STATUS       = `^http(s?):\/\/(www\.)?twitter\.com\/([A-Za-z0-9-_\.]+\/status\/|statuses\/|i\/web\/status\/)([0-9]+)$`
	REGEXP_URL_INSTAGRAM            = `^http(s?):\/\/(www\.)?instagram\.com\/p\/[^/]+\/(\?[^/]+)?$`
	REGEXP_URL_FACEBOOK_VIDEO       = `^http(s?):\/\/(www\.)?facebook\.com\/(.*?)\/videos\/(.*?)?$`
	REGEXP_URL_FACEBOOK_VIDEO_WATCH = `^http(s?):\/\/(www\.)?facebook\.com\/watch\/(.*?)?$`
	REGEXP_FACEBOOK_VIDEO_HD        = `(hd_src):\"(.+?)\"`
	REGEXP_FACEBOOK_VIDEO_SD        = `(sd_src):\"(.+?)\"`
	REGEXP_URL_IMGUR_SINGLE         = `^http(s?):\/\/(i\.)?imgur\.com\/[A-Za-z0-9]+(\.gifv)?$`
	REGEXP_URL_IMGUR_ALBUM          = `^http(s?):\/\/imgur\.com\/(a\/|gallery\/|r\/[^\/]+\/)[A-Za-z0-9]+(#[A-Za-z0-9]+)?$`
	REGEXP_URL_STREAMABLE           = `^http(s?):\/\/(www\.)?streamable\.com\/([0-9a-z]+)$`
	REGEXP_URL_GFYCAT               = `^http(s?):\/\/gfycat\.com\/(gifs\/detail\/)?[A-Za-z]+$`
	REGEXP_URL_GOOGLEDRIVE          = `^http(s?):\/\/drive\.google\.com\/file\/d\/[^/]+\/view$`
	REGEXP_URL_GOOGLEDRIVE_FOLDER   = `^http(s?):\/\/drive\.google\.com\/(drive\/folders\/|open\?id=)([^/]+)$`
	REGEXP_URL_FLICKR_PHOTO         = `^http(s)?:\/\/(www\.)?flickr\.com\/photos\/([0-9]+)@([A-Z0-9]+)\/([0-9]+)(\/)?(\/in\/album-([0-9]+)(\/)?)?$`
	REGEXP_URL_FLICKR_ALBUM         = `^http(s)?:\/\/(www\.)?flickr\.com\/photos\/(([0-9]+)@([A-Z0-9]+)|[A-Za-z0-9]+)\/(albums\/(with\/)?|(sets\/)?)([0-9]+)(\/)?$`
	REGEXP_URL_FLICKR_ALBUM_SHORT   = `^http(s)?:\/\/((www\.)?flickr\.com\/gp\/[0-9]+@[A-Z0-9]+\/[A-Za-z0-9]+|flic\.kr\/s\/[a-zA-Z0-9]+)$`
)

var (
	RegexpFilename              *regexp.Regexp
	RegexpUrlTwitter            *regexp.Regexp
	RegexpUrlTwitterStatus      *regexp.Regexp
	RegexpUrlInstagram          *regexp.Regexp
	RegexpUrlFacebookVideo      *regexp.Regexp
	RegexpUrlFacebookVideoWatch *regexp.Regexp
	RegexpFacebookVideoHD       *regexp.Regexp
	RegexpFacebookVideoSD       *regexp.Regexp
	RegexpUrlImgurSingle        *regexp.Regexp
	RegexpUrlImgurAlbum         *regexp.Regexp
	RegexpUrlStreamable         *regexp.Regexp
	RegexpUrlGfycat             *regexp.Regexp
	RegexpUrlFlickrPhoto        *regexp.Regexp
	RegexpUrlFlickrAlbum        *regexp.Regexp
	RegexpUrlFlickrAlbumShort   *regexp.Regexp
)

func compileRegex() error {
	var err error

	RegexpFilename, err = regexp.Compile(REGEXP_FILENAME)
	if err != nil {
		return err
	}
	RegexpUrlTwitter, err = regexp.Compile(REGEXP_URL_TWITTER)
	if err != nil {
		return err
	}
	RegexpUrlTwitterStatus, err = regexp.Compile(REGEXP_URL_TWITTER_STATUS)
	if err != nil {
		return err
	}
	RegexpUrlInstagram, err = regexp.Compile(REGEXP_URL_INSTAGRAM)
	if err != nil {
		return err
	}
	RegexpUrlFacebookVideo, err = regexp.Compile(REGEXP_URL_FACEBOOK_VIDEO)
	if err != nil {
		return err
	}
	RegexpUrlFacebookVideoWatch, err = regexp.Compile(REGEXP_URL_FACEBOOK_VIDEO_WATCH)
	if err != nil {
		return err
	}
	RegexpFacebookVideoHD, err = regexp.Compile(REGEXP_FACEBOOK_VIDEO_HD)
	if err != nil {
		return err
	}
	RegexpFacebookVideoSD, err = regexp.Compile(REGEXP_FACEBOOK_VIDEO_SD)
	if err != nil {
		return err
	}
	RegexpUrlImgurSingle, err = regexp.Compile(REGEXP_URL_IMGUR_SINGLE)
	if err != nil {
		return err
	}
	RegexpUrlImgurAlbum, err = regexp.Compile(REGEXP_URL_IMGUR_ALBUM)
	if err != nil {
		return err
	}
	RegexpUrlStreamable, err = regexp.Compile(REGEXP_URL_STREAMABLE)
	if err != nil {
		return err
	}
	RegexpUrlGfycat, err = regexp.Compile(REGEXP_URL_GFYCAT)
	if err != nil {
		return err
	}
	RegexpUrlFlickrPhoto, err = regexp.Compile(REGEXP_URL_FLICKR_PHOTO)
	if err != nil {
		return err
	}
	RegexpUrlFlickrAlbum, err = regexp.Compile(REGEXP_URL_FLICKR_ALBUM)
	if err != nil {
		return err
	}
	RegexpUrlFlickrAlbumShort, err = regexp.Compile(REGEXP_URL_FLICKR_ALBUM_SHORT)
	if err != nil {
		return err
	}

	return nil
}
