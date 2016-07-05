package main

import (
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
