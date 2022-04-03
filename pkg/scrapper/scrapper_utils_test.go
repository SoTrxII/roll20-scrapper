package scrapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoll20Image(t *testing.T) {
	const RELATIVE_URL = "/foo/bar"
	const R20_BASE_URL = "https://app.roll.net"
	url := getAbsoluteUrlToImage(RELATIVE_URL)
	assert.Equal(t, R20_BASE_URL+RELATIVE_URL, url)
}

func TestExternalImage(t *testing.T) {
	const ABSOLUTE_URL = "foo/bar"
	url := getAbsoluteUrlToImage(ABSOLUTE_URL)
	assert.Equal(t, ABSOLUTE_URL, url)
}

func TestGetPlayerIDAbsoluteURL(t *testing.T) {
	const URL = "https://app.roll20.net/user/1"
	id, err := getPlayerIdFromURL(URL)
	assert.Nil(t, err)
	assert.Equal(t, 1, id)
}

func TestGetPlayerIDRelativeURL(t *testing.T) {
	const URL = "/user/2"
	id, err := getPlayerIdFromURL(URL)
	assert.Nil(t, err)
	assert.Equal(t, 2, id)
}

func TestGetPlayerIDWrongInput(t *testing.T) {
	const URL = ".."
	id, err := getPlayerIdFromURL(URL)
	assert.Error(t, err)
	assert.Equal(t, -1, id)
}
