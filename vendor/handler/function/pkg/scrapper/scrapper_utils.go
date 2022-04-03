package scrapper

import (
	"strconv"
	"strings"
)

// Resolves an image url. If the image is stored on the roll20 CDN
// adds the roll20 prefix, if its stored elsewhere (AWS), return the URL as is
func getAbsoluteUrlToImage(url string) string {
	if strings.HasPrefix(url, "/") {
		return "https://app.roll.net" + url
	}
	return url
}

// Given a player details url, extracts the roll20ID
func getPlayerIdFromURL(url string) (int, error) {
	splitUrl := strings.Split(url, "/")
	playerRoll20Id, err := strconv.Atoi(splitUrl[len(splitUrl)-1])
	if err != nil {
		return -1, err
	}
	return playerRoll20Id, nil
}
