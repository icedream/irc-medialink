package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	irc "github.com/thoj/go-ircevent"
	"gopkg.in/alecthomas/kingpin.v2"
	"mvdan.cc/xurls"

	"github.com/icedream/irc-medialink/manager"
	"github.com/icedream/irc-medialink/parsers/soundcloud"
	"github.com/icedream/irc-medialink/parsers/twitter"
	"github.com/icedream/irc-medialink/parsers/web"
	"github.com/icedream/irc-medialink/parsers/wikipedia"
	"github.com/icedream/irc-medialink/parsers/youtube"
	"github.com/icedream/irc-medialink/version"
)

func must(err error) {
	if err == nil {
		return
	}

	log.Fatal(err)
}

func main() {
	fmt.Println(version.MakeHumanReadableVersionString(false, false))
	if timestamp, ok := version.FormattedAppBuildTime(); ok {
		fmt.Printf("\tbuild timestamp: %s\n", timestamp)
	}
	fmt.Printf("\t\u00A9 %d\u2013%d %s\n", 2016, 2020, "Carl Kittelberger")
	fmt.Println("")

	var youtubeAPIKey string

	var soundcloudClientID string
	var soundcloudClientSecret string

	var twitterClientID string
	var twitterClientSecret string

	var webEnableImages bool

	var debug bool
	var noInvite bool
	var useTLS bool
	var server string
	var password string
	var timeout time.Duration
	var pingFreq time.Duration

	ownerNickname := "Icedream"
	ownerChannel := "#MediaLink"

	nickname := version.AppName
	ident := strings.ToLower(version.AppName)
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

	// Support config
	kingpin.Flag("owner-channel", "Channel to refer to for support of this bot instance.").StringVar(&ownerChannel)
	kingpin.Flag("owner-nickname", "User nickname to refer to for support of this bot instance.").StringVar(&ownerNickname)

	// Youtube config
	kingpin.Flag("youtube-key", "The API key to use to access the YouTube API.").StringVar(&youtubeAPIKey)

	// SoundCloud config
	kingpin.Flag("soundcloud-id", "The SoundCloud ID.").StringVar(&soundcloudClientID)
	kingpin.Flag("soundcloud-secret", "The SoundCloud secret.").StringVar(&soundcloudClientSecret)

	// Twitter config
	kingpin.Flag("twitter-id", "The Twitter ID.").StringVar(&twitterClientID)
	kingpin.Flag("twitter-secret", "The Twitter secret.").StringVar(&twitterClientSecret)

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
	if len(youtubeAPIKey) > 0 {
		youtubeParser := &youtube.Parser{
			Config: &youtube.Config{APIKey: youtubeAPIKey},
		}
		must(m.RegisterParser(youtubeParser))
	} else {
		log.Println("No YouTube API key provided, YouTube parsing via API is disabled.")
	}

	// Load soundcloud parser
	if len(soundcloudClientID) > 0 && len(soundcloudClientSecret) > 0 {
		soundcloudParser := &soundcloud.Parser{
			Config: &soundcloud.Config{
				ClientID:     soundcloudClientID,
				ClientSecret: soundcloudClientSecret,
			},
		}
		must(m.RegisterParser(soundcloudParser))
	} else {
		log.Println("No SoundCloud client ID or secret provided, SoundCloud parsing via API is disabled.")
	}

	// Load twitter parser
	if len(twitterClientID) > 0 && len(twitterClientSecret) > 0 {
		twitterParser := &twitter.Parser{
			Config: &twitter.Config{
				ClientID:     twitterClientID,
				ClientSecret: twitterClientSecret,
			},
		}
		must(m.RegisterParser(twitterParser))
	} else {
		log.Println("No Twitter client ID or secret provided, Twitter parsing via API is disabled.")
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
	handleChannelModeChanges := func(channel, modes string) {
		// Is this MODE for a channel?
		isChannel := strings.HasPrefix(channel, "#")

		if !isChannel {
			return
		}

		// TODO - Handle mode params

		add := true
		for _, mode := range modes {
			switch mode {
			case '+':
				add = true
			case '-':
				add = false
			default:
				if add {
					setChannelMode(channel, mode)
				} else {
					unsetChannelMode(channel, mode)
				}
			}
		}

		log.Println("New modes for", channel, "are", getChannelModes(channel))
	}
	conn.AddCallback("MODE", func(e *irc.Event) {
		handleChannelModeChanges(e.Arguments[0], e.Arguments[1])
	})
	conn.AddCallback("324", func(e *irc.Event) { // handle RPL_CHANNELMODEIS
		// TODO - Handle mode params (fourth argument)
		// First argument is actually our nickname here
		handleChannelModeChanges(e.Arguments[1], e.Arguments[2])
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
							conn.Privmsgf(targetChannel, "If you ever run into trouble with me (or find any bugs), please use the channel %s for contact on this IRC.", ownerChannel)
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
	handleText := func(nick, target, source, msg string) {
		msg = stripIrcFormatting(msg)

		// Ignore user if they just joined
		if shouldIgnore := m.TrackUser(target, source); shouldIgnore {
			log.Print("This message will be ignored since the user just joined.")
			return
		}

		urlStr := xurls.Relaxed.FindString(msg)

		// Parse URL!
		u, err := url.ParseRequestURI(urlStr)
		if err != nil {
			u, err = url.ParseRequestURI("http://" + urlStr)
		}
		if err != nil {
			log.Print(err)
			return
		}

		// Check if this URL has been recently parsed before (antiflood)
		shouldIgnore := m.TrackUrl(target, u)
		if shouldIgnore {
			log.Printf("WARNING: URL antiflood triggered, dropping URL for %s: %s", target, u)
			return
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
	conn.AddCallback("NOTICE", func(e *irc.Event) {
		go func(event *irc.Event) {

			// TODO - handle channel notice
			// TODO - handle private noice

			//sender := event.Nick
			target := event.Arguments[0]
			isChannel := true
			if strings.EqualFold(target, conn.GetNick()) {
				// Private notice to us!
				target = event.Nick
				isChannel = false
			}
			if strings.EqualFold(target, conn.GetNick()) {
				// Emergency switch to avoid endless loop,
				// dropping all messages from the bot to the bot!
				log.Printf("BUG - Emergency switch, caught message from bot to bot: %s", event.Arguments)
				return
			}

			msg := event.Message()
			msg = stripIrcFormatting(msg)
			log.Printf("<%s @ %s> Notice: %s", event.Nick, target, msg)

			// Ignore system/internal messages
			if len(e.Nick) <= 0 || len(target) <= 0 ||
				strings.EqualFold(e.Nick, "NickServ") ||
				strings.EqualFold(e.Nick, "ChanServ") ||
				strings.EqualFold(e.Nick, "Global") {
				return
			}

			if !isChannel {
				// Explain who we are and what we do
				conn.Noticef(target, "Hi, I parse links people post to chat rooms to give some information about them. I also allow people to search for YouTube videos and SoundCloud sounds straight from IRC. If you have questions or got any bug reports, please direct them to %s in %s, thank you!", ownerNickname, ownerChannel)
				return
			}

			handleText(event.Nick, target, event.Source, msg)
		}(e)
	})
	// Inject our own version
	conn.RemoveCallback("CTCP_VERSION", 0)
	conn.AddCallback("CTCP_VERSION", func(e *irc.Event) {
		conn.Connection.Notice(e.Nick, (&ctcpMessage{
			Command: "VERSION",
			Params:  []string{version.MakeHumanReadableVersionString(true, false), "based on", irc.VERSION},
		}).String())
	})
	// Inject our own userinfo
	conn.RemoveCallback("CTCP_USERINFO", 0)
	conn.AddCallback("CTCP_USERINFO", func(e *irc.Event) {
		conn.Connection.Notice(e.Nick, (&ctcpMessage{
			Command: "USERINFO",
			Params:  []string{"IRC bot running", version.MakeHumanReadableVersionString(true, true)},
		}).String())
	})
	conn.AddCallback("CTCP_ACTION", func(e *irc.Event) {
		//sender := event.Nick
		target := e.Arguments[0]
		isChannel := true
		if strings.EqualFold(target, conn.GetNick()) {
			// Private message to us!
			target = e.Nick
			isChannel = false
		}
		if strings.EqualFold(target, conn.GetNick()) {
			// Emergency switch to avoid endless loop,
			// dropping all messages from the bot to the bot!
			log.Printf("BUG - Emergency switch, caught message from bot to bot: %s", e.Arguments)
			return
		}

		msg := stripIrcFormatting(strings.Join(e.Arguments, " "))
		log.Printf("<%s @ %s> * %s %s", e.Nick, target, e.Nick, msg)

		// Ignore system/internal messages
		if len(e.Nick) <= 0 || len(target) <= 0 ||
			strings.EqualFold(e.Nick, "NickServ") ||
			strings.EqualFold(e.Nick, "ChanServ") ||
			strings.EqualFold(e.Nick, "Global") {
			return
		}

		if !isChannel {
			// Explain who we are and what we do
			conn.Privmsgf(target, "Hi, I parse links people post to chat rooms to give some information about them. I also allow people to search for YouTube videos and SoundCloud sounds straight from IRC. If you have questions or got any bug reports, please direct them to %s in %s, thank you!", ownerNickname, ownerChannel)
			return
		}

		handleText(e.Nick, target, e.Source, msg)
	})
	conn.AddCallback("CTCP", func(e *irc.Event) {
		if len(e.Arguments) < 1 {
			return
		}

		switch {
		case strings.EqualFold(e.Arguments[0], "FINGER"):
			conn.Connection.Notice(e.Nick, (&ctcpMessage{
				Command: "FINGER",
				Params:  []string{"IRC bot running", version.MakeHumanReadableVersionString(true, true)},
			}).String())
		default:
			// Ignore
		}
	})
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

			// Parse CTCP requests
			msg := event.Message()
			msg = stripIrcFormatting(msg)
			log.Printf("<%s @ %s> Message: %s", event.Nick, target, msg)

			// Ignore system/internal messages
			if len(e.Nick) <= 0 || len(target) <= 0 ||
				strings.EqualFold(e.Nick, "NickServ") ||
				strings.EqualFold(e.Nick, "ChanServ") ||
				strings.EqualFold(e.Nick, "Global") {
				return
			}

			if !isChannel {
				// Explain who we are and what we do
				conn.Privmsgf(target, "Hi, I parse links people post to chat rooms to give some information about them. I also allow people to search for YouTube videos and SoundCloud sounds straight from IRC. If you have questions or got any bug reports, please direct them to %s in %s, thank you!", ownerNickname, ownerChannel)
				return
			}

			handleText(event.Nick, target, event.Source, msg)
		}(e)
	})

	// connect
	must(conn.Connect(server))

	// listen for signals
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigc
		log.Println("Requesting bot shutdown due to received signal:", sig)
		conn.Quit()
	}()

	log.Print("Now looping.")
	conn.Loop()
}
