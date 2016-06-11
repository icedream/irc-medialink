package manager

import (
	"crypto/sha512"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/thoj/go-ircevent"

	cache "github.com/patrickmn/go-cache"
)

func (m *Manager) initAntiflood() {
	m.cache = cache.New(1*time.Minute, 5*time.Second)
}

func (m *Manager) TrackUrl(target string, u *url.URL) (shouldIgnore bool) {
	key := normalizeUrlAntiflood(target, u)

	if _, ok := m.cache.Get(key); ok {
		// The URL has been used recently, should ignore
		shouldIgnore = true
	} else {
		m.cache.Add(key, nil, cache.DefaultExpiration)
	}

	return
}

func (m *Manager) TrackOutput(target, t string) (shouldNotSend bool) {
	key := normalizeTextAntiflood(target, t)

	if _, ok := m.cache.Get(key); ok {
		// The URL has been used recently, should ignore
		shouldNotSend = true
	} else {
		m.cache.Add(key, nil, cache.DefaultExpiration)
	}

	return
}

func (m *Manager) AntifloodIrcConn(c *irc.Connection) *ircConnectionProxy {
	return &ircConnectionProxy{Connection: c, m: m}
}

func normalizeUrlAntiflood(target string, u *url.URL) string {
	uc := &(*u)

	// Normalize host
	uc.Host = strings.ToLower(uc.Host)
	if strings.HasPrefix(uc.Host, "www.") {
		uc.Host = uc.Host[4:]
	}

	// Normalize query
	uc.RawQuery = uc.Query().Encode()

	s := sha512.New()
	s.Write([]byte(u.String()))
	return fmt.Sprintf("LINK/%s/%X", strings.ToUpper(target), s.Sum([]byte{}))
}

func normalizeTextAntiflood(target, text string) string {
	s := sha512.New()
	s.Write([]byte(text))
	return fmt.Sprintf("TEXT/%s/%X", strings.ToUpper(target), s.Sum([]byte{}))
}

// Proxies several methods of the IRC connection in order to drop repeated messages
type ircConnectionProxy struct {
	*irc.Connection

	m *Manager
}

func (proxy *ircConnectionProxy) Action(target, message string) {
	if proxy.m.TrackOutput(target, message) {
		log.Printf("WARNING: Output antiflood triggered, dropping message for %s: %s", target, message)
		return
	}

	proxy.Connection.Action(target, message)
}

func (proxy *ircConnectionProxy) Actionf(target, format string, a ...interface{}) {
	proxy.Action(target, fmt.Sprintf(format, a...))
}

func (proxy *ircConnectionProxy) Privmsg(target, message string) {
	if proxy.m.TrackOutput(target, message) {
		log.Printf("WARNING: Output antiflood triggered, dropping message for %s: %s", target, message)
		return
	}

	proxy.Connection.Privmsg(target, message)
}

func (proxy *ircConnectionProxy) Privmsgf(target, format string, a ...interface{}) {
	proxy.Privmsg(target, fmt.Sprintf(format, a...))
}

func (proxy *ircConnectionProxy) Notice(target, message string) {
	if proxy.m.TrackOutput(target, message) {
		log.Printf("WARNING: Output antiflood triggered, dropping message for %s: %s", target, message)
		return
	}

	proxy.Connection.Notice(target, message)
}

func (proxy *ircConnectionProxy) Noticef(target, format string, a ...interface{}) {
	proxy.Notice(target, fmt.Sprintf(format, a...))
}
