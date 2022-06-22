# MediaLink IRC Bot

[![Build Status](https://travis-ci.org/icedream/irc-medialink.svg?branch=master)](https://travis-ci.org/icedream/irc-medialink)

This IRC bot automatically parses links posted in chat rooms and prints information about them.

Currently explicit support has been built in for:

- YouTube
- SoundCloud
- Twitter
- Wikipedia
- Direct image links

Generally, for websites that are not directly supported the bot will print the page title.

## How to run the bot

In order to properly run the bot, you need to [register a SoundCloud application](http://soundcloud.com/you/apps/new) and [get a YouTube Data API key](https://console.developers.google.com/apis/api/youtube/overview) for it and then feed the API data to the bot through the command line arguments.

The bot can be installed through Go using this command:

	go install github.com/icedream/irc-medialink

Then you can find out which options you can pass to the bot directly by running (assuming you put your `$GOPATH/bin` folder into your `PATH`):

	irc-medialink --help

You need to at least pass the `--server`, `--youtube-key`, `--soundcloud-id` and `--soundcloud-secret` parameters.

### ...with Docker

You can use the `icedream/irc-medialink` image in order to run this bot in Docker. You can pull it using this command:

	docker pull icedream/irc-medialink

An example with docker-compose would look like this:

```yaml
version: '2'

services:
  mybot:
    image: icedream/irc-medialink
    command:
      - go-wrapper
      - run
      - --youtube-key=<insert your youtube key here>
      - --soundcloud-id=<insert your soundcloud id here>
      - --soundcloud-secret=<insert your soundcloud secret here>
      - --twitter-id=<insert your soundcloud id here>
      - --twitter-secret=<insert your soundcloud secret here>
      - --server=<irc server host>:<irc server port>
      - --nickserv-pw=<nickserv password>
      - --nick=MyBot
      - --ident=botty
      - --password=<server password>
      - --channels=#channel1,#channel2,...
    restart: always
```

## Support

This bot is officially tested and running on the Rizon IRC network (irc.rizon.net) though also being able to run on other IRC networks.

For support on Rizon IRC please use the channel #MediaLink there to get in contact with Icedream.

## License

This project is licensed under the **GNU General Public License Version 2 or later**. For more information check the [LICENSE](LICENSE) file.
