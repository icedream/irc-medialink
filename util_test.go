package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustParseUrl(u string) *url.URL {
	if uri, err := url.Parse(u); err == nil {
		return uri
	} else {
		panic(err)
	}
}

func Test_GetYouTubeId(t *testing.T) {
	assert.Equal(t, "aYz-9jUlav-", getYouTubeId(mustParseUrl("http://youtube.com/watch?v=aYz-9jUlav-")))
	assert.Equal(t, "aYz-9jUlav-", getYouTubeId(mustParseUrl("https://youtube.com/watch?v=aYz-9jUlav-")))
	assert.Equal(t, "aYz-9jUlav-", getYouTubeId(mustParseUrl("http://www.youtube.com/watch?v=aYz-9jUlav-")))
	assert.Equal(t, "aYz-9jUlav-", getYouTubeId(mustParseUrl("https://www.youtube.com/watch?v=aYz-9jUlav-")))
	assert.Equal(t, "aYz-9jUlav-", getYouTubeId(mustParseUrl("http://youtu.be/aYz-9jUlav-")))
	assert.Equal(t, "aYz-9jUlav-", getYouTubeId(mustParseUrl("https://youtu.be/aYz-9jUlav-")))
	assert.Equal(t, "aYz-9jUlav-", getYouTubeId(mustParseUrl("http://www.youtu.be/aYz-9jUlav-")))
	assert.Equal(t, "aYz-9jUlav-", getYouTubeId(mustParseUrl("https://www.youtu.be/aYz-9jUlav-")))
}

func Benchmark_GetYouTubeId(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getYouTubeId(mustParseUrl("http://youtube.com/watch?v=aYz-9jUlav-"))
	}
}
