package function

import (
	"fmt"
	handler2 "github.com/openfaas/templates-sdk/go-http"
	config_parser "handler/function/pkg/config-parser"
	"handler/function/pkg/scrapper"
	"net/http"
	"net/url"
)

const QS_ID_URL_NAME = "gameId"
const QS_CODE_URL_NAME = "gameCode"

// Handle a function invocation
func Handle(req handler2.Request) (handler2.Response, error) {
	var err error
	// Retrieve runtime values from env & QS
	var keys = []string{"ROLL20_USERNAME", "ROLL20_PASSWORD", "ROLL20_BASE_URL"}
	values, err := config_parser.ParseEnv(keys)
	if err != nil {
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: []byte("Unexpected error while parsing env")}, err
	}
	qs, err := url.ParseQuery(req.QueryString)
	if err != nil {
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: []byte("Unexpected error while parsing qs")}, err
	}
	gameId := qs.Get(QS_ID_URL_NAME)
	if len(gameId) == 0 {
		err = fmt.Errorf("The provided gameId is invalid %s\n", gameId)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: []byte(err.Error())}, nil
	}
	gameCode := qs.Get(QS_CODE_URL_NAME)
	if len(gameCode) == 0 {
		err = fmt.Errorf("The provided gameCode is invalid %s\n", gameCode)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: []byte(err.Error())}, nil
	}

	// Join the roll20 game
	s, err := scrapper.NewScrapper(values["ROLL20_BASE_URL"], &scrapper.Roll20Account{Login: values["ROLL20_USERNAME"], Password: values["ROLL20_PASSWORD"]})
	if err != nil {
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: []byte("Unexpected error")}, err
	}
	err = s.JoinGame(gameId, gameCode)
	if err != nil {
		errMessage := fmt.Sprintf("Couldn't join roll20 game with gameid %s and gamecode %s. Reason : %s\n", gameId, gameCode, err)
		return handler2.Response{StatusCode: http.StatusBadRequest, Body: []byte(errMessage)}, err
	}

	return handler2.Response{
		StatusCode: http.StatusNoContent,
	}, err
}
