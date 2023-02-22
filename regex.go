package main

import (
	"regexp"
)

//TODO: Reddit short url ... https://redd.it/post_code

const (
	regexpUrlTwitter              = `^http(s?):\/\/pbs(-[0-9]+)?\.twimg\.com\/media\/[^\./]+\.(jpg|png)((\:[a-z]+)?)$`
	regexpUrlTwitterStatus        = `^http(s?):\/\/(www\.)?twitter\.com\/([A-Za-z0-9-_\.]+\/status\/|statuses\/|i\/web\/status\/)([0-9]+)$`
	regexpUrlInstagram            = `^http(s?):\/\/(www\.)?instagram\.com\/p\/[^/]+\/(\?[^/]+)?$`
	regexpUrlInstagramReel        = `^http(s?):\/\/(www\.)?instagram\.com\/reel\/[^/]+\/(\?[^/]+)?$`
	regexpUrlImgurSingle          = `^http(s?):\/\/(i\.)?imgur\.com\/[A-Za-z0-9]+(\.gifv)?$`
	regexpUrlImgurAlbum           = `^http(s?):\/\/imgur\.com\/(a\/|gallery\/|r\/[^\/]+\/)[A-Za-z0-9]+(#[A-Za-z0-9]+)?$`
	regexpUrlStreamable           = `^http(s?):\/\/(www\.)?streamable\.com\/([0-9a-z]+)$`
	regexpUrlGfycat               = `^http(s?):\/\/gfycat\.com\/(gifs\/detail\/)?[A-Za-z]+$`
	regexpUrlFlickrPhoto          = `^http(s)?:\/\/(www\.)?flickr\.com\/photos\/([0-9]+)@([A-Z0-9]+)\/([0-9]+)(\/)?(\/in\/album-([0-9]+)(\/)?)?$`
	regexpUrlFlickrAlbum          = `^http(s)?:\/\/(www\.)?flickr\.com\/photos\/(([0-9]+)@([A-Z0-9]+)|[A-Za-z0-9]+)\/(albums\/(with\/)?|(sets\/)?)([0-9]+)(\/)?$`
	regexpUrlFlickrAlbumShort     = `^http(s)?:\/\/((www\.)?flickr\.com\/gp\/[0-9]+@[A-Z0-9]+\/[A-Za-z0-9]+|flic\.kr\/s\/[a-zA-Z0-9]+)$`
	regexpUrlTistory              = `^http(s?):\/\/t[0-9]+\.daumcdn\.net\/cfile\/tistory\/([A-Z0-9]+?)(\?original)?$`
	regexpUrlTistoryLegacy        = `^http(s?):\/\/[a-z0-9]+\.uf\.tistory\.com\/(image|original)\/[A-Z0-9]+$`
	regexpUrlTistoryLegacyWithCDN = `^http(s)?:\/\/[0-9a-z]+.daumcdn.net\/[a-z]+\/[a-zA-Z0-9\.]+\/\?scode=mtistory&fname=http(s?)%3A%2F%2F[a-z0-9]+\.uf\.tistory\.com%2F(image|original)%2F[A-Z0-9]+$`
	regexpUrlPossibleTistorySite  = `^http(s)?:\/\/[0-9a-zA-Z\.-]+\/(m\/)?(photo\/)?[0-9]+$`
	regexpUrlRedditPost           = `^http(s?):\/\/(www\.)?reddit\.com\/r\/([0-9a-zA-Z'_]+)?\/comments\/([0-9a-zA-Z'_]+)\/?([0-9a-zA-Z'_]+)?(.*)?$`
)

var (
	regexUrlTwitter              *regexp.Regexp
	regexUrlTwitterStatus        *regexp.Regexp
	regexUrlInstagram            *regexp.Regexp
	regexUrlInstagramReel        *regexp.Regexp
	regexUrlImgurSingle          *regexp.Regexp
	regexUrlImgurAlbum           *regexp.Regexp
	regexUrlStreamable           *regexp.Regexp
	regexUrlGfycat               *regexp.Regexp
	regexUrlFlickrPhoto          *regexp.Regexp
	regexUrlFlickrAlbum          *regexp.Regexp
	regexUrlFlickrAlbumShort     *regexp.Regexp
	regexUrlTistory              *regexp.Regexp
	regexUrlTistoryLegacy        *regexp.Regexp
	regexUrlTistoryLegacyWithCDN *regexp.Regexp
	regexUrlPossibleTistorySite  *regexp.Regexp
	regexUrlRedditPost           *regexp.Regexp
)

func compileRegex() error {
	var err error

	if regexUrlTwitter, err = regexp.Compile(regexpUrlTwitter); err != nil {
		return err
	}
	if regexUrlTwitterStatus, err = regexp.Compile(regexpUrlTwitterStatus); err != nil {
		return err
	}
	if regexUrlInstagram, err = regexp.Compile(regexpUrlInstagram); err != nil {
		return err
	}
	if regexUrlInstagramReel, err = regexp.Compile(regexpUrlInstagramReel); err != nil {
		return err
	}
	if regexUrlImgurSingle, err = regexp.Compile(regexpUrlImgurSingle); err != nil {
		return err
	}
	if regexUrlImgurAlbum, err = regexp.Compile(regexpUrlImgurAlbum); err != nil {
		return err
	}
	if regexUrlStreamable, err = regexp.Compile(regexpUrlStreamable); err != nil {
		return err
	}
	if regexUrlGfycat, err = regexp.Compile(regexpUrlGfycat); err != nil {
		return err
	}
	if regexUrlFlickrPhoto, err = regexp.Compile(regexpUrlFlickrPhoto); err != nil {
		return err
	}
	if regexUrlFlickrAlbum, err = regexp.Compile(regexpUrlFlickrAlbum); err != nil {
		return err
	}
	if regexUrlFlickrAlbumShort, err = regexp.Compile(regexpUrlFlickrAlbumShort); err != nil {
		return err
	}
	if regexUrlTistory, err = regexp.Compile(regexpUrlTistory); err != nil {
		return err
	}
	if regexUrlTistoryLegacy, err = regexp.Compile(regexpUrlTistoryLegacy); err != nil {
		return err
	}
	if regexUrlTistoryLegacyWithCDN, err = regexp.Compile(regexpUrlTistoryLegacyWithCDN); err != nil {
		return err
	}
	if regexUrlPossibleTistorySite, err = regexp.Compile(regexpUrlPossibleTistorySite); err != nil {
		return err
	}
	if regexUrlRedditPost, err = regexp.Compile(regexpUrlRedditPost); err != nil {
		return err
	}

	return nil
}
