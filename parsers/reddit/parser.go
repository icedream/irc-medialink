package reddit

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"runtime"
	"strings"

	"github.com/vartanbeno/go-reddit/v2/reddit"

	"github.com/icedream/irc-medialink/parsers"
	"github.com/icedream/irc-medialink/version"
)

const (
	header = "\x0307Reddit\x03"
)

var emptyURLValues = url.Values{}

// Parser implements parsing for Reddit URLs.
type Parser struct {
	api    *reddit.Client
	Config *Config
}

// Init initializes this parser.
func (p *Parser) Init() error {
	// <platform>:<app ID>:<version string> (by /u/<reddit username>)
	// TODO make username configurable
	userAgent := fmt.Sprintf("%s:%s:%s (by /u/%s)", runtime.GOOS, "tech.icedream.medialink", strings.TrimLeft(version.AppVersion, "v"), p.Config.RedditUsername)
	log.Println("Using user agent:", userAgent)

	var err error
	var creds *reddit.Credentials
	if len(p.Config.ClientID) > 0 {
		creds = &reddit.Credentials{
			ID:     p.Config.ClientID,
			Secret: p.Config.ClientSecret,
		}
	}

	opts := []reddit.Opt{
		reddit.WithUserAgent(userAgent),
		reddit.WithApplicationOnlyOAuth(true),
	}

	if creds != nil {
		log.Println("Using reddit credentials")
		p.api, err = reddit.NewClient(*creds, opts...)
	} else {
		log.Println("Using readonly reddit client")
		p.api, err = reddit.NewReadonlyClient(opts...)
	}
	if err != nil {
		return err
	}

	return nil
}

// Name returns the parser's descriptive name.
func (p *Parser) Name() string {
	return "Reddit"
}

var (
	rxPostRoute = regexp.MustCompile(".+/comments/(?P<id>[a-z0-9]+)(?:/.*|$)")
	rxWikiRoute = regexp.MustCompile("^/r/(?P<name>[^/]+)/wiki/(?:revisions/|edit/)?(?P<id>[a-z0-9]+)/?$")
	// rxCommentRoute   = regexp.MustCompile(".+/comments/(?P<id>[a-z0-9]+)/comment/(?P<cid>[a-z0-9]+)(?:/.*)?")
	rxSubredditRoute = regexp.MustCompile("/r/(?P<name>[^/]+)$")
)

// Parse parses the given URL.
func (p *Parser) Parse(u *url.URL, referer *url.URL) (result parsers.ParseResult) {
	ctx := context.TODO()

	if !strings.EqualFold(u.Host, "reddit.com") &&
		!strings.EqualFold(u.Host, "www.reddit.com") {
		result.Ignored = true
		return
	}

	if m := rxPostRoute.FindStringSubmatch(u.Path); m != nil {
		// path points to a post
		id := m[1]
		pac, _, err := p.api.Post.Get(ctx, id)
		if err != nil {
			result.Error = err
			return
		}
		result.Information = []map[string]interface{}{
			{
				"IsUpload":    true,
				"IsArticle":   true,
				"IsPost":      true,
				"Author":      pac.Post.Author,
				"PublishedAt": pac.Post.Created.Time,
				"Spoiler":     pac.Post.Spoiler,
				"Upvotes":     uint64(pac.Post.Score),
				"Comments":    uint64(pac.Post.NumberOfComments),
				"IsSelfPost":  pac.Post.IsSelfPost,
				"Locked":      pac.Post.Locked,
				"Title":       pac.Post.Title,
				"Description": pac.Post.Body,
				"Header":      header,
			},
		}
		if pac.Post.Edited != nil {
			result.Information[0]["ModifiedAt"] = pac.Post.Edited.Time
		}
		if pac.Post.NSFW {
			result.Information[0]["AgeRestriction"] = "NSFW"
		}
	} else if m := rxSubredditRoute.FindStringSubmatch(u.Path); m != nil {
		// path points to a post
		name := m[1]
		pac, _, err := p.api.Subreddit.Get(ctx, name)
		if err != nil {
			result.Error = err
			return
		}
		result.Information = []map[string]interface{}{
			{
				"IsGroup":     true,
				"Name":        pac.Name,
				"PublishedAt": pac.Created.Time,
				"Subscribers": uint64(pac.Subscribers),
				"Title":       pac.Title,
				"Description": pac.Description,
				"Header":      header,
			},
		}
		if pac.NSFW {
			result.Information[0]["AgeRestriction"] = "NSFW"
		}
	} else if m := rxWikiRoute.FindStringSubmatch(u.Path); m != nil {
		// path points to a post
		subreddit := m[1]
		page := m[2]
		pac, _, err := p.api.Wiki.Page(ctx, subreddit, page)
		if err != nil {
			result.Error = err
			return
		}
		result.Information = []map[string]interface{}{
			{
				"Description": pac.Content,
				"Header":      header,
			},
		}
		if pac.RevisionDate != nil {
			result.Information[0]["ModifiedAt"] = pac.RevisionDate.Time
		}
	} else {
		// all other paths
		result.Ignored = true
	}

	return
}
