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

	rxNewlines = regexp.MustCompile(`(?:\r?\n)*`)
)

const (
	noTitleStr  = "(no title)"
	maxHtmlSize = 8 * 1024
)

type Parser struct{}

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

	// Make request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		result.Error = err
		return
	}
	if referer != nil {
		req.Header.Set("Referer", referer.String())
	}
	if resp, err := http.DefaultTransport.RoundTrip(req); err != nil {
		log.Print("HTTP Get failed")
		result.Error = err
		return
	} else {
		log.Printf("Web parser result: %+v", resp)
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
		log.Print(contentType[0:sep])
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
			log.Print("Parsing image...")

			// No need to limit the reader to a specific size here as
			// image.DecodeConfig only reads as much as needed anyways.
			if m, imgType, err := image.DecodeConfig(resp.Body); err != nil {
				result.UserError = ErrCorruptedImage
			} else {
				info := map[string]interface{}{
					"IsUpload":  true,
					"ImageSize": image.Point{X: m.Width, Y: m.Height},
					"ImageType": strings.ToUpper(imgType),
				}
				if resp.ContentLength > 0 {
					info["Size"] = uint64(resp.ContentLength)
				}
				result.Information = []map[string]interface{}{info}
				log.Printf("Got through: %+v!", info)
			}
		default:
			// TODO - Implement generic head info?
			log.Printf("web parser: Ignoring content of type %s", resp.Header.Get("content-type"))
			result.Ignored = true
		}
	}

	return
}
