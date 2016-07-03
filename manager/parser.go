package manager

import (
	"errors"
	"log"
	"net/url"
	"reflect"

	"github.com/icedream/irc-medialink/parsers"
)

var (
	ErrAlreadyLoaded = errors.New("Already loaded.")
)

type Parser interface {
	Init() error
	Name() string
	Parse(u *url.URL, referer *url.URL) parsers.ParseResult
}

func (m *Manager) GetParsers() []Parser {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()

	result := make([]Parser, len(m.registeredParsers))
	copy(result, m.registeredParsers)
	return result
}

func (m *Manager) RegisterParser(parser Parser) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	// Make sure that parser hasn't already been loaded in some way or another
	t := reflect.TypeOf(parser)
	for _, p := range m.registeredParsers {
		if reflect.TypeOf(p) == t {
			return ErrAlreadyLoaded
		}
	}

	// Initialize parser
	log.Printf("Initializing %s parser...", parser.Name())
	if err := parser.Init(); err != nil {
		return err
	}

	m.registeredParsers = append(m.registeredParsers, parser)
	log.Printf("Registered %s parser!", parser.Name())

	return nil
}

func (m *Manager) Parse(currentUrl *url.URL) (string, parsers.ParseResult) {
	var referer *url.URL
	attempt := 0
followLoop:
	for currentUrl != nil {
		attempt++
		if attempt > 15 {
			log.Printf("WARNING: Potential infinite loop for url %s, abort parsing", currentUrl)
			break
		}
		for _, p := range m.GetParsers() {
			var refererCopy *url.URL
			if referer != nil {
				refererCopy = &url.URL{}
				*refererCopy = *referer
			}
			currentUrlCopy := &url.URL{}
			*currentUrlCopy = *currentUrl
			r := p.Parse(currentUrlCopy, refererCopy)
			if r.Ignored {
				continue
			}
			if r.FollowUrl != nil {
				if *currentUrl == *r.FollowUrl {
					log.Printf("WARNING: Ignoring request to follow to same URL, ignoring.")
					break followLoop
				}
				referer = currentUrl
				currentUrl = r.FollowUrl
				continue followLoop
			}
			return p.Name(), r
		}
		currentUrl = nil
	}

	// No parser matches, link ignored
	return "", parsers.ParseResult{
		Ignored: true,
	}
}
