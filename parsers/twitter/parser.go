package twitter

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/icedream/irc-medialink/parsers"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	header = "\x0300,02Twitter"
)

var (
	// ErrNotFound is returned when a URL points to something that the Twitter API can not find.
	ErrNotFound = errors.New("not found")
)

// Parser implements parsing of Twitter URLs.
type Parser struct {
	Config *Config
	Client *twitter.Client
}

// Init initializes this parser.
func (p *Parser) Init() error {
	config := &clientcredentials.Config{
		ClientID:     p.Config.ClientID,
		ClientSecret: p.Config.ClientSecret,
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	httpClient := config.Client(oauth2.NoContext)
	p.Client = twitter.NewClient(httpClient)
	return nil
}

// Name returns the parser's descriptive name.
func (p *Parser) Name() string {
	return "Twitter"
}

type twitterReferenceType byte

const (
	nonTwitterReference twitterReferenceType = iota
	tweetReference
	profileReference
)

func parseTwitterURL(uri *url.URL) (twitterReferenceType, string) {
	u := &(*uri)
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Must be an HTTP URL
	if u.Scheme != "http" && u.Scheme != "https" {
		return nonTwitterReference, ""
	}

	// Remove www. prefix from hostname
	if strings.HasPrefix(u.Host, "www.") {
		u.Host = u.Host[4:]
	}

	/*
		Examples of valid links:
		- https://twitter.com/random_carl/status/1273087230526054402
		- https://twitter.com/random_carl/status/1273087230526054402?s=20
		- https://twitter.com/ScottTheWoz/status/1284307659743731718/photo/1
		- https://twitter.com/random_carl
	*/

	switch strings.ToLower(u.Host) {
	case "twitter.com":
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
func (p *Parser) Parse(u *url.URL, referer *url.URL) (result parsers.ParseResult) {
	// Parse Twitter URL
	idType, id := parseTwitterURL(u)
	if idType == nonTwitterReference {
		result.Ignored = true
		return // nothing relevant found in this URL
	}

	switch idType {
	case tweetReference:
		id, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			result.Error = err
			return
		}
		tweet, resp, err := p.Client.Statuses.Show(id, &twitter.StatusShowParams{})
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
			r["Author"] = "@" + tweet.User.Name
			r["AuthorIsVerified"] = tweet.User.Verified
		}
		if tweet.ExtendedTweet != nil {
			r["Description"] = tweet.ExtendedTweet.FullText
		}

		// parse publishedAt
		if t, err := time.Parse(time.RFC3339, tweet.CreatedAt); err == nil {
			r["PublishedAt"] = t
		} else {
			log.Print(err)
		}

		r["Reposts"] = tweet.RetweetCount
		// TODO - maybe process included URLs of tweet?
		// r["Url"] = tweet.Entities.Urls
		r["Comments"] = tweet.ReplyCount
		r["Favorites"] = tweet.FavoriteCount
		r["Header"] = header
		result.Information = append(result.Information, r)

	case profileReference:
		user, resp, err := p.Client.Users.Show(&twitter.UserShowParams{
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
