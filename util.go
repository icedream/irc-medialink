package main

import (
	"net/url"
	"regexp"
	"strings"
)

const (
	runeIrcBold      = '\x02'
	runeIrcColor     = '\x03'
	runeIrcReset     = '\x0f'
	runeIrcReverse   = '\x16'
	runeIrcItalic    = '\x1d'
	runeIrcUnderline = '\x1f'
)

var (
	rxIrcColor = regexp.MustCompile(string(runeIrcColor) + "([0-9]*(,[0-9]*)?)")
)

func stripIrcFormatting(text string) string {
	text = strings.Replace(text, string(runeIrcBold), "", -1)
	text = strings.Replace(text, string(runeIrcReset), "", -1)
	text = strings.Replace(text, string(runeIrcReverse), "", -1)
	text = strings.Replace(text, string(runeIrcItalic), "", -1)
	text = strings.Replace(text, string(runeIrcUnderline), "", -1)
	text = rxIrcColor.ReplaceAllLiteralString(text, "")
	return text
}

func getYouTubeId(uri *url.URL) string {
	u := &(*uri)
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Must be an HTTP URL
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}

	// Remove www. prefix from hostname
	if strings.HasPrefix(u.Host, "www.") {
		u.Host = u.Host[4:]
	}

	switch strings.ToLower(u.Host) {
	case "youtu.be":
		// http://youtu.be/{id}
		if s, err := url.QueryUnescape(strings.TrimLeft(u.Path, "/")); err == nil {
			return s
		} else {
			return ""
		}
	case "youtube.com":
		// http://youtube.com/watch?v={id}
		return u.Query().Get("v")
	}

	return ""
}
