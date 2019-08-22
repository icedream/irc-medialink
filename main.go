package main

import (
	"log"
	"strings"
	"time"

	"net/url"

	irc "github.com/thoj/go-ircevent"
	"gopkg.in/alecthomas/kingpin.v2"
	"mvdan.cc/xurls"

	"github.com/icedream/irc-medialink/manager"
	"github.com/icedream/irc-medialink/parsers/soundcloud"
	"github.com/icedream/irc-medialink/parsers/web"
	"github.com/icedream/irc-medialink/parsers/wikipedia"
	"github.com/icedream/irc-medialink/parsers/youtube"
)

func must(err error) {
	if err == nil {
		return
	}

	log.Fatal(err)
}

func main() {
	var youtubeApiKey string

	var soundcloudClientId string
	var soundcloudClientSecret string

	var webEnableImages bool

	var debug bool
	var noInvite bool
	var useTLS bool
	var server string
	var password string
	var timeout time.Duration
	var pingFreq time.Duration

	nickname := "YouTubeBot"
	ident := "youtube"
	var nickservPw string
	channels := []string{}

	// IRC config
	kingpin.Flag("nick", "The nickname.").Short('n').StringVar(&nickname)
	kingpin.Flag("ident", "The ident.").Short('i').StringVar(&ident)
	kingpin.Flag("debug", "Enables debug mode.").Short('d').BoolVar(&debug)
	kingpin.Flag("no-invite", "Disables auto-join on invite.").BoolVar(&noInvite)
	kingpin.Flag("tls", "Use TLS.").BoolVar(&useTLS)
	kingpin.Flag("server", "The server to connect to.").Short('s').StringVar(&server)
	kingpin.Flag("password", "The password to use for logging into the IRC server.").Short('p').StringVar(&password)
	kingpin.Flag("timeout", "The timeout on the connection.").Short('t').DurationVar(&timeout)
	kingpin.Flag("pingfreq", "The ping frequency.").DurationVar(&pingFreq)
	kingpin.Flag("nickserv-pw", "NickServ password.").StringVar(&nickservPw)
	kingpin.Flag("channels", "Channels to join.").Short('c').StringsVar(&channels)

	// Youtube config
	kingpin.Flag("youtube-key", "The API key to use to access the YouTube API.").StringVar(&youtubeApiKey)

	// SoundCloud config
	kingpin.Flag("soundcloud-id", "The SoundCloud ID.").StringVar(&soundcloudClientId)
	kingpin.Flag("soundcloud-secret", "The SoundCloud secret.").StringVar(&soundcloudClientSecret)

	// Web parser config
	kingpin.Flag("images", "Enables parsing links of images. Disabled by default for legal reasons.").BoolVar(&webEnableImages)

	kingpin.Parse()

	if len(nickname) == 0 {
		log.Fatal("Nickname must be longer than 0 chars.")
	}
	if len(ident) == 0 {
		log.Fatal("Ident must be longer than 0 chars.")
	}

	// Manager
	m := manager.NewManager()

	// Load youtube parser
	if len(youtubeApiKey) > 0 {
		youtubeParser := &youtube.Parser{
			Config: &youtube.Config{ApiKey: youtubeApiKey},
		}
		must(m.RegisterParser(youtubeParser))
	} else {
		log.Println("No YouTube API key provided, YouTube parsing via API is disabled.")
	}

	// Load soundcloud parser
	if len(soundcloudClientId) > 0 && len(soundcloudClientSecret) > 0 {
		soundcloudParser := &soundcloud.Parser{
			Config: &soundcloud.Config{
				ClientId:     soundcloudClientId,
				ClientSecret: soundcloudClientSecret,
			},
		}
		must(m.RegisterParser(soundcloudParser))
	} else {
		log.Println("No SoundCloud client ID or secret provided, SoundCloud parsing via API is disabled.")
	}

	// Load wikipedia parser
	must(m.RegisterParser(new(wikipedia.Parser)))

	// Load web parser
	webParser := &web.Parser{
		EnableImages: webEnableImages,
	}
	must(m.RegisterParser(webParser))

	// IRC
	conn := m.AntifloodIrcConn(irc.IRC(nickname, ident))
	conn.Debug = debug
	conn.VerboseCallbackHandler = conn.Debug
	conn.UseTLS = useTLS
	conn.Password = password
	if timeout > time.Duration(0) {
		conn.Timeout = timeout
	}
	if pingFreq > time.Duration(0) {
		conn.PingFreq = pingFreq
	}

	joinChan := make(chan string)
	inviteChan := make(chan string)

	// register callbacks
	conn.AddCallback("001", func(e *irc.Event) { // handle RPL_WELCOME
		// nickserv login
		if len(nickservPw) > 0 {
			conn.Privmsg("NickServ", "IDENTIFY "+nickservPw)
			log.Print("Sent NickServ login request.")
		}

		// I am a bot! (+B user mode)
		conn.Mode(conn.GetNick(), "+B-iw")

		// Join configured channels
		if len(channels) > 0 {
			conn.Join(strings.Join(channels, ","))
		}
	})
	conn.AddCallback("JOIN", func(e *irc.Event) {
		// Is this JOIN not about us?
		if !strings.EqualFold(e.Nick, conn.GetNick()) {
			// Save this user's details for a temporary ignore
			m.NotifyUserJoined(e.Arguments[0], e.Source)
			return
		}

		// Request channel modes
		resetChannelModes(e.Arguments[0])
		conn.Mode(e.Arguments[0])

		// Asynchronous notification
		select {
		case joinChan <- e.Arguments[0]:
		default:
		}
	})
	conn.AddCallback("PART", func(e *irc.Event) {
		// Is this PART not about us?
		if !strings.EqualFold(e.Nick, conn.GetNick()) {
			return
		}

		deleteChannelModes(e.Arguments[0])
	})
	conn.AddCallback("MODE", func(e *irc.Event) {
		// Is this MODE for a channel?
		isChannel := strings.HasPrefix(e.Arguments[0], "#")

		if !isChannel {
			return
		}

		add := true
		for _, mode := range e.Arguments[1] {
			switch mode {
			case '+':
				add = true
			case '-':
				add = false
			default:
				if add {
					setChannelMode(e.Arguments[0], mode)
				} else {
					unsetChannelMode(e.Arguments[0], mode)
				}
			}
		}
	})
	if !noInvite {
		conn.AddCallback("INVITE", func(e *irc.Event) {
			// Is this INVITE not for us?
			if !strings.EqualFold(e.Arguments[0], conn.GetNick()) {
				return
			}

			// Asynchronous notification
			select {
			case inviteChan <- e.Arguments[1]:
			default:
			}

			// We have been invited, autojoin!
			go func(sourceNick string, targetChannel string) {
			joinWaitLoop:
				for {
					select {
					case channel := <-joinChan:
						if strings.EqualFold(channel, targetChannel) {
							// TODO - Thanks message
							time.Sleep(1 * time.Second)
							conn.Privmsgf(targetChannel, "Thanks for inviting me, %s! I am %s, the friendly bot that shows information about links posted in this channel. I hope I can be of great help for everyone here in %s! :)", sourceNick, conn.GetNick(), targetChannel)
							time.Sleep(2 * time.Second)
							conn.Privmsg(targetChannel, "If you ever run into trouble with me (or find any bugs), please use the channel #MediaLink for contact on this IRC.")
							break joinWaitLoop
						}
					case channel := <-inviteChan:
						if strings.EqualFold(channel, targetChannel) {
							break joinWaitLoop
						}
					case <-time.After(time.Minute):
						log.Printf("WARNING: Timed out waiting for us to join %s as we got invited", targetChannel)
						break joinWaitLoop
					}
				}
			}(e.Nick, e.Arguments[1])
			conn.Join(e.Arguments[1])
		})
	}
	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		go func(event *irc.Event) {
			//sender := event.Nick
			target := event.Arguments[0]
			isChannel := true
			if strings.EqualFold(target, conn.GetNick()) {
				// Private message to us!
				target = event.Nick
				isChannel = false
			}
			if strings.EqualFold(target, conn.GetNick()) {
				// Emergency switch to avoid endless loop,
				// dropping all messages from the bot to the bot!
				log.Printf("BUG - Emergency switch, caught message from bot to bot: %s", event.Arguments)
				return
			}
			msg := stripIrcFormatting(event.Message())

			log.Printf("<%s @ %s> %s", event.Nick, target, msg)

			// Ignore user if they just joined
			if shouldIgnore := m.TrackUser(target, event.Source); shouldIgnore {
				log.Print("This message will be ignored since the user just joined.")
				return
			}

			urlStr := xurls.Relaxed.FindString(msg)

			switch {
			case !isChannel:
				// Explain who we are and what we do
				conn.Privmsgf(target, "Hi, I parse links people post to chat rooms to give some information about them. I also allow people to search for YouTube videos and SoundCloud sounds straight from IRC. If you have questions or got any bug reports, please direct them to Icedream in #MediaLink, thank you!")
			case len(urlStr) > 0: // URL?
				// Parse URL!
				u, err := url.ParseRequestURI(urlStr)
				if err != nil {
					u, err = url.ParseRequestURI("http://" + urlStr)
				}
				if err != nil {
					log.Print(err)
					break
				}

				// Check if this URL has been recently parsed before (antiflood)
				shouldIgnore := m.TrackUrl(target, u)
				if shouldIgnore {
					log.Printf("WARNING: URL antiflood triggered, dropping URL for %s: %s", target, u)
					break
				}

				_, result := m.Parse(u)
				if result.Error != nil {
					log.Print(result.Error)
				}
				if result.UserError != nil {
					if s, err := tplString("error", result.UserError); err != nil {
						log.Print(err)
					} else {
						s = stripIrcFormattingIfChannelBlocksColors(target, s)
						conn.Privmsg(target, s)
					}
				}
				if result.Error == nil && result.UserError == nil && result.Information != nil {
					for _, i := range result.Information {
						if s, err := tplString("link-info", i); err != nil {
							log.Print(err)
						} else {
							s = stripIrcFormattingIfChannelBlocksColors(target, s)
							conn.Privmsg(target, s)
						}
					}
				}
			}

		}(e)
	})

	// connect
	must(conn.Connect(server))

	// listen for errors
	log.Print("Now looping.")
	conn.Loop()
}
