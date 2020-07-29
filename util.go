package main

import (
	"regexp"
	"strings"
)

const (
	runeIrcBold          = '\x02'
	runeIrcColor         = '\x03'
	runeIrcReset         = '\x0f'
	runeIrcReverse       = '\x16'
	runeIrcItalic        = '\x1d'
	runeIrcStrikethrough = '\x1e'
	runeIrcUnderline     = '\x1f'
)

var (
	rxIrcColor = regexp.MustCompile(`[` + strings.Join([]string{
		string(runeIrcBold),
		string(runeIrcUnderline),
		string(runeIrcReset),
		string(runeIrcReverse),
		string(runeIrcItalic),
		string(runeIrcStrikethrough),
	}, "") + `]|` + string(runeIrcColor) + `(\d\d?(,\d\d?)?)?`)
)

func stripIrcFormatting(text string) string {
	return rxIrcColor.ReplaceAllLiteralString(text, "")
}
