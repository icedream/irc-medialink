package main

import (
	"strings"
	"sync"
)

const (
	colorBlock = "c"
)

var (
	channelModes    = map[string]string{}
	channelModeLock sync.RWMutex
)

func setChannelMode(channel string, mode rune) {
	channelModeLock.Lock()
	defer channelModeLock.Unlock()
	channel = strings.ToLower(channel)

	modes, ok := channelModes[channel]
	if !ok {
		modes = ""
	}
	for _, knownMode := range modes {
		if knownMode == mode {
			return
		}
	}
	modes += string(mode)
	channelModes[channel] = modes
}

func resetChannelModes(channel string) {
	channelModeLock.Lock()
	defer channelModeLock.Unlock()
	channel = strings.ToLower(channel)

	channelModes[channel] = ""
}

func unsetChannelMode(channel string, mode rune) {
	channelModeLock.Lock()
	defer channelModeLock.Unlock()
	channel = strings.ToLower(channel)

	modes := channelModes[channel]
	index := strings.IndexRune(modes, mode)
	if index >= 0 {
		modes = modes[0:index] + modes[index+1:]
	}
	channelModes[channel] = modes
}

func getChannelModes(channel string) string {
	channelModeLock.RLock()
	defer channelModeLock.RUnlock()
	channel = strings.ToLower(channel)
	retval, ok := channelModes[channel]
	if !ok {
		return ""
	}
	return retval
}

func hasChannelMode(channel string, mode rune) bool {
	modes := getChannelModes(channel)
	return strings.IndexRune(modes, mode) >= 0
}

func deleteChannelModes(channel string) {
	channelModeLock.Lock()
	defer channelModeLock.Unlock()
	channel = strings.ToLower(channel)
	delete(channelModes, channel)
}

func stripIrcFormattingIfChannelBlocksColors(channel string, text string) string {
	if strings.Contains(getChannelModes(channel), "c") {
		text = stripIrcFormatting(text)
	}
	return text
}
