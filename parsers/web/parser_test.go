package web

import (
	"net/url"
	"testing"
	"time"

	"github.com/icedream/irc-medialink/parsers"
	"github.com/stretchr/testify/assert"
)

func mustNewParser(t *testing.T) *Parser {
	p := new(Parser)
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

func Test_Parser_Parse_IRCBotScience_NoTitle(t *testing.T) {
	p := mustNewParser(t)
	result := p.Parse(&url.URL{
		Scheme: "https",
		Host:   "irc-bot-science.clsr.net",
		Path:   "notitle",
	}, nil)

	t.Logf("Result: %+v", result)
	assert.False(t, result.Ignored)
	assert.Nil(t, result.Error)
	assert.Nil(t, result.UserError)
	assert.Len(t, result.Information, 1)
	assert.Equal(t, noTitleStr, result.Information[0]["Title"])
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
	assert.True(t, result.Ignored)
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
	assert.True(t, result.Ignored)
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
	assert.False(t, result.Ignored)
	assert.Nil(t, result.Error)
	assert.Nil(t, result.UserError)
	assert.Len(t, result.Information, 1)
	assert.Equal(t, "If this title is printed, it works correctly.", result.Information[0]["Title"])

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
	assert.False(t, result.Ignored)
	assert.Nil(t, result.Error)
	assert.Nil(t, result.UserError)
	assert.NotNil(t, result.FollowURL)
	assert.Equal(t, originalURL.String(), result.FollowURL.String())
}

func Test_Parser_Parse_Hash(t *testing.T) {
	p := mustNewParser(t)
	originalURL := &url.URL{
		Scheme: "https",
		Host:   "www.google.com",
		Path:   "/#invalid",
	}
	result := p.Parse(originalURL, nil)

	t.Logf("Result: %+v", result)
	assert.False(t, result.Ignored)
	assert.Nil(t, result.Error)
	assert.Nil(t, result.UserError)
}
