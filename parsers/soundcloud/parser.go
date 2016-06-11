package soundcloud

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/icedream/irc-medialink/parsers"
	"github.com/yanatan16/golang-soundcloud/soundcloud"
)

const (
	header = "\x0307SoundCloud\x03"
)

var (
	emptyUrlValues = url.Values{}
)

type Parser struct {
	api    *soundcloud.Api
	http   *http.Client
	Config *Config
}

func (p *Parser) Init() error {
	p.api = &soundcloud.Api{
		ClientId:     p.Config.ClientId,
		ClientSecret: p.Config.ClientSecret,
	}
	p.http = &http.Client{}

	return nil
}

func (p *Parser) Name() string {
	return "SoundCloud"
}

func (p *Parser) Parse(u *url.URL, referer *url.URL) (result parsers.ParseResult) {
	if !strings.EqualFold(u.Host, "soundcloud.com") &&
		!strings.EqualFold(u.Host, "www.soundcloud.com") {
		result.Ignored = true
		return
	}

	if r, err := p.v2resolve(u.String()); err != nil {
		result.UserError = err
		return
	} else {
		log.Printf("SoundCloud parser: Got link of kind %s.", r.Kind)
		switch r.Kind {
		case v2KindUser:
			user := r.AsUser()

			info := map[string]interface{}{
				"Header":      header,
				"IsProfile":   true,
				"Name":        user.Username,
				"City":        user.City,
				"CountryCode": user.CountryCode, // TODO - Convert user.CountryCode to a human-readable country
				"Url":         user.PermalinkURL,
				"Followers":   user.FollowersCount,
				"Uploads":     user.TrackCount,
				"Playlists":   user.PlaylistCount,
				"IsVerified":  user.Verified,
				// TODO - Mark premium account
			}

			if len(user.FullName) > 0 {
				info["Name"] = info["Name"].(string) + " (" + user.FullName + ")"
			}

			result.Information = []map[string]interface{}{info}
		case v2KindGroup:
			group := r.AsGroup()

			info := map[string]interface{}{
				"Header":      header,
				"IsGroup":     true,
				"Title":       fmt.Sprintf("Group: %s", group.Name),
				"Author":      group.Creator.Username,
				"Url":         group.PermalinkURL,
				"PublishedAt": group.CreatedAt.ToTime(""),
			}

			result.Information = []map[string]interface{}{info}
		case v2KindTrack:
			track := r.AsTrack()
			log.Println("Track: %+v", track)

			info := map[string]interface{}{
				"Header":      header,
				"IsUpload":    true,
				"Title":       track.Title,
				"Author":      track.User.Username,
				"Url":         track.PermalinkURL,
				"Favorites":   track.LikesCount,
				"Reposts":     track.RepostsCount,
				"Plays":       track.PlaybackCount,
				"Comments":    track.CommentCount,
				"PublishedAt": track.CreatedAt.ToTime(""),
				"Downloads":   track.DownloadCount,
				// Doing /1000 here to get rid of the fraction
				"Duration": (time.Duration(track.Duration) / 1000) * time.Second,
			}

			result.Information = []map[string]interface{}{info}
		case v2KindPlaylist:
			pl := r.AsPlaylist()

			info := map[string]interface{}{
				"Header":      header,
				"IsPlaylist":  true,
				"Title":       "Playlist: " + pl.Title,
				"Author":      pl.User.Username,
				"Url":         pl.PermalinkURL,
				"PublishedAt": pl.CreatedAt.ToTime(""),
				"Tracks":      pl.TrackCount,
				"Favorites":   pl.LikesCount,
				"Reposts":     pl.RepostsCount,
				// Doing /1000 here to get rid of the fraction
				"Duration": (time.Duration(pl.Duration) / 1000) * time.Second,
			}

			result.Information = []map[string]interface{}{info}
		default:
			result.Ignored = true
		}
	}

	return
}
