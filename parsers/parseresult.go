package parsers

import "net/url"

// ParseResult contains the general structure for parsers to return information about URLs.
type ParseResult struct {
	// Ignored is set to true when a parser can't do anything with the given URL and wants to pass it off to another parser.
	Ignored bool

	// Error is set by the parser when it runs into a technical error.
	Error error

	// UserError is set by the parser when it wants to notify channel users about an error.
	UserError error

	// Information contains the generated URL information.
	Information []map[string]interface{}

	// FollowURL is set by a parser whenever the framework should restart the parsing process with a new URL.
	// This can happen if a URL is actually an alias for another, more processable or direct URL.
	FollowURL *url.URL
}
