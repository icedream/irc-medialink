package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TemplateFuncs_comma(t *testing.T) {
	f := tplFuncMap["comma"].(func(num uint64) string)

	assert.Equal(t, "1", f(1))
	assert.Equal(t, "10", f(10))
	assert.Equal(t, "100", f(100))
	assert.Equal(t, "1,000", f(1000))
	assert.Equal(t, "1,000,000", f(1000000))
}

func Test_TemplateFuncs_compactnum(t *testing.T) {
	f := tplFuncMap["compactnum"].(func(num uint64) string)

	assert.Equal(t, "1", f(1))
	assert.Equal(t, "10", f(10))
	assert.Equal(t, "100", f(100))
	assert.Equal(t, "999", f(999))
	assert.Equal(t, "1k", f(1000))
	assert.Equal(t, "9k", f(9000))
	assert.Equal(t, "9.9k", f(9900))
	assert.Equal(t, "10k", f(9999))
	assert.Equal(t, "10k", f(10000))
	assert.Equal(t, "100k", f(100000))
	assert.Equal(t, "999k", f(999000))
	assert.Equal(t, "999.9k", f(999900))
	assert.Equal(t, "1M", f(999999))
	assert.Equal(t, "1M", f(1000000))
}
