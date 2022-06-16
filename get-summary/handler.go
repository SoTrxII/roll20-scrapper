package function

import (
	"encoding/json"
	"fmt"
	handler2 "github.com/openfaas/templates-sdk/go-http"
	config_parser "handler/function/pkg/config-parser"
	http_helpers "handler/function/pkg/http-helpers"
	"handler/function/pkg/scrapper"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const QS_GAME_URL_NAME = "gameId"

// swagger:route GET /get-summary Summary get-summary
//
// Retrieve basic info about a roll20 campaign
//
//     Produces:
//     - application/json
//     Parameters:
//       + name: gameId
//         in: query
//         description: Roll20 ID of the game to parse. For link https://app.roll20.net/join/1/59lzQg --> Game ID is "1"
//         required: true
//         type: integer
//         format: int32
// responses:
//  200: []Player Complete list of players for the requested game
//	400: ErrorTemplate Missing or invalid game ID provided
//  500: ErrorTemplate Configuration error, either env variables missing or provided roll20 credentials invalid
func Handle(req handler2.Request) (handler2.Response, error) {
	log.Println("Get summary handler has been woken up")
	var err error
	// Retrieve runtime values from env & QS
	var keys = []string{"ROLL20_USERNAME", "ROLL20_PASSWORD", "ROLL20_BASE_URL"}
	values, err := config_parser.ParseEnv(keys)
	if err != nil {
		log.Printf("Invalid env : %s\n", err)
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: http_helpers.FormatError("Unexpected error while parsing env")}, err
	}
	qs, err := url.ParseQuery(req.QueryString)
	if err != nil {
		log.Printf("Invalid QS : %s. Error : %s \n", qs, err)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError("Unexpected error while parsing qs")}, err
	}
	gameId := qs.Get(QS_GAME_URL_NAME)
	if _, err = strconv.Atoi(gameId); len(gameId) == 0 || err != nil {
		log.Printf("Wrong gameid provided: %s. Error : %s \n", gameId, err)
		errMessage := fmt.Sprintf("The provided gameId is invalid %s\n", gameId)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError(errMessage)}, err
	}
	log.Println("Now fetching summary for campaign " + gameId)

	// Scrap the players from the game
	s, err := scrapper.NewScrapper(values["ROLL20_BASE_URL"], &scrapper.Roll20Account{Login: values["ROLL20_USERNAME"], Password: values["ROLL20_PASSWORD"]}, nil)
	if err != nil {
		log.Printf("The scrapper instance couldn't be initialized. Error %s\n", err)
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: http_helpers.FormatError("Unexpected error")}, err
	}
	summary, err := s.GetSummary(gameId)
	if err != nil {
		log.Printf("Unexpected error : %s\n", err.Error())
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: http_helpers.FormatError(err.Error())}, err
	}
	log.Println("Summary have been successfully scrapped from campaign " + gameId)
	// If all players have been picked up, send them back with a 200
	summaryJson, err := json.Marshal(summary)
	return handler2.Response{
		StatusCode: http.StatusOK,
		Body:       summaryJson,
		Header: map[string][]string{
			"Content-type": {"application/json"},
		},
	}, err
}
