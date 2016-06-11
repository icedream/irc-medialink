package parsers

import "net/url"

type ParseResult struct {
	Ignored     bool
	Error       error
	UserError   error
	Information []map[string]interface{}
	FollowUrl   *url.URL
}
