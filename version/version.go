package version

import (
	"strconv"
	"time"
)

var (
	// AppName is the application name
	AppName = "MediaLink"

	// AppVersion is the version string generated from `git describe`.
	AppVersion = ""

	// appBuildTimestampStr is the build time as Unix timestamp (seconds).
	// This is a string field to allow build-time override via ldflag -X.
	appBuildTimestampStr = ""

	// AppSourceURL is the URL to the source code for this application.
	AppSourceURL = "https://github.com/icedream/irc-medialink"
)

// AppBuildTime attempts to convert the built-time set string for the build timestamp to a Golang-native time.Time instance.
// Returns ok = false if either no info exists or conversion fails,
// otherwise returns ok = true and the converted time.
func AppBuildTime() (timestamp time.Time, ok bool) {
	if len(appBuildTimestampStr) > 0 {
		if parsedInt, err := strconv.ParseInt(appBuildTimestampStr, 10, 64); err == nil {
			ok = true
			timestamp = time.Unix(parsedInt, 0)
		}
	}
	return
}

// FormattedAppBuildTime attempts to generate a properly RFC1123-formatted timestamp from the built-time set string for the build timestamp.
// Returns infoExists = false if that information simply does not exist,
// Otherwise returns infoExists = true and either a properly RFC1123-formatted timestamp or, if conversion to time.Time internally fails, just the original timestamp string.
func FormattedAppBuildTime() (timestamp string, infoExists bool) {
	t, ok := AppBuildTime()
	if ok {
		timestamp = t.Format(time.RFC1123)
		infoExists = true
		return
	}

	if len(appBuildTimestampStr) > 0 {
		timestamp = appBuildTimestampStr
		infoExists = true
		return
	}

	return
}

// MakeHumanReadableVersionString generates a human-readable one-line string from the information stored in the binary at build-time.
func MakeHumanReadableVersionString(addBuildInfo bool, addSourceCodeInfo bool) string {
	versionString := AppName
	if len(AppVersion) > 0 {
		versionString += " " + AppVersion
	}
	if addBuildInfo {
		if timestamp, ok := FormattedAppBuildTime(); ok {
			versionString += " built " + timestamp
		}
	}
	if addSourceCodeInfo && len(AppSourceURL) > 0 {
		versionString += " - source code available at " + AppSourceURL
	}
	return versionString
}
