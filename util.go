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

const (
	runeCTCPDelimiter      = '\x01'
	runeCTCPParamDelimiter = ' '
)

var (
	rxCTCP = regexp.MustCompile(`^` +
		string(runeCTCPDelimiter) +
		`([^` + string(runeCTCPParamDelimiter) + `]+)(` + string(runeCTCPParamDelimiter) + `.+)` + // not using \s is intentional here
		string(runeCTCPDelimiter) + `?` +
		`$`)
)

type ctcpMessage struct {
	Command string
	Params  []string
}

func (msg *ctcpMessage) String() string {
	str := string(runeCTCPDelimiter) + strings.ToUpper(msg.Command)
	if msg.Params != nil && len(msg.Params) > 0 {
		str += string(runeCTCPParamDelimiter) + strings.Join(msg.Params, string(runeCTCPParamDelimiter))
	}
	str += string(runeCTCPDelimiter)
	return str
}

func (msg *ctcpMessage) ParamLine() string {
	return strings.Join(msg.Params, string(runeCTCPParamDelimiter))
}

func parseCTCP(msg string) (parsedMessage *ctcpMessage, ok bool) {
	matches := rxCTCP.FindStringSubmatch(msg)
	if matches == nil {
		return
	}

	parsedMessage = &ctcpMessage{}

	parsedMessage.Command = matches[1]
	if len(matches) > 2 {
		parsedMessage.Params = strings.Split(matches[2], string(runeCTCPParamDelimiter))
	} else {
		parsedMessage.Params = []string{}
	}

	ok = true
	return
}
