//go:build integration
// +build integration

package function

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	handler2 "github.com/openfaas/templates-sdk/go-http"
	"github.com/stretchr/testify/assert"
	"handler/function/pkg/scrapper"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"testing"
)

const projectDirName = "roll20-scrapper"

func LoadEnv(t *testing.T) {
	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))
	err := godotenv.Load(string(rootPath) + `/.env.yaml`)
	if err != nil {
		log.Printf(err.Error())
		t.SkipNow()
	}
}

// Actual get player from an existing campaign
func TestGetPlayers(t *testing.T) {
	LoadEnv(t)
	game_id, err := strconv.Atoi(os.Getenv("TESTING_CAMPAIGN_ID"))
	if err != nil {
		t.Fatalf("testing campaign id invalid -> %s", os.Getenv("TESTING_CAMPAIGN_ID"))
	}
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: fmt.Sprintf("gameId=%d", game_id),
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusMultiStatus {
		fmt.Println("Could not list roll20 game players")
		t.FailNow()
	}
	var players []scrapper.Player
	json.Unmarshal(res.Body, &players)
	// Searching for at least a single GM
	hasGm := false
	for _, p := range players {
		hasGm = hasGm || p.IsGm
	}
	if !hasGm {
		fmt.Printf("No Gms were found for game %d, this is inconsistent\n", game_id)
		t.FailNow()
	}
	fmt.Printf("%s", res.Body)

}

func TestGetPlayersOnNotJoignedGame(t *testing.T) {
	// This game shouldn't exists
	gameId := 99999
	LoadEnv(t)
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: fmt.Sprintf("gameId=%d", gameId),
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Error(t, err)
}
