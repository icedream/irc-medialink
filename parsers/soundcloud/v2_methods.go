package soundcloud

//go:generate go run ../../util/apigen/main.go --pkg soundcloud v2.yml

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type v2Kind string

const (
	// FIXME - not yet supported
	//v2KindApp      v2Kind = "app"
	v2KindGroup    v2Kind = "group"
	v2KindPlaylist v2Kind = "playlist"
	v2KindTrack    v2Kind = "track"
	v2KindUser     v2Kind = "user"
)

var (
	// ErrUnsupportedKind is returned if the parser can't detect what type of reference a SoundCloud URL is pointing to.
	ErrUnsupportedKind = errors.New("Unsupported kind")
)

func (p *Parser) v2url(path string, urlvalues url.Values) *url.URL {
	u := &url.URL{
		Scheme:   "https",
		Host:     "api.soundcloud.com",
		RawQuery: urlvalues.Encode(),
		Path:     path,
	}
	q := u.Query()
	q.Set("client_id", p.Config.ClientID)
	q.Set("app_version", "0.0") // TODO - Versioning
	u.RawQuery = q.Encode()
	log.Printf("SoundCloud parser: v2url: %s", u.String())
	return u
}

func (p *Parser) v2call(path string, urlvalues url.Values) (v2Result, error) {
	resp, err := p.http.Do(&http.Request{
		URL:    p.v2url(path, urlvalues),
		Method: "GET",
		Header: http.Header{
			"Accept":     []string{"application/json"},
			"User-Agent": []string{"Icedream-YouTubeIRC/0.0"}, // TODO - Versioning
		},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	b := new(bytes.Buffer)
	io.Copy(b, resp.Body)
	log.Printf("v2call: Response is %s", b.String())
	return b.Bytes(), nil
}

func (p *Parser) v2resolve(u string) (*v2ResolveResult, error) {
	r, err := p.v2call("/resolve", url.Values{
		"url": []string{u},
	})
	if err != nil {
		return nil, err
	}

	// First find out the kind via the most generic type possible
	obj := new(v2Object)
	if err := r.Decoder().Decode(obj); err != nil {
		return nil, err
	}

	// Now decode as correct type
	switch obj.Kind {
	case v2KindTrack:
		exactObj := new(v2Track)
		if err := r.Decoder().Decode(exactObj); err != nil {
			return nil, err
		}
		return &v2ResolveResult{exactObj, obj.Kind}, nil
	case v2KindUser:
		exactObj := new(v2User)
		if err := r.Decoder().Decode(exactObj); err != nil {
			return nil, err
		}
		return &v2ResolveResult{exactObj, obj.Kind}, nil
	case v2KindGroup:
		exactObj := new(v2Group)
		if err := r.Decoder().Decode(exactObj); err != nil {
			return nil, err
		}
		return &v2ResolveResult{exactObj, obj.Kind}, nil
	case v2KindPlaylist:
		exactObj := new(v2Playlist)
		if err := r.Decoder().Decode(exactObj); err != nil {
			return nil, err
		}
		return &v2ResolveResult{exactObj, obj.Kind}, nil
	default:
		return nil, ErrUnsupportedKind
	}
}

type v2Result []byte

func (r v2Result) Decoder() *json.Decoder {
	return json.NewDecoder(bytes.NewReader([]byte(r)))
}

type v2ResolveResult struct {
	result interface{}
	Kind   v2Kind
}

func (r *v2ResolveResult) AsUser() *v2User {
	return r.result.(*v2User)
}

// FIXME - not supported yet
/*func (r *v2ResolveResult) AsApp() *v2App {
	return r.result.(*v2App)
}*/

func (r *v2ResolveResult) AsTrack() *v2Track {
	return r.result.(*v2Track)
}

func (r *v2ResolveResult) AsPlaylist() *v2Playlist {
	return r.result.(*v2Playlist)
}

func (r *v2ResolveResult) AsGroup() *v2Group {
	return r.result.(*v2Group)
}

// FIXME - See https://github.com/golang/go/issues/9037 - Revert to using time.Time in the API as soon as this is fixed!

type timeString string

func (t timeString) ToTime(layout string) (parsedTime time.Time) {
	if t == "" {
		return // default time value
	}

	if len(layout) == 0 {
		// TODO - I need to look up whether SoundCloud simply localizes this now because if yes, THAT'S TROUBLE.
		layout = "2006/01/02 15:04:05 -0700"
	}

	parsedTime, err := time.Parse(layout, string(t))
	if err != nil {
		// It's just a temporary fix anyways
		panic(err)
	}

	return
}
