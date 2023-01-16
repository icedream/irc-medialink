package twitter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/icedream/irc-medialink/parsers"
	"github.com/icedream/irc-medialink/util/clone"
)

const (
	header = "\x0300,02Twitter"
)

// ErrNotFound is returned when a URL points to something that the Twitter API can not find.
var ErrNotFound = errors.New("not found")

// Parser implements parsing of Twitter URLs.
type Parser struct {
	Config *Config
}

// Init initializes this parser.
func (p *Parser) Init(ctx context.Context) error {
	if len(p.Config.ClientID) == 0 {
		return errors.New("a Twitter client ID is required")
	}
	if len(p.Config.ClientSecret) == 0 {
		return errors.New("a Twitter client secret is required")
	}
	return nil
}

// Name returns the parser's descriptive name.
func (p *Parser) Name() string {
	return "Twitter"
}

func (p *Parser) getTwitterClient(ctx context.Context) *twitter.Client {
	config := &clientcredentials.Config{
		ClientID:     p.Config.ClientID,
		ClientSecret: p.Config.ClientSecret,
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	httpClient := config.Client(ctx)
	return twitter.NewClient(httpClient)
}

type twitterReferenceType byte

const (
	nonTwitterReference twitterReferenceType = iota
	tweetReference
	profileReference
)

func parseTwitterURL(uri *url.URL) (twitterReferenceType, string) {
	u := clone.CloneURL(uri)
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Must be an HTTP URL
	if u.Scheme != "http" && u.Scheme != "https" {
		return nonTwitterReference, ""
	}
	/*
		Examples of valid links:
		- https://twitter.com/random_carl/status/1273087230526054402
		- https://twitter.com/random_carl/status/1273087230526054402?s=20
		- https://twitter.com/ScottTheWoz/status/1284307659743731718/photo/1
		- https://twitter.com/random_carl
	*/

	switch strings.ToLower(u.Host) {
	case "www.twitter.com", "twitter.com":
		parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(parts) == 1 { // /:username
			return profileReference, parts[0]
		}
		if len(parts) > 1 && parts[1] == "status" { // /:username/status/:id[/:extra]
			// TODO - handle explicit photo linking
			return tweetReference, parts[2]
		}
	}

	return nonTwitterReference, ""
}

// Parse parses the given URL.
func (p *Parser) Parse(ctx context.Context, u *url.URL, referer *url.URL) (result parsers.ParseResult) {
	// Parse Twitter URL
	idType, id := parseTwitterURL(u)
	if idType == nonTwitterReference {
		result.Ignored = true
		return // nothing relevant found in this URL
	}

	client := p.getTwitterClient(ctx)

	switch idType {
	case tweetReference:
		id, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			result.Error = err
			return
		}
		tweet, resp, err := client.Statuses.Show(id, &twitter.StatusShowParams{})
		if err != nil {
			result.Error = err
			return
		}
		if resp.StatusCode == http.StatusNotFound {
			result.Error = ErrNotFound
			return
		}
		if resp.StatusCode != http.StatusOK {
			result.Error = fmt.Errorf("Server returned: %s", resp.Status)
			return
		}

		// Collect information
		result.Information = []map[string]interface{}{}
		r := map[string]interface{}{
			"IsUpload": true,
		}
		r["Title"] = tweet.Text
		if tweet.User != nil {
			r["Author"] = "@" + tweet.User.ScreenName
			r["AuthorIsVerified"] = tweet.User.Verified
		}
		if tweet.ExtendedTweet != nil {
			r["Description"] = tweet.ExtendedTweet.FullText
		}

		// parse publishedAt
		if t, err := time.Parse(time.RubyDate, tweet.CreatedAt); err == nil {
			r["PublishedAt"] = t
		} else {
			log.Print(err)
		}

		r["Reposts"] = uint64(tweet.RetweetCount)
		// TODO - maybe process included URLs of tweet?
		// r["Url"] = tweet.Entities.Urls
		r["Comments"] = uint64(tweet.ReplyCount)
		r["Favorites"] = uint64(tweet.FavoriteCount)
		r["Header"] = header
		result.Information = append(result.Information, r)

	case profileReference:
		user, resp, err := client.Users.Show(&twitter.UserShowParams{
			ScreenName: id,
		})
		if err != nil {
			result.Error = err
			return
		}
		if resp.StatusCode == http.StatusNotFound {
			result.Error = ErrNotFound
			return
		}
		if resp.StatusCode != http.StatusOK {
			result.Error = fmt.Errorf("Server returned: %s", resp.Status)
			return
		}

		// Collect information
		result.Information = []map[string]interface{}{
			{
				"Header":      header,
				"IsProfile":   true,
				"Title":       user.ScreenName,
				"Name":        user.Name,
				"Description": user.Description,
				"ShortUrl":    user.URL,
				"Favorites":   uint64(user.FavouritesCount),
				"Comments":    uint64(user.StatusesCount),
				"Followers":   uint64(user.FollowersCount),
				"Listings":    uint64(user.ListedCount),
				"Verified":    user.Verified,
				"Location":    user.Location,
			},
		}

	default:
		result.Ignored = true
	}

	return
}
