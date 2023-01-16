package manager

import (
	"crypto/sha512"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	irc "github.com/thoj/go-ircevent"
	"golang.org/x/net/idna"

	"github.com/icedream/irc-medialink/util/clone"
)

func (m *Manager) initAntiflood() {
	m.cache = cache.New(1*time.Minute, 5*time.Second)
}

func (m *Manager) TrackUser(target string, source string) (shouldIgnore bool) {
	key := normalizeUserAntiflood(target, source)

	if _, ok := m.cache.Get(key); ok {
		// User just joined here recently, ignore them
		shouldIgnore = true
	}

	return
}

func (m *Manager) NotifyUserJoined(target string, source string) error {
	key := normalizeUserAntiflood(target, source)

	// When a user joins, he will be ignored for the first 30 seconds,
	// enough to prevent parsing links from people who only join to spam their
	// links immediately
	if _, exists := m.cache.Get(key); !exists {
		return m.cache.Add(key, nil, 30*time.Second)
	}

	return nil
}

func (m *Manager) TrackUrl(target string, u *url.URL) (shouldIgnore bool, err error) {
	key := normalizeUrlAntiflood(target, u)

	if _, ok := m.cache.Get(key); ok {
		// The URL has been used recently, should ignore
		shouldIgnore = true
	} else {
		err = m.cache.Add(key, nil, cache.DefaultExpiration)
	}

	return
}

func (m *Manager) TrackOutput(target, t string) (shouldNotSend bool, err error) {
	key := normalizeTextAntiflood(target, t)

	if _, ok := m.cache.Get(key); ok {
		// The URL has been used recently, should ignore
		shouldNotSend = true
	} else {
		err = m.cache.Add(key, nil, cache.DefaultExpiration)
	}

	return
}

func (m *Manager) AntifloodIrcConn(c *irc.Connection) *ircConnectionProxy {
	return &ircConnectionProxy{Connection: c, m: m}
}

func normalizeUrlAntiflood(target string, u *url.URL) string {
	uc := clone.CloneURL(u)

	// Normalize hostname punycode.
	//
	// Ignoring the error is correct here since we still want to work with a
	// partially converted hostname. See idna docs.
	uc.Host, _ = idna.Punycode.ToASCII(uc.Host)

	// Normalize hostname casing
	uc.Host = strings.ToLower(uc.Host)

	// Normalize query
	uc.RawQuery = uc.Query().Encode()

	// Normalize scheme
	uc.Scheme = strings.ToLower(uc.Scheme)

	// Fill in default ports if none were passed in this URL
	if len(uc.Port()) == 0 {
		uc.Host = strings.TrimSuffix(uc.Host, ":")
		switch uc.Scheme {
		case "http":
			uc.Host += ":80"
		case "https":
			uc.Host += ":443"
		}
	}

	s := sha512.New()
	s.Write([]byte(uc.String()))
	return fmt.Sprintf("LINK/%s/%X", strings.ToUpper(target), s.Sum([]byte{}))
}

func normalizeTextAntiflood(target, text string) string {
	s := sha512.New()
	s.Write([]byte(text))
	return fmt.Sprintf("TEXT/%s/%X", strings.ToUpper(target), s.Sum([]byte{}))
}

func normalizeUserAntiflood(target, source string) string {
	sourceSplitHost := strings.SplitN(source, "@", 2)
	if len(sourceSplitHost) > 1 {
		sourceSplitHostname := strings.Split(sourceSplitHost[1], ".")
		if len(sourceSplitHostname) > 1 &&
			strings.EqualFold(sourceSplitHostname[len(sourceSplitHostname)-1], "IP") {
			sourceSplitHostname[0] = "*"
		}
		source = fmt.Sprintf("%s!%s@%s", "*", "*", strings.Join(sourceSplitHostname, "."))
	}
	return fmt.Sprintf("USER/%s/%s", strings.ToUpper(target), source)
}

// Proxies several methods of the IRC connection in order to drop repeated messages
type ircConnectionProxy struct {
	*irc.Connection

	m *Manager
}

func (proxy *ircConnectionProxy) Action(target, message string) {
	if shouldNotSend, err := proxy.m.TrackOutput(target, message); err != nil {
		log.Printf("WARNING: Output antiflood returned an error, dropping message for %s: %s", target, err.Error())
		return
	} else if shouldNotSend {
		log.Printf("WARNING: Output antiflood triggered, dropping message for %s: %s", target, message)
		return
	}

	proxy.Connection.Action(target, message)
}

func (proxy *ircConnectionProxy) Actionf(target, format string, a ...interface{}) {
	proxy.Action(target, fmt.Sprintf(format, a...))
}

func (proxy *ircConnectionProxy) Privmsg(target, message string) {
	if shouldNotSend, err := proxy.m.TrackOutput(target, message); err != nil {
		log.Printf("WARNING: Output antiflood returned an error, dropping message for %s: %s", target, err.Error())
		return
	} else if shouldNotSend {
		log.Printf("WARNING: Output antiflood triggered, dropping message for %s: %s", target, message)
		return
	}

	proxy.Connection.Privmsg(target, message)
}

func (proxy *ircConnectionProxy) Privmsgf(target, format string, a ...interface{}) {
	proxy.Privmsg(target, fmt.Sprintf(format, a...))
}

func (proxy *ircConnectionProxy) Notice(target, message string) {
	if shouldNotSend, err := proxy.m.TrackOutput(target, message); err != nil {
		log.Printf("WARNING: Output antiflood returned an error, dropping message for %s: %s", target, err.Error())
		return
	} else if shouldNotSend {
		log.Printf("WARNING: Output antiflood triggered, dropping message for %s: %s", target, message)
		return
	}

	proxy.Connection.Notice(target, message)
}

func (proxy *ircConnectionProxy) Noticef(target, format string, a ...interface{}) {
	proxy.Notice(target, fmt.Sprintf(format, a...))
}
