package wikipedia

//go:generate go run ../../util/apigen/main.go --pkg wikipedia v1.yml

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/icedream/irc-medialink/parsers"
)

// Parser implements parsing of Wikipedia URLs.
type Parser struct{}

// Name returns the parser's descriptive name.
func (p *Parser) Name() string {
	return "Wikipedia"
}

// Init initializes the parser.
func (p *Parser) Init(_ context.Context) error {
	return nil
}

// Parse parses the given URL.
func (p *Parser) Parse(ctx context.Context, u *url.URL, referer *url.URL) (result parsers.ParseResult) {
	if !(strings.EqualFold(u.Scheme, "http") ||
		strings.EqualFold(u.Scheme, "https")) ||
		(!strings.HasSuffix(strings.ToLower(u.Host), ".wikipedia.org") &&
			!strings.EqualFold(u.Host, "wikipedia.org")) {
		result.Ignored = true
		return
	}

	switch {
	case strings.HasPrefix(u.Path, "/wiki/"):
		// Wiki entry link
		titleEscaped := u.Path[6:]
		if len(titleEscaped) <= 0 {
			break
		}

		// We're using the original host for link localization
		// or en.wikipedia.org for (www.)wikipedia.org
		if strings.EqualFold(u.Host, "wikipedia.org") ||
			strings.EqualFold(u.Host, "www.wikipedia.org") {
			u.Host = "en.wikipedia.org"
		}
		req, err := http.NewRequestWithContext(ctx, "GET", "https://"+u.Host+"/api/rest_v1/page/summary/"+titleEscaped, nil)
		if err != nil {
			result.Error = err
			return
		}
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			result.Error = err
			return
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			result.UserError = errors.New(r.Status)
			return
		}
		data := v1Summary{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			result.Error = err
			return
		}

		result.Information = []map[string]interface{}{
			{
				"Header":      "\x031,0Wikipedia\x03",
				"Description": prepareSummary(data.Title, data.Extract),
			},
		}
	}

	result.Ignored = result.Information == nil
	return
}
