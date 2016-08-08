package web

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/icedream/irc-medialink/parsers"
	"github.com/icedream/irc-medialink/util/limitedio"
	"github.com/yhat/scrape"
)

var (
	ErrCorruptedImage = errors.New("Corrupted image.")

	rxNewlines = regexp.MustCompile(`(?:\r?\n)+`)
)

const (
	runeHash    = '#'
	noTitleStr  = "(no title)"
	maxHtmlSize = 8 * 1024
)

type Parser struct {
	EnableImages bool
}

func (p *Parser) Init() error {
	return nil
}

func (p *Parser) Name() string {
	return "Web"
}

func (p *Parser) Parse(u *url.URL, referer *url.URL) (result parsers.ParseResult) {
	// Ignore non-HTTP link
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		result.Ignored = true
		return
	}

	// Remove hash reference from URL since that's not meant to be in the request
	if strings.Contains(u.Path, string(runeHash)) {
		u = &(*u) // avoid modifying original URL object
		u.Path = u.Path[0:strings.IndexRune(u.Path, runeHash)]
	}

	// Make request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		result.Error = err
		return
	}
	if referer != nil {
		req.Header.Set("Referer", referer.String())
	}
	req.Header.Set("User-Agent", "MediaLink IRC Bot")
	if resp, err := http.DefaultTransport.RoundTrip(req); err != nil {
		result.Error = err
		return
	} else {
		defer resp.Body.Close()
		if 300 <= resp.StatusCode && resp.StatusCode < 400 {
			if u2, err := resp.Location(); err == nil && u2 != nil && *u2 != *u {
				result.FollowUrl = u2
				return
			}
		}
		if resp.StatusCode >= 400 {
			result.UserError = errors.New(resp.Status)
			return
		}
		if resp.StatusCode != 200 {
			result.Ignored = true
			return
		}
		contentType := resp.Header.Get("content-type")
		sep := strings.IndexRune(contentType, ';')
		if sep < 0 {
			sep = len(contentType)
		}
		switch strings.ToLower(contentType[0:sep]) {
		case "text/html":
			// Parse the page
			var contentLength int
			if resp.ContentLength < 0 || resp.ContentLength > maxHtmlSize {
				contentLength = maxHtmlSize
			} else {
				contentLength = int(resp.ContentLength)
			}
			limitedBody := limitedio.NewLimitedReader(resp.Body, contentLength)
			root, err := html.Parse(limitedBody)
			if err != nil {
				result.Error = err
				return
			}
			// Search for the title
			result.Information = []map[string]interface{}{
				map[string]interface{}{
					"IsUpload": false,
				},
			}
			title, ok := scrape.Find(root, scrape.ByTag(atom.Title))
			if ok {
				// Got it!
				result.Information[0]["Title"] = rxNewlines.ReplaceAllString(scrape.Text(title), " ")
			} else {
				// No title found
				result.Information[0]["Title"] = noTitleStr
			}
		case "image/png", "image/jpeg", "image/gif":
			if p.EnableImages {

				// No need to limit the reader to a specific size here as
				// image.DecodeConfig only reads as much as needed anyways.
				if m, imgType, err := image.DecodeConfig(resp.Body); err != nil {
					result.UserError = ErrCorruptedImage
				} else {
					info := map[string]interface{}{
						"IsUpload":  true,
						"ImageSize": image.Point{X: m.Width, Y: m.Height},
						"ImageType": strings.ToUpper(imgType),
						"Title":     u.Path[strings.LastIndex(u.Path, "/")+1:],
					}
					if resp.ContentLength > 0 {
						info["Size"] = uint64(resp.ContentLength)
					}
					result.Information = []map[string]interface{}{info}
				}
				break
			}

			fallthrough
		default:
			// TODO - Implement generic head info?
			result.Ignored = true
		}
	}

	return
}
