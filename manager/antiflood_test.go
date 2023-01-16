package manager_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/icedream/irc-medialink/manager"
)

var (
	punycodeURL1 = &url.URL{
		Scheme: "https",
		Host:   "💻.icedream.pw",
		Path:   "/",
	}
	punycodeURL2 = &url.URL{
		Scheme: "https",
		Host:   "xn--3s8h.icedream.pw",
		Path:   "/",
	}
)

func TestAntiflood_URL_Punycode(t *testing.T) {
	m := manager.NewManager()
	t.Log(punycodeURL1.String())
	shouldIgnore, err := m.TrackUrl("test", punycodeURL1)
	require.NoError(t, err)
	require.False(t, shouldIgnore)
	shouldIgnore, err = m.TrackUrl("test", punycodeURL2)
	require.NoError(t, err)
	require.True(t, shouldIgnore)
}
