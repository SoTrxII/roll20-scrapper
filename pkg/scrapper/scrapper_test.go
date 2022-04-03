package scrapper

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

func SetupTestServer(campaignDataPath string) *httptest.Server {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Dir(filename)
	sample, _ := os.Open(path.Join(dir, campaignDataPath))
	sampleData, _ := ioutil.ReadAll(sample)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/campaigns/details/") {
			w.Write(sampleData)
		} else {
			w.WriteHeader(200)
		}
	}))
	os.Setenv("ROLL20_BASE_URL", mockServer.URL)
	os.Setenv("ROLL20_USERNAME", "mock")
	os.Setenv("ROLL20_PASSWORD", "mock")
	return mockServer

}

// Setup a server always answering the status code
func SetupConstantServer(code int) *httptest.Server {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}))
	os.Setenv("ROLL20_BASE_URL", mockServer.URL)
	os.Setenv("ROLL20_USERNAME", "mock")
	os.Setenv("ROLL20_PASSWORD", "mock")
	return mockServer
}

func countProp(players []Player, cond func(player *Player) bool) uint8 {
	// We kind of need a true reduce implementation in Go
	var count uint8
	for _, p := range players {
		if cond(&p) {
			count++
		}
	}
	return count
}

// Count the number of GMs in a Roll20 game
func countGMs(players []Player) uint8 {
	return countProp(players, func(player *Player) bool {
		return player.IsGm
	})
}

// Count the number of Players in a Roll20 game, excluding GMs
func countPlayers(players []Player) uint8 {
	return countProp(players, func(player *Player) bool {
		return !player.IsGm
	})
}

// A single GM (the game creator) in the Roll20 game
func TestSingleGM(t *testing.T) {
	mockServer := SetupTestServer("../../assets/sample_campaign_page.html")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	assert.Nil(t, err)
	var players *[]Player
	players, err = scrapper.GetPlayers("")
	assert.Nil(t, err)
	// A single GM
	assert.Equal(t, uint8(1), countGMs(*players))
	// And 6 players
	assert.Equal(t, uint8(7), countPlayers(*players))
	mockServer.Close()

}

// Multiple GMs (the game creator + some players with GM perms) in the Roll20 Game
func TestMultipleGMs(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_page_multiple_gms.html")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	assert.Nil(t, err)
	var players *[]Player
	players, err = scrapper.GetPlayers("")
	assert.Nil(t, err)

	// Twos GM
	assert.Equal(t, uint8(2), countGMs(*players))
	// And 9 players
	assert.Equal(t, uint8(9), countPlayers(*players))
	mockServer.Close()
}

// Some players couldn't be parsed for some reason. This may never happen again but has been the
// case at least once
func TestMissingPlayers(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_missing_id.html")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	assert.Nil(t, err)
	var players *[]Player
	players, err = scrapper.GetPlayers("")
	assert.Error(t, err, IncompleteError{})

	// One GM
	assert.Equal(t, uint8(1), countGMs(*players))
	// 5 non-ignored players
	assert.Equal(t, uint8(5), countPlayers(*players))
	// 2 ignored players (so one comma to split them)
	assert.Equal(t, 1, strings.Count(err.Error(), ","))
	assert.Contains(t, err.Error(), "Ignored Player 1")
	assert.Contains(t, err.Error(), "Ignored Player 2")
	mockServer.Close()
}

func TestMissingGM(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_missing_gm.html")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	assert.Nil(t, err)
	players, err := scrapper.GetPlayers("")
	assert.Error(t, err)
	assert.Nil(t, players)
	mockServer.Close()
}

func TestMissingAvatar(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_no_gm_avatar.html")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	assert.Nil(t, err)
	players, err := scrapper.GetPlayers("")
	// A single GM
	assert.Equal(t, uint8(1), countGMs(*players))
	// And 6 players
	assert.Equal(t, uint8(7), countPlayers(*players))
	mockServer.Close()
}

// Simply joining a game as a player
func TestJoinGame(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_page_multiple_gms.html")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	assert.Nil(t, err)
	err = scrapper.JoinGame("", "")
	assert.Nil(t, err)
	mockServer.Close()
}

// Simply joining a game as a player
func TestJoinWrongURL(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_page_multiple_gms.html")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	assert.Nil(t, err)
	err = scrapper.JoinGame("%DSLSDLSMM", "##MMMD%%")
	assert.Error(t, err)
	mockServer.Close()
}

// Simply joining a game as a player
func TestFailingJoiningGame(t *testing.T) {
	// Login with a good server
	mockServer := SetupConstantServer(200)
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	mockServer.Close()
	assert.Nil(t, err)

	// Swapping the good server for a failing one
	mockServer = SetupConstantServer(500)
	scrapper.baseUrl = mockServer.URL
	err = scrapper.JoinGame("", "")
	assert.Error(t, err)
	mockServer.Close()
}

// When the autologin on scrapper creation fail, end gracefully
func TestFailingLogin(t *testing.T) {
	mockServer := SetupConstantServer(500)
	_, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	assert.Error(t, err)
	mockServer.Close()
}

// Corrupted DOM / Roll20 update
func TestInvalidDocument(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/non_existant.html")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"})
	assert.Nil(t, err)
	var players *[]Player
	players, err = scrapper.GetPlayers("")
	assert.Error(t, err)
	assert.Nil(t, players)
	mockServer.Close()
}
