package manager

import (
	"context"
	"errors"
	"log"
	"net/url"
	"reflect"

	"github.com/icedream/irc-medialink/parsers"
)

// ErrAlreadyLoaded is returned when a parser attempting to register is already found to be loaded with the same ID.
var ErrAlreadyLoaded = errors.New("already loaded")

// Parser describes the core functionality of any parser used to analyze URLs.
type Parser interface {
	Init(ctx context.Context) error
	Name() string
	Parse(ctx context.Context, u *url.URL, referer *url.URL) parsers.ParseResult
}

// GetParsers returns a slice of currently loaded parsers.
func (m *Manager) GetParsers() []Parser {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()

	result := make([]Parser, len(m.registeredParsers))
	copy(result, m.registeredParsers)
	return result
}

// RegisterParser is called by a parser package to register itself automatically.
func (m *Manager) RegisterParser(ctx context.Context, parser Parser) error {
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
	if err := parser.Init(ctx); err != nil {
		return err
	}

	m.registeredParsers = append(m.registeredParsers, parser)
	log.Printf("Registered %s parser!", parser.Name())

	return nil
}

// Parse goes through all loaded parsers in order to analyze a given URL.
func (m *Manager) Parse(ctx context.Context, currentURL *url.URL) (string, parsers.ParseResult) {
	var referer *url.URL
	attempt := 0
followLoop:
	for currentURL != nil {
		attempt++
		if attempt > 15 {
			log.Printf("WARNING: Potential infinite loop for url %s, abort parsing", currentURL)
			break
		}
		for _, p := range m.GetParsers() {
			var refererCopy *url.URL
			if referer != nil {
				refererCopy = &url.URL{}
				*refererCopy = *referer
			}
			currentURLCopy := &url.URL{}
			*currentURLCopy = *currentURL
			// TODO - implement individual timeouts
			r := p.Parse(ctx, currentURLCopy, refererCopy)
			if r.Ignored {
				continue
			}
			if r.FollowURL != nil {
				if *currentURL == *r.FollowURL {
					log.Printf("WARNING: Ignoring request to follow to same URL, ignoring.")
					break followLoop
				}
				referer = currentURL
				currentURL = r.FollowURL
				log.Printf("Redirect %s => %s", referer.String(), currentURL.String())
				continue followLoop
			}
			log.Printf("Parser match %s - %s %+v", currentURL.String(), p.Name(), r)
			return p.Name(), r
		}
		currentURL = nil
	}

	// No parser matches, link ignored
	log.Printf("No parser match %s", currentURL)
	return "", parsers.ParseResult{
		Ignored: true,
	}
}
