#!/bin/sh -e
go build \
    -ldflags '-X github.com/icedream/irc-medialink/version.AppVersion='$(git describe --dirty --tags --always)' -X github.com/icedream/irc-medialink/version.appBuildTimestampStr='$(date +%s)' -X github.com/icedream/irc-medialink/version.AppName='"${APPLICATION_NAME:-MediaLink}"' -X github.com/icedream/irc-medialink/version.SupportIRCChannel='"${SUPPORT_IRC_CHANNEL:-"#medialink"}"' '"${EXTRA_LDFLAGS}" \
    "$@"
