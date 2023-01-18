# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog] and this project adheres to
[Semantic Versioning].

Types of changes are:
* **Security** in case of vulnerabilities.
* **Deprecated** for soon-to-be removed features.
* **Added** for new features.
* **Changed** for changes in existing functionality.
* **Removed** for now removed features.
* **Fixed** for any bug fixes.

## [Unreleased]
### Fixed
- `https://youtube.com/@alias` style URLs now are properly detected as channels.
- Fix Japanese text excerpts sometimes cutting off text in the middle of
  multibyte characters, turning them into placeholder characters.

### Changed
- Web parser will now scan for `og:url` (canonical URL) of a page and do a
  second request against it. This is necessary for the detection of
  `https://youtube.com/@alias` style URLs.
- The maximum HTML body size parsed by the web parser has been increased to 1 MB
  as YouTube's Open Graph meta tags are embedded a lot further into the document
  than usual.
- Web parser will now use a more descriptive user agent, including segments that
  match better against known crawler patterns. This should improve parsing of
  websites which would otherwise require user interaction. The new user agent
  format will look like this with `<version>` being the MediaLink version:

  `Mozilla/5.0 (compatible; MediaLink/<version>; bot; Go-http-client/1.1; like WhatsApp/2.*; +https://github.com/icedream/irc-medialink) MediaLink/<version>`
- URLs marked as articles in Open Graph (e. g. Twitter and Mastodon posts, blog
  pages…) will now have an excerpt of the description included in chat.

## [1.2.0] - 2023-01-17
### Added
* Implement Open Graph support.
* Allow configuration of accepted languages for websites (`--web-language=…`, defaults to `*`).
* Add parser for Reddit subreddits and posts.

### Changed
* Update several dependencies.
* Update header for parsed YouTube links to be closer to the modern logo.

### Fixed
* Fix antiflood handling of punycode URLs.
* Fix antiflood not taking ports in URLs into account.
* Fix bot potentially getting blacklisted due to user agent.
* Parsing can now time out properly for each URL (`--parse-timeout=…`, defaults to `10s`).
* Fix internal cache errors being ignored.
* Fix potential parser crosstalk in URL handling.
* Fix YouTube `/live/` video paths not being handled at all.

## [1.1.3] - 2022-02-11
### Changed
* Update several dependencies.

## [1.1.2] - 2021-12-27
### Changed
* Update several dependencies.
  - Adapt to go-ircevent changes to how they handle `CTCP VERSION`.

## [1.1.1] - 2020-08-02
### Added
* Implement proper CTCP support and build metadata. #15
* Bot now handles CTCP VERSION and USERINFO.
* Bot can now start query with inviters to allow joining channels with keys. #16

### Changed
* Improve YouTube live stream details output.
  - Mark YouTube live streams with white on red LIVE prefix.
  - Detect past and upcoming YouTube live streams properly.
  - Add YouTube live uptime info.

### Fixed
* Fix channel mode parsing: In some channels the bot was unable to properly detect colors being blocked.
* Improve code to strip IRC formatting.
* Fix crash in antiflood due to server-internal nicknames.
* Fix handling of invalid links.
* Fix image stats output no longer working.
* Gracefully handle connections errors on startup.
* Fix Twitter timestamp parsing.
* Fix numeric types passed to template.
* Fix wrong name version being used for crediting tweets to authors.
* Handle issues with joining channels properly. #16

## [1.1.0] - 2020-07-18
### Added
* Register channel modes and changes, strip formatting if channel is `+c`.
* Implement Twitter URL analysis via API.

### Changed
* Increase default max HTML size to 32k.
* Explicitly state to web servers that any content language is fine.

### Fixed
* Fix missing template files in Docker image.
* Fix missing CA certificates on Alpine Docker image.
* Fix access to SoundCloud API.
* Adjust time format because SoundCloud randomly changed it.

## [1.0.2] - 2016-07-06
### Added
* Introduce `--images` flag to enable image link parsing and disable image link parsing by default.
* Add filename to image information.
* Allow disabling auto-join on invite (`--no-invite`).

### Changed
* Ignore users for 30 seconds after they join.
* Do not register SoundCloud/YouTube parsers if no auth details given for them.

### Fixed
* Use unique user agent to get around HTTP error 429 on Reddit. #12

## [1.0.1] - 2016-06-11
### Changed
* YouTube parser will onw add NSFW marker for age-restricted videos.
* Use actual cross char (×) for image dimensions. #7

### Fixed
* Fix YouTube link info having bad color formatting on AndroIRC. #11
* Fix endless redirects caused by crosstalk between parsers. #1
* Fix web parser.
  - Fix bot being vulnerable to oversized HTML content. #2
  - Fix bot being vulnerable to multi-line titles. #2
  - Fix hashes being sent as part of requests. #8
* Fix wikipedia parser.
  - Fix handling of www.wikipedia.org and wikipedia.org links.
  - Fix HTTP(S) URL filter logic.
* Fix youtube parser.
  - Fix information for channels.
* Fix "us" typo in join message.

## [1.0.0]
### Added
* Add parsing of YouTube links via API.
* Add parsing of SoundCloud links via API.
* Add parsing of Wikipedia links via API.
* Add parsing of image links.
* Add handling of generic HTML pages, will print title.

[Unreleased]: https://github.com/icedream/irc-medialink/compare/v1.2.0..vHEAD
[1.2.0]: https://github.com/icedream/irc-medialink/compare/v1.1.3..v1.2.0
[1.1.3]: https://github.com/icedream/irc-medialink/compare/v1.1.2..v1.1.3
[1.1.2]: https://github.com/icedream/irc-medialink/compare/v1.1.1..v1.1.2
[1.1.1]: https://github.com/icedream/irc-medialink/compare/v1.1.0..v1.1.1
[1.1.0]: https://github.com/icedream/irc-medialink/compare/v1.0.2..v1.1.0
[1.0.1]: https://github.com/icedream/irc-medialink/compare/v1.0.0..v1.0.1
[1.0.0]: https://github.com/icedream/irc-medialink/releases/v1.0.0
[1.0.2]: https://github.com/icedream/irc-medialink/compare/v1.0.1..v1.0.2

[Keep a Changelog]: http://keepachangelog.com/en/1.0.0/
[Semantic Versioning]: http://semver.org/spec/v2.0.0.html

[_release_link_format]: https://github.com/icedream/irc-medialink/compare/v{previous_tag}..v{tag}
[_breaking_change_token]: BREAKING
