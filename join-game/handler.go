package function

import (
	"fmt"
	handler2 "github.com/openfaas/templates-sdk/go-http"
	config_parser "handler/function/pkg/config-parser"
	http_helpers "handler/function/pkg/http-helpers"
	"handler/function/pkg/scrapper"
	"net/http"
	"net/url"
)

const QS_ID_URL_NAME = "gameId"
const QS_CODE_URL_NAME = "gameCode"

// swagger:route GET /join-game Players join-game
//
// Makes the bot account join the game as a player
//
// This is a mandatory step for every other request, as the bot account won't have access to a game before joining it
//     Produces:
//     - application/json
//     Parameters:
//       + name: gameId
//         in: query
//         description: Roll20 ID of the game to parse. For link https://app.roll20.net/join/1/59lzQg --> Game ID is "1"
//         required: true
//		   example: 1
//         type: integer
//         format: int32
//      + name: gameCode
//         in: query
//         description: Roll20 Code to join the game. This is usually the last part of the join link. For link https://app.roll20.net/join/1/59lzQg --> Join code is "59lzQg"
//		   example: "59lzQg"
//         required: true
//         type: string
// responses:
//  204: description: Game successfully joined
//	400: ErrorTemplate Missing or invalid game ID or gameCode provided
//  500: ErrorTemplate Configuration error, either env variables missing or provided roll20 credentials invalid
func Handle(req handler2.Request) (handler2.Response, error) {
	var err error
	// Retrieve runtime values from env & QS
	var keys = []string{"ROLL20_USERNAME", "ROLL20_PASSWORD", "ROLL20_BASE_URL"}
	values, err := config_parser.ParseEnv(keys)
	if err != nil {
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: http_helpers.FormatError("Unexpected error while parsing env")}, err
	}
	qs, err := url.ParseQuery(req.QueryString)
	if err != nil {
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError("Unexpected error while parsing qs")}, err
	}
	gameId := qs.Get(QS_ID_URL_NAME)
	if len(gameId) == 0 {
		err = fmt.Errorf("The provided gameId is invalid %s\n", gameId)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError(err.Error())}, nil
	}
	gameCode := qs.Get(QS_CODE_URL_NAME)
	if len(gameCode) == 0 {
		err = fmt.Errorf("The provided gameCode is invalid %s\n", gameCode)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError(err.Error())}, nil
	}

	// Join the roll20 game
	s, err := scrapper.NewScrapper(values["ROLL20_BASE_URL"], &scrapper.Roll20Account{Login: values["ROLL20_USERNAME"], Password: values["ROLL20_PASSWORD"]})
	if err != nil {
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: http_helpers.FormatError("Unexpected error")}, err
	}
	err = s.JoinGame(gameId, gameCode)
	if err != nil {
		errMessage := fmt.Sprintf("Couldn't join roll20 game with gameid %s and gamecode %s. Reason : %s\n", gameId, gameCode, err)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError(errMessage)}, err
	}

	return handler2.Response{
		StatusCode: http.StatusNoContent,
	}, err
}
