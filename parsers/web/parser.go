package web

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"  // GIF support for image
	_ "image/jpeg" // JPEG support for image
	_ "image/png"  // PNG support for image
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/icedream/irc-medialink/parsers"
	"github.com/icedream/irc-medialink/util/limitedio"
	"github.com/icedream/irc-medialink/version"
)

var (
	// ErrCorruptedImage is returned when an image type has been detected but the contents are unreadable.
	ErrCorruptedImage = errors.New("corrupted image")

	rxNewlines = regexp.MustCompile(`(?:\r?\n)+`)
)

const (
	noTitleStr  = "(no title)"
	maxHTMLSize = 32 * 1024
)

// Parser implements parsing of standard HTML web pages.
type Parser struct {
	UserAgent string
	Config    Config
}

// Init initializes this parser.
func (p *Parser) Init() error {
	if len(version.AppVersion) > 0 {
		p.UserAgent = fmt.Sprintf("%s/%s", version.AppName, strings.TrimLeft(version.AppVersion, "v"))
	} else {
		p.UserAgent = version.AppName
	}

	// Turn out HTTP user agent into something that is more "common" and less likely to be blacklisted
	p.UserAgent = fmt.Sprintf("Mozilla/5.0 (compatible; actually an IRC bot) %s", p.UserAgent)
	return nil
}

// Name returns the descriptive name of this parser.
func (p *Parser) Name() string {
	return "Web"
}

func (p *Parser) enrichImageInfo(body io.Reader, result *parsers.ParseResult) {
	if !p.Config.EnableImages {
		return
	}

	// No need to limit the reader to a specific size here as
	// image.DecodeConfig only reads as much as needed anyways.
	if m, imgType, err := image.DecodeConfig(body); err != nil {
		result.UserError = ErrCorruptedImage
	} else {
		info := map[string]interface{}{
			"ImageSize": image.Point{X: m.Width, Y: m.Height},
			"ImageType": strings.ToUpper(imgType),
		}
		result.Information = []map[string]interface{}{info}
	}
}

func mimeTypeToName(imgType string) string {
	switch strings.ToUpper(imgType) {
	case "image/png":
		return "PNG"
	case "image/jpeg":
		return "JPEG"
	case "image/gif":
		return "GIF"
	case "image/tiff", "image/tif", "image/x-tif":
		return "TIFF"
	case "image/bmp":
		return "BMP"
	default:
		return imgType
	}
}

// Parse analyzes a given URL.
func (p *Parser) Parse(u *url.URL, referer *url.URL) (result parsers.ParseResult) {
	// Ignore non-HTTP link
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		result.Ignored = true
		return
	}

	// Remove fragment from URL since that's not meant to be in the request
	if len(u.Fragment) > 0 {
		u = &(*u)
		u.Fragment = ""
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
	req.Header.Set("User-Agent", p.UserAgent)
	if len(p.Config.AcceptLanguage) > 0 {
		req.Header.Set("Accept-Language", p.Config.AcceptLanguage)
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		result.Error = err
		return
	}
	defer resp.Body.Close()
	if 300 <= resp.StatusCode && resp.StatusCode < 400 {
		if u2, err := resp.Location(); err == nil && u2 != nil && *u2 != *u {
			result.FollowURL = u2
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
		if resp.ContentLength < 0 || resp.ContentLength > maxHTMLSize {
			contentLength = maxHTMLSize
		} else {
			contentLength = int(resp.ContentLength)
		}
		limitedBody := limitedio.NewLimitedReader(resp.Body, contentLength)

		root, err := html.Parse(limitedBody)
		if err != nil {
			result.Error = err
			return
		}

		// Search for opengraph data first
		og := opengraph.NewOpenGraph()
		for _, meta := range scrape.FindAll(root, scrape.ByTag(atom.Meta)) {
			if meta.Attr != nil {
				m := make(map[string]string)
				for _, a := range meta.Attr {
					m[atom.String([]byte(a.Key))] = a.Val
				}
				og.ProcessMeta(m)
			}
		}

		result.Information = []map[string]interface{}{
			{
				"Description": og.Description,
				"Title":       og.Title,
				"Header":      og.SiteName,
				"Determiner":  og.Determiner,
			},
		}
		mergeFirstMedia := false
		switch og.Type {
		case "article":
			result.Information[0]["IsUpload"] = true
			mergeFirstMedia = true
		case "profile":
			mergeFirstMedia = true
		case "music.song":
			result.Information[0]["IsUpload"] = true
			result.Information[0]["IsSong"] = true
			mergeFirstMedia = true
		case "music.musician":
			result.Information[0]["IsProfile"] = true
			result.Information[0]["IsMusician"] = true
			result.Information[0]["IsArtist"] = true
		case "video.other":
			result.Information[0]["IsUpload"] = true
			mergeFirstMedia = true
		}
		if m := og.Article; m != nil {
			var info map[string]interface{}
			if mergeFirstMedia {
				info = result.Information[0]
				mergeFirstMedia = false
			} else {
				result.Information = append(result.Information, info)
				info = map[string]interface{}{}
			}
			info["IsArticle"] = true
			info["Author"] = strings.Join(m.Authors, ", ")
			info["Authors"] = m.Authors
			info["Tags"] = m.Tags
			info["Section"] = m.Section
			info["ModifiedTime"] = m.ModifiedTime
			info["ExpirationTime"] = m.ExpirationTime
			info["PublishedTime"] = m.PublishedTime
		}
		if m := og.Book; m != nil {
			var info map[string]interface{}
			if mergeFirstMedia {
				info = result.Information[0]
				mergeFirstMedia = false
			} else {
				info = map[string]interface{}{}
			}
			info["IsBook"] = true
			info["Author"] = strings.Join(m.Authors, ", ")
			info["Authors"] = m.Authors
			info["ISBN"] = m.ISBN
			info["Tags"] = m.Tags
			info["ReleaseDate"] = m.ReleaseDate
			result.Information = append(result.Information, info)
		}
		if m := og.Profile; m != nil {
			var info map[string]interface{}
			if mergeFirstMedia {
				info = result.Information[0]
				mergeFirstMedia = false
			} else {
				info = map[string]interface{}{}
				result.Information = append(result.Information, info)
			}
			info["IsProfile"] = true
			info["Name"] = fmt.Sprintf("%s %s", m.FirstName, m.LastName)
			if len(m.Username) > 0 {
				info["Title"] = m.Username
			}
			info["Gender"] = m.Gender
		}
		for _, m := range og.Videos {
			var info map[string]interface{}
			if mergeFirstMedia {
				info = result.Information[0]
				mergeFirstMedia = false
			} else {
				info = map[string]interface{}{}
				result.Information = append(result.Information, info)
			}
			info["IsUpload"] = true
			info["Tags"] = m.Tags
			if m.Duration != 0 {
				info["Duration"] = time.Second * time.Duration(m.Duration)
			}
		}
		// for _, m := range og.Images {
		// 	var info map[string]interface{}
		// 	if mergeFirstMedia {
		// 		info = result.Information[0]
		// 		mergeFirstMedia = false
		// 	} else {
		// 		info = map[string]interface{}{}
		// 		result.Information = append(result.Information, info)
		// 	}
		// 	info["IsUpload"] = true
		// 	if m.Width != 0 && m.Height != 0 {
		// 		info["ImageSize"] = image.Point{X: int(m.Width), Y: int(m.Height)}
		// 	}
		// 	if len(m.Type) > 0 {
		// 		info["ImageType"] = mimeTypeToName(m.Type)
		// 	}
		// }

		if len(og.Title) == 0 {
			// Search for the title as fallback
			result.Information = []map[string]interface{}{
				{
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
		}
	case "image/png", "image/jpeg", "image/gif":
		if p.Config.EnableImages {
			p.enrichImageInfo(resp.Body, &result)

			if result.UserError == nil {
				result.Information[0]["IsUpload"] = true
				result.Information[0]["Title"] = u.Path[strings.LastIndex(u.Path, "/")+1:]
				if resp.ContentLength > 0 {
					result.Information[0]["Size"] = uint64(resp.ContentLength)
				}
			}
			break
		}

		fallthrough
	default:
		// TODO - Implement generic head info?
		result.Ignored = true
	}

	return
}
