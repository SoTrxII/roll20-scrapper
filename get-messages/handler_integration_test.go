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

// Actual get messages from an existing campaign
func TestGetMessages(t *testing.T) {
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
	if res.StatusCode != http.StatusOK {
		fmt.Println("Could not list roll20 game players")
		t.FailNow()
	}
	var messages []scrapper.Message
	json.Unmarshal(res.Body, &messages)
	fmt.Printf("%+v\n", messages)
	fmt.Printf("%d\n", len(messages))

}

func TestGetMessageOnNotJoignedGame(t *testing.T) {
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
