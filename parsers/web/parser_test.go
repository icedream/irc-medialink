package web

import (
	"bytes"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/icedream/irc-medialink/parsers"
)

const (
	validTestHTMLTitle   = "Testing"
	validTestImageWidth  = 0xea
	validTestImageHeight = 0xae
)

var (
	validTestHTML  = fmt.Sprintf(`<!doctype html><html><head><title>%s</title></head><body><h1>Testing</h1></body></html>`, html.EscapeString(validTestHTMLTitle))
	validTestImage image.Image
	validTestGIF   []byte
	validTestPNG   []byte
	validTestJPEG  []byte
)

func init() {
	// generate random image
	validTestImageRect := image.Rect(0, 0, validTestImageWidth, validTestImageHeight)
	validTestImage := image.NewAlpha(validTestImageRect)
	for y := 0; y < validTestImage.Rect.Dy(); y++ {
		for x := 0; x < validTestImage.Rect.Dx(); x++ {
			v := rand.Uint32()
			validTestImage.Set(x, y, color.RGBA{
				R: uint8(v % 0xff),
				G: uint8((v >> 8) % 0xff),
				B: uint8((v >> 16) % 0xff),
				A: uint8((v >> 24) % 0xff),
			})
		}
	}

	// compress to different formats for parsers
	validTestEncodeBuffer := new(bytes.Buffer)
	if err := png.Encode(validTestEncodeBuffer, validTestImage); err != nil {
		panic(err)
	}
	validTestPNG = make([]byte, validTestEncodeBuffer.Len())
	copy(validTestPNG, validTestEncodeBuffer.Bytes())
	validTestEncodeBuffer.Reset()

	if err := jpeg.Encode(validTestEncodeBuffer, validTestImage, &jpeg.Options{}); err != nil {
		panic(err)
	}
	validTestJPEG = make([]byte, validTestEncodeBuffer.Len())
	copy(validTestJPEG, validTestEncodeBuffer.Bytes())
	validTestEncodeBuffer.Reset()

	if err := gif.Encode(validTestEncodeBuffer, validTestImage, &gif.Options{}); err != nil {
		panic(err)
	}
	validTestGIF = make([]byte, validTestEncodeBuffer.Len())
	copy(validTestGIF, validTestEncodeBuffer.Bytes())
	validTestEncodeBuffer.Reset()
}

func getDefaultHTMLResponder() httpmock.Responder {
	header := http.Header{}
	header.Set("content-type", "text/html; charset=utf-8")
	return httpmock.ResponderFromResponse(&http.Response{
		Status:        fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)),
		StatusCode:    http.StatusOK,
		Body:          httpmock.NewRespBodyFromString(validTestHTML),
		Header:        header,
		ContentLength: int64(len([]byte(validTestHTML))),
	})
}

func getDefaultGIFResponder() httpmock.Responder {
	header := http.Header{}
	header.Set("content-type", "image/gif")
	header.Set("content-disposition", "attachment; filename=\"test.gif\"")
	return httpmock.ResponderFromResponse(&http.Response{
		Status:        fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)),
		StatusCode:    http.StatusOK,
		Body:          httpmock.NewRespBodyFromBytes(validTestGIF),
		Header:        header,
		ContentLength: int64(len(validTestGIF)),
	})
}

func getDefaultPNGResponder() httpmock.Responder {
	header := http.Header{}
	header.Set("content-type", "image/png")
	header.Set("content-disposition", "attachment; filename=\"test.png\"")
	return httpmock.ResponderFromResponse(&http.Response{
		Status:        fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)),
		StatusCode:    http.StatusOK,
		Body:          httpmock.NewRespBodyFromBytes(validTestPNG),
		Header:        header,
		ContentLength: int64(len(validTestPNG)),
	})
}

func getDefaultJPEGResponder() httpmock.Responder {
	header := http.Header{}
	header.Set("content-type", "image/jpeg")
	header.Set("content-disposition", "attachment; filename=\"test.jpeg\"")
	return httpmock.ResponderFromResponse(&http.Response{
		Status:        fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)),
		StatusCode:    http.StatusOK,
		Body:          httpmock.NewRespBodyFromBytes(validTestJPEG),
		Header:        header,
		ContentLength: int64(len(validTestJPEG)),
	})
}

func mustNewParser(t *testing.T) *Parser {
	p := new(Parser)
	p.Config = Config{}
	if !assert.Nil(t, p.Init(), "Parser.Init must throw no errors") {
		panic("Can't run test without a proper parser")
	}
	return p
}

func parseWithTimeout(p *Parser, t *testing.T, timeout time.Duration, u *url.URL, ref *url.URL) (retval parsers.ParseResult) {
	resultChan := make(chan parsers.ParseResult)
	go func(resultChan chan<- parsers.ParseResult, p *Parser, u *url.URL, ref *url.URL) {
		resultChan <- p.Parse(u, ref)
	}(resultChan, p, u, ref)

	select {
	case r := <-resultChan:
		retval = r
		return
	case <-time.After(timeout):
		t.Fatal("Didn't succeed parsing URL in time")
		return
	}
}

func Test_Parser_Parse_Simple(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://example.com/test",
		getDefaultHTMLResponder())

	p := mustNewParser(t)
	originalURL := &url.URL{
		Scheme: "http",
		Host:   "example.com",
		Path:   "/test",
	}
	result := p.Parse(originalURL, nil)

	require.Equal(t, httpmock.GetTotalCallCount(), 1)

	t.Logf("Result: %+v", result)
	require.False(t, result.Ignored)
	require.Nil(t, result.Error)
	require.Nil(t, result.UserError)
	require.Len(t, result.Information, 1)
	require.Equal(t, validTestHTMLTitle, result.Information[0]["Title"])
}

func Test_Parser_Parse_IRCBotScience_NoTitle(t *testing.T) {
	p := mustNewParser(t)
	result := p.Parse(&url.URL{
		Scheme: "https",
		Host:   "irc-bot-science.clsr.net",
		Path:   "notitle",
	}, nil)

	t.Logf("Result: %+v", result)
	require.False(t, result.Ignored)
	require.Nil(t, result.Error)
	require.Nil(t, result.UserError)
	require.Len(t, result.Information, 1)
	require.Equal(t, noTitleStr, result.Information[0]["Title"])
}

func Test_Parser_Parse_IRCBotScience_LongHeaders(t *testing.T) {
	p := mustNewParser(t)
	result := parseWithTimeout(p, t, 5*time.Second, &url.URL{
		Scheme: "https",
		Host:   "irc-bot-science.clsr.net",
		Path:   "longheaders",
	}, nil)
	for result.FollowURL != nil {
		result = parseWithTimeout(p, t, 5*time.Second, result.FollowURL, nil)
	}

	t.Logf("Result: %+v", result)
	// It just shouldn't panic. Erroring out is fine.
	require.True(t, result.Ignored || result.Error != nil, result.Ignored)
}

func Test_Parser_Parse_IRCBotScience_BigHeader(t *testing.T) {
	p := mustNewParser(t)
	result := parseWithTimeout(p, t, 5*time.Second, &url.URL{
		Scheme: "https",
		Host:   "irc-bot-science.clsr.net",
		Path:   "bigheader",
	}, nil)
	for result.FollowURL != nil {
		result = parseWithTimeout(p, t, 5*time.Second, result.FollowURL, nil)
	}

	t.Logf("Result: %+v", result)
	// It just shouldn't panic. Erroring out is fine.
	require.True(t, result.Ignored || result.Error != nil)
}

func Test_Parser_Parse_IRCBotScience_Large(t *testing.T) {
	p := mustNewParser(t)

	result := parseWithTimeout(p, t, 5*time.Second, &url.URL{
		Scheme: "https",
		Host:   "irc-bot-science.clsr.net",
		Path:   "large",
	}, nil)
	for result.FollowURL != nil {
		result = parseWithTimeout(p, t, 5*time.Second, result.FollowURL, nil)
	}

	t.Logf("Result: %+v", result)
	require.False(t, result.Ignored)
	require.Nil(t, result.Error)
	require.Nil(t, result.UserError)
	require.Len(t, result.Information, 1)
	require.Equal(t, "If this title is printed, it works correctly.", result.Information[0]["Title"])
}

func Test_Parser_Parse_IRCBotScience_Redirect(t *testing.T) {
	p := mustNewParser(t)
	originalURL := &url.URL{
		Scheme: "https",
		Host:   "irc-bot-science.clsr.net",
		Path:   "redirect",
	}
	result := p.Parse(originalURL, nil)

	t.Logf("Result: %+v", result)
	require.False(t, result.Ignored)
	require.Nil(t, result.Error)
	require.Nil(t, result.UserError)
	require.NotNil(t, result.FollowURL)
	require.Equal(t, originalURL.String(), result.FollowURL.String())
}

func Test_Parser_Parse_Hash(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://example.com/test",
		getDefaultHTMLResponder())

	p := mustNewParser(t)
	originalURL := &url.URL{
		Scheme:   "http",
		Host:     "example.com",
		Path:     "/test",
		Fragment: "invalid",
	}
	result := p.Parse(originalURL, nil)

	require.Equal(t, httpmock.GetTotalCallCount(), 1)

	t.Logf("Result: %+v", result)
	require.False(t, result.Ignored)
	require.Nil(t, result.Error)
	require.Nil(t, result.UserError)
}

func Test_Parser_Parse_Image_GIF(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://example.com/test",
		getDefaultGIFResponder())

	p := mustNewParser(t)
	p.Config.EnableImages = true
	originalURL := &url.URL{
		Scheme: "http",
		Host:   "example.com",
		Path:   "/test",
	}
	result := p.Parse(originalURL, nil)

	require.Equal(t, httpmock.GetTotalCallCount(), 1)

	t.Logf("Result: %+v", result)
	require.False(t, result.Ignored)
	require.Nil(t, result.Error)
	require.Nil(t, result.UserError)
	require.Len(t, result.Information, 1)
	point, ok := (result.Information[0]["ImageSize"]).(image.Point)
	require.True(t, ok)
	require.EqualValues(t, validTestImageWidth, point.X)
	require.EqualValues(t, validTestImageHeight, point.Y)
	require.Equal(t, "GIF", result.Information[0]["ImageType"])
	require.EqualValues(t, len(validTestGIF), result.Information[0]["Size"])
}

func Test_Parser_Parse_Image_PNG(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://example.com/test",
		getDefaultPNGResponder())

	p := mustNewParser(t)
	p.Config.EnableImages = true
	originalURL := &url.URL{
		Scheme: "http",
		Host:   "example.com",
		Path:   "/test",
	}
	result := p.Parse(originalURL, nil)

	require.Equal(t, httpmock.GetTotalCallCount(), 1)

	t.Logf("Result: %+v", result)
	require.False(t, result.Ignored)
	require.Nil(t, result.Error)
	require.Nil(t, result.UserError)
	require.Len(t, result.Information, 1)
	point, ok := (result.Information[0]["ImageSize"]).(image.Point)
	require.True(t, ok)
	require.EqualValues(t, validTestImageWidth, point.X)
	require.EqualValues(t, validTestImageHeight, point.Y)
	require.Equal(t, "PNG", result.Information[0]["ImageType"])
	require.EqualValues(t, len(validTestPNG), result.Information[0]["Size"])
}

func Test_Parser_Parse_Image_JPEG(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://example.com/test",
		getDefaultJPEGResponder())

	p := mustNewParser(t)
	p.Config.EnableImages = true
	originalURL := &url.URL{
		Scheme: "http",
		Host:   "example.com",
		Path:   "/test",
	}
	result := p.Parse(originalURL, nil)

	require.Equal(t, httpmock.GetTotalCallCount(), 1)

	t.Logf("Result: %+v", result)
	require.False(t, result.Ignored)
	require.Nil(t, result.Error)
	require.Nil(t, result.UserError)
	require.Len(t, result.Information, 1)
	point, ok := (result.Information[0]["ImageSize"]).(image.Point)
	require.True(t, ok)
	require.EqualValues(t, validTestImageWidth, point.X)
	require.EqualValues(t, validTestImageHeight, point.Y)
	require.Equal(t, "JPEG", result.Information[0]["ImageType"])
	require.EqualValues(t, len(validTestJPEG), result.Information[0]["Size"])
}
