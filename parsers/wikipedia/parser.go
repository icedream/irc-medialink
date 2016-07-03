package wikipedia

//go:generate go run ../../util/apigen/main.go --pkg wikipedia v1.yml

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/icedream/irc-medialink/parsers"
)

type Parser struct{}

func (p *Parser) Name() string {
	return "Wikipedia"
}

func (p *Parser) Init() error {
	return nil
}

func (p *Parser) Parse(u *url.URL, referer *url.URL) (result parsers.ParseResult) {
	if !strings.HasSuffix(strings.ToLower(u.Host), ".wikipedia.org") ||
		strings.EqualFold(u.Host, "wikipedia.org") {
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
		r, err := http.Get("https://" + u.Host + "/api/rest_v1/page/summary/" + titleEscaped)
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
			map[string]interface{}{
				"Header":      "\x031,0Wikipedia\x03",
				"Description": prepareSummary(data.Title, data.Extract),
			},
		}
	}

	result.Ignored = result.Information == nil
	return
}
