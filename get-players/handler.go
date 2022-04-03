package function

import (
	"encoding/json"
	"fmt"
	handler2 "github.com/openfaas/templates-sdk/go-http"
	config_parser "handler/function/pkg/config-parser"
	"handler/function/pkg/scrapper"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const QS_GAME_URL_NAME = "gameId"

// swagger:route GET /get-players Thing get-thing
//
// This is the summary for getting a thing by its UUID
//
// This is the description for getting a thing by its UUID. Which can be longer.
//
// responses:
//   200: Players
//   404: ErrorResponse
//   500: ErrorResponse
func Handle(req handler2.Request) (handler2.Response, error) {
	log.Println("Get players handler has been woken up")
	var err error
	// Retrieve runtime values from env & QS
	var keys = []string{"ROLL20_USERNAME", "ROLL20_PASSWORD", "ROLL20_BASE_URL"}
	values, err := config_parser.ParseEnv(keys)
	if err != nil {
		log.Printf("Invalid env : %s\n", err)
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: []byte("Unexpected error while parsing env")}, err
	}
	qs, err := url.ParseQuery(req.QueryString)
	if err != nil {
		log.Printf("Invalid QS : %s. Error : %s \n", qs, err)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: []byte("Unexpected error while parsing qs")}, err
	}
	gameId := qs.Get(QS_GAME_URL_NAME)
	if _, err = strconv.Atoi(gameId); len(gameId) == 0 || err != nil {
		log.Printf("Wrong gameid provided: %s. Error : %s \n", gameId, err)
		errMessage := fmt.Sprintf("The provided gameId is invalid %s\n", gameId)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: []byte(errMessage)}, err
	}
	log.Println("Now fetching players for campaign " + gameId)

	// Scrap the players from the game
	s, err := scrapper.NewScrapper(values["ROLL20_BASE_URL"], &scrapper.Roll20Account{Login: values["ROLL20_USERNAME"], Password: values["ROLL20_PASSWORD"]})
	if err != nil {
		log.Printf("The scrapper instance couldn't be initialized. Error %s\n", err)
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: []byte("Unexpected error")}, err
	}
	players, err := s.GetPlayers(gameId)
	if err != nil {
		// If the scrapper did not succeed with all the players, indicate it
		re, ok := err.(*scrapper.IncompleteError)
		if ok {
			playersJson, err := json.Marshal(players)
			if err != nil {
				return handler2.Response{Body: []byte("Unexpected error")}, err
			}
			log.Println(re.Error())
			return handler2.Response{
				StatusCode: http.StatusMultiStatus,
				Body:       playersJson,
			}, nil
		}
		log.Printf("Unexpected error : %s\n", err.Error())
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: []byte(err.Error())}, err
	}
	log.Println("All players have been successfully scrapped from campaign " + gameId)
	// If all players have been picked up, send them back with a 200
	playersJson, err := json.Marshal(players)
	return handler2.Response{
		StatusCode: http.StatusOK,
		Body:       playersJson,
		Header: map[string][]string{
			"Content-type": []string{"application/json"},
		},
	}, err
}
