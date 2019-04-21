package apm

import (
	"github.com/stretchr/testify/assert"
	apmmodel "go.elastic.co/apm/model"
	"net/url"
	"testing"
)

func TestTagsToAPM(t *testing.T) {
	assert.ElementsMatch(t, TagsToAPM(map[string]string{
		"hello":       "world",
		"http.status": "200",
	}), apmmodel.StringMap{
		apmmodel.StringMapItem{
			Key:   "hello",
			Value: "world",
		},
		apmmodel.StringMapItem{
			Key:   "http_status",
			Value: "200",
		},
	})
}

func TestTagsToURL(t *testing.T) {
	parsed, _ := url.Parse("http://google.com:8080/hello")
	assert.Equal(t, TagsToURL(map[string]string{
		"http.host": "google.com:8080",
		"http.path": "/hello",
	}), UrlToAPM(*parsed))
}
