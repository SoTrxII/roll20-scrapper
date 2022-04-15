package scrapper

import (
	"fmt"
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

func SetupTestServer(campaignDataPath string, urlMatch string) *httptest.Server {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Dir(filename)
	sample, _ := os.Open(path.Join(dir, campaignDataPath))
	sampleData, _ := ioutil.ReadAll(sample)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, urlMatch) {
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
	mockServer := SetupTestServer("../../assets/sample_campaign_page.html", "/campaigns/details/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	var players *[]Player
	players, err = scrapper.GetPlayers("")
	assert.Nil(t, err)
	// A single GM
	assert.Equal(t, uint8(1), countGMs(*players))
	// And 6 players, ignoring the scrapper
	assert.Equal(t, uint8(6), countPlayers(*players))
	mockServer.Close()

}

// A single GM (the game creator) in the Roll20 game
func TestIncludingScrapper(t *testing.T) {
	mockServer := SetupTestServer("../../assets/sample_campaign_page.html", "/campaigns/details/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, &Options{IgnoreSelf: false})
	assert.Nil(t, err)
	var players *[]Player
	players, err = scrapper.GetPlayers("")
	assert.Nil(t, err)
	// A single GM
	assert.Equal(t, uint8(1), countGMs(*players))
	// And 7 players, including the scrapper
	assert.Equal(t, uint8(7), countPlayers(*players))
	mockServer.Close()

}

// Multiple GMs (the game creator + some players with GM perms) in the Roll20 Game
func TestMultipleGMs(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_page_multiple_gms.html", "/campaigns/details/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	var players *[]Player
	players, err = scrapper.GetPlayers("")
	assert.Nil(t, err)

	// Twos GM
	assert.Equal(t, uint8(2), countGMs(*players))
	// And 8 players, ignoring the scrapper
	assert.Equal(t, uint8(8), countPlayers(*players))
	mockServer.Close()
}

// Some players couldn't be parsed for some reason. This may never happen again but has been the
// case at least once
func TestMissingPlayers(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_missing_id.html", "/campaigns/details/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
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
	mockServer := SetupTestServer("./../../assets/sample_campaign_missing_gm.html", "/campaigns/details/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	players, err := scrapper.GetPlayers("")
	assert.Error(t, err)
	assert.Nil(t, players)
	mockServer.Close()
}

func TestMissingAvatar(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_no_gm_avatar.html", "/campaigns/details/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	players, err := scrapper.GetPlayers("")
	// A single GM
	assert.Equal(t, uint8(1), countGMs(*players))
	// And 6 players
	assert.Equal(t, uint8(6), countPlayers(*players))
	mockServer.Close()
}

// Simply joining a game as a player
func TestJoinGame(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_page_multiple_gms.html", "/campaigns/details/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	err = scrapper.JoinGame("", "")
	assert.Nil(t, err)
	mockServer.Close()
}

// Simply joining a game as a player
func TestJoinWrongURL(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_page_multiple_gms.html", "/campaigns/details/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	err = scrapper.JoinGame("%DSLSDLSMM", "##MMMD%%")
	assert.Error(t, err)
	mockServer.Close()
}

// Simply joining a game as a player
func TestFailingJoiningGame(t *testing.T) {
	// Login with a good server
	mockServer := SetupConstantServer(200)
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
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
	_, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Error(t, err)
	mockServer.Close()
}

// Corrupted DOM / Roll20 update
func TestInvalidDocument(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/non_existant.html", "/campaigns/details/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	var players *[]Player
	players, err = scrapper.GetPlayers("")
	assert.Error(t, err)
	assert.Nil(t, players)
	mockServer.Close()
}

// Corrupted DOM / Roll20 update
func TestGetMessagesOfValidPage(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_chat_archive.html", "/campaigns/chatarchive/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	var messages []Message
	err = scrapper.getMessagesOfPage("", 1, &messages)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(messages))

	println(messages)
	mockServer.Close()
}

// This page doesn't exists
func TestGetMessagesOfInvalidPage(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_chat_archive.html", "/campaigns/chatarchive/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	var messages []Message
	err = scrapper.getMessagesOfPage("", 100, &messages)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(messages))
	println(messages)
	mockServer.Close()
}

func TestGetDomInternetDown(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_chat_archive.html", "/campaigns/chatarchive/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	mockServer.Close()
	dom, err := scrapper.getDomOfRoute("/nevermind")
	println(err.Error())
	assert.Nil(t, dom)
	assert.Error(t, err)
	mockServer.Close()
}

func TestGetDomEmptyBody(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_chat_archive.html", "/campaigns/chatarchive/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	mockServer.Close()
	// Switching server for one that will no answer with a body
	mockServer = SetupConstantServer(204)
	scrapper.baseUrl = mockServer.URL
	dom, err := scrapper.getDomOfRoute("/campaigns/chatarchive/")
	println(err.Error())
	assert.Nil(t, dom)
	assert.Error(t, err)
	mockServer.Close()
}

// Get messages with a set limit
func TestGetMessagesWithLimit(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_chat_archive.html", "/campaigns/chatarchive/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	messages, err := scrapper.GetMessages("", 3, nil)
	fmt.Printf("%+v", *messages)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(*messages))

	mockServer.Close()
}

// Get messages with a set limit
func TestGetMessagesWithLimit0(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_chat_archive.html", "/campaigns/chatarchive/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	messages, err := scrapper.GetMessages("", 0, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(*messages))

	mockServer.Close()
}

// Ignoring whisper is the default
func TestGetMessagesIgnoreWhispers(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_chat_archive.html", "/campaigns/chatarchive/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	options := NewMessageOptions()
	// Including all message
	options.IncludeWhispers = false
	messages, err := scrapper.GetMessages("", ^uint(0), options)
	assert.Nil(t, err)
	// The sample as 70 messages per page. 5 of them are whispers. They are 3 virtual page. Total rolls should be 195
	assert.Equal(t, 3*(70-5), len(*messages))
	mockServer.Close()
}

// Filtering to have only rolls
func TestGetMessagesOnlyRolls(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_chat_archive.html", "/campaigns/chatarchive/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	options := NewMessageOptions()
	// Including all message
	options.IncludeWhispers = false
	options.IncludeChat = false
	messages, err := scrapper.GetMessages("", ^uint(0), options)
	assert.Nil(t, err)
	// The sample as 70 messages per page. 19 of them are whispers. They are 3 virtual page. Total rolls should be 57
	assert.Equal(t, 3*19, len(*messages))
	mockServer.Close()
}

// Combining options
func TestGetMessagesWithLimitAndFilter(t *testing.T) {
	mockServer := SetupTestServer("./../../assets/sample_campaign_chat_archive.html", "/campaigns/chatarchive/")
	scrapper, err := NewScrapper(os.Getenv("ROLL20_BASE_URL"), &Roll20Account{Login: "_", Password: "_"}, nil)
	assert.Nil(t, err)
	options := NewMessageOptions()
	// Including all message
	options.IncludeWhispers = false
	options.IncludeChat = false
	messages, err := scrapper.GetMessages("", 21, options)
	assert.Nil(t, err)
	// The sample as 70 messages per page. 19 of them are whispers. They are 3 virtual page. With a limit of 21,
	assert.Equal(t, 21, len(*messages))

	mockServer.Close()
}
