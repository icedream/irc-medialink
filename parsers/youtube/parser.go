package youtube

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	iso8601duration "github.com/ChannelMeter/iso8601duration"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"

	"github.com/icedream/irc-medialink/parsers"
)

const (
	nonYouTubeReference youtubeReference = iota
	videoReference
	channelNameReference
	channelIDReference
	playlistReference

	header = "\x0301,00You\x0300,04Tube"
)

// ErrNotFound is returned when a YouTube URL does not point at anything the API can find.
var ErrNotFound = errors.New("not found")

type youtubeReference uint8

// Parser implements parsing of YouTube URLs via API.
type Parser struct {
	Config  *Config
	Service *youtube.Service
}

func parseYouTubeURL(uri *url.URL, followRedirects int) (youtubeReference, string) {
	u := &(*uri)
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Must be an HTTP URL
	if u.Scheme != "http" && u.Scheme != "https" {
		return nonYouTubeReference, ""
	}

	switch strings.ToLower(u.Host) {
	case "youtu.be":
		// http://youtu.be/{id}
		if s, err := url.QueryUnescape(strings.TrimLeft(u.Path, "/")); err == nil {
			return videoReference, s
		}
	case "youtube.com", "www.youtube.com":
		if u.Path == "/watch" {
			// http://youtube.com/watch?v={id}
			return videoReference, u.Query().Get("v")
		} else if strings.HasPrefix(u.Path, "/live/") {
			// http://youtube.com/live/{id}
			return videoReference, strings.Trim(u.Path[6:], "/")
		} else if strings.HasPrefix(u.Path, "/channel/") && !strings.HasSuffix(u.Path, "/live") {
			// https://www.youtube.com/channel/{channelid}
			return channelIDReference, strings.Trim(u.Path[9:], "/")
		} else if strings.HasPrefix(u.Path, "/c/") && !strings.HasSuffix(u.Path, "/live") {
			// http://youtube.com/c/{channelname}
			return channelNameReference, strings.Trim(u.Path[3:], "/")
		} else if strings.HasPrefix(u.Path, "/user/") && !strings.HasSuffix(u.Path, "/live") {
			// http://youtube.com/user/{channelname}
			return channelNameReference, strings.Trim(u.Path[6:], "/")
		} else if strings.HasPrefix(u.Path, "/playlist") {
			// https://www.youtube.com/playlist?list=PLq34c5GJGiJJlrG9-ByMbuQkTvaFtIflO
			return playlistReference, u.Query().Get("list")
		} else if followRedirects > 0 && len(u.Path) > 1 && !strings.Contains(u.Path[1:], "/") {
			// Maybe https://youtube.com/{channelname}.
			// Does this actually redirect to a channel?
			req, err := http.NewRequest("HEAD", u.String(), nil)
			if err != nil {
				log.Printf("Failed to create HEAD request from %s: %s", u, err)
				return nonYouTubeReference, ""
			}
			resp, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				log.Printf("Failed to check for channel from %s: %s", u, err)
				return nonYouTubeReference, ""
			}
			if resp.StatusCode >= 300 && resp.StatusCode < 400 {
				if lu, err := resp.Location(); err == nil && lu != nil {
					return parseYouTubeURL(lu, followRedirects-1)
				}
			}
		}
	}

	return nonYouTubeReference, ""
}

// Init initializes the parser.
func (p *Parser) Init() error {
	// youtube api
	client := &http.Client{
		Transport: &transport.APIKey{Key: p.Config.APIKey},
	}
	srv, err := youtube.New(client)
	if err != nil {
		return err
	}
	p.Service = srv
	return nil
}

// Name returns the parser's descriptive name.
func (p *Parser) Name() string {
	return "YouTube"
}

// Parse parses the given URL.
func (p *Parser) Parse(u *url.URL, referer *url.URL) (result parsers.ParseResult) {
	// Parse YouTube URL
	idType, id := parseYouTubeURL(u, 2)
	if idType == nonYouTubeReference {
		result.Ignored = true
		return // nothing relevant found in this URL
	}

	switch idType {
	case videoReference:
		// Get YouTube video info
		list, err := p.Service.Videos.List([]string{
			"contentDetails",
			"id",
			"liveStreamingDetails",
			"snippet",
			"statistics",
		}).Id(id).Do()
		if err != nil {
			result.Error = err
			return
		}

		// Any info available?
		if len(list.Items) < 1 {
			result.UserError = ErrNotFound
			return
		}

		// Collect information
		result.Information = []map[string]interface{}{}
		for _, item := range list.Items {
			r := map[string]interface{}{
				"ShortUrl": fmt.Sprintf("https://youtu.be/%v", url.QueryEscape(item.Id)),
				"IsUpload": true,
			}
			if item.Snippet != nil {
				r["Title"] = item.Snippet.Title
				r["Author"] = item.Snippet.ChannelTitle
				r["Description"] = item.Snippet.Description
				r["Category"] = item.Snippet.CategoryId
				r["Tags"] = item.Snippet.Tags

				// parse publishedAt
				if t, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt); err == nil {
					r["PublishedAt"] = t
				} else {
					log.Print(err)
				}
			}
			if item.ContentDetails != nil {
				// parse duration
				if d, err := iso8601duration.FromString(item.ContentDetails.Duration); err == nil {
					r["Duration"] = d.ToDuration()
				} else {
					log.Print(err)
				}

				if item.ContentDetails.ContentRating != nil {
					if item.ContentDetails.ContentRating.YtRating == "ytAgeRestricted" {
						r["AgeRestriction"] = "NSFW"
					}
				}
			}
			if item.Statistics != nil {
				r["Views"] = item.Statistics.ViewCount
				r["Comments"] = item.Statistics.CommentCount
				r["Likes"] = item.Statistics.LikeCount
				r["Dislikes"] = item.Statistics.DislikeCount
				r["Favorites"] = item.Statistics.FavoriteCount
			}
			if item.LiveStreamingDetails != nil {
				hasStartTime := len(item.LiveStreamingDetails.ActualStartTime) > 0
				hasScheduledStartTime := len(item.LiveStreamingDetails.ScheduledStartTime) > 0
				hasEndTime := len(item.LiveStreamingDetails.ActualEndTime) > 0
				hasScheduledEndTime := len(item.LiveStreamingDetails.ScheduledEndTime) > 0
				isLive := hasStartTime && !hasEndTime
				var startTime, endTime, scheduledStartTime, scheduledEndTime time.Time
				if hasStartTime {
					parsed, err := time.Parse(time.RFC3339, item.LiveStreamingDetails.ActualStartTime)
					if err == nil {
						startTime = parsed
					}
				}
				if hasEndTime {
					parsed, err := time.Parse(time.RFC3339, item.LiveStreamingDetails.ActualEndTime)
					if err == nil {
						endTime = parsed
					}
				}
				if hasScheduledStartTime {
					parsed, err := time.Parse(time.RFC3339, item.LiveStreamingDetails.ScheduledStartTime)
					if err == nil {
						scheduledStartTime = parsed
					}
				}
				if hasScheduledEndTime {
					parsed, err := time.Parse(time.RFC3339, item.LiveStreamingDetails.ScheduledEndTime)
					if err == nil {
						scheduledEndTime = parsed
					}
				}
				r["IsLive"] = isLive
				r["IsUpcomingLive"] = !hasStartTime
				r["IsFinishedLive"] = hasStartTime && hasEndTime
				r["ScheduledStartTime"] = scheduledStartTime
				r["ScheduledEndTime"] = scheduledEndTime
				r["ActualStartTime"] = startTime
				r["ActualEndTime"] = endTime
				r["Viewers"] = item.LiveStreamingDetails.ConcurrentViewers
			}
			r["Header"] = header
			result.Information = append(result.Information, r)
		}
	case channelIDReference, channelNameReference:
		// Get YouTube channel info
		cl := p.Service.Channels.List([]string{
			"id",
			"snippet",
			"statistics",
		})
		if idType == channelNameReference {
			cl = cl.ForUsername(id)
		} else {
			cl = cl.Id(id)
		}
		list, err := cl.Do()
		if err != nil {
			result.Error = err
			return
		}

		// Any info available?
		if len(list.Items) < 1 {
			result.UserError = ErrNotFound
			return
		}

		// Collect information
		result.Information = []map[string]interface{}{}
		for _, item := range list.Items {
			r := map[string]interface{}{
				"Header":      header,
				"IsProfile":   true,
				"Name":        item.Snippet.Title,
				"CountryCode": item.Snippet.Country,
				"Description": item.Snippet.Description,
				"ShortUrl":    item.Snippet.CustomUrl,
				"Comments":    item.Statistics.CommentCount,
				"Videos":      item.Statistics.VideoCount,
				"Views":       item.Statistics.ViewCount,
			}
			if !item.Statistics.HiddenSubscriberCount {
				r["Followers"] = item.Statistics.SubscriberCount
			}
			result.Information = append(result.Information, r)
		}
	case playlistReference:
		// Get YouTube channel info
		list, err := p.Service.Playlists.List([]string{
			"id",
			"snippet",
		}).Id(id).Do()
		if err != nil {
			result.Error = err
			return
		}

		// Any info available?
		if len(list.Items) < 1 {
			result.UserError = ErrNotFound
			return
		}

		// Collect information
		result.Information = []map[string]interface{}{}
		for _, item := range list.Items {
			r := map[string]interface{}{
				"Header":      header,
				"IsPlaylist":  true,
				"Title":       "Playlist: " + item.Snippet.Title,
				"Author":      item.Snippet.ChannelTitle,
				"PublishedAt": item.Snippet.PublishedAt,
				"Description": item.Snippet.Description,
			}

			// parse publishedAt
			if t, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt); err == nil {
				r["PublishedAt"] = t
			} else {
				log.Print(err)
			}

			result.Information = append(result.Information, r)
		}
	default:
		result.Ignored = true
	}

	return
}
