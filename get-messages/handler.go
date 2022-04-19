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
const LIMIT_URL_NAME = "limit"
const WHISPER_URL_NAME = "includeWhispers"
const ROLLS_URL_NAME = "includeRolls"
const CHAT_URL_NAME = "includeChats"

// swagger:route GET /get-messages Players get-messages
//
// Retrieve all messages for a specific roll20 game.
//
// The player can either be GMs or not. There can be multiple GMs in a single game
//     Produces:
//     - application/json
//     Parameters:
//       + name: gameId
//         in: query
//         description: Roll20 ID of the game to parse. For link https://app.roll20.net/join/1/59lzQg --> Game ID is "1"
//         required: true
//         type: integer
//         format: int32
//       + name: limit
//         in: query
//         description: Max number of messages to parse. Default is all available
//         required: false
//         type: integer
//         format: uint
//       + name: includeWhispers
//         in: query
//         description: Include whispers in messages. Default is false
//         required: false
//         type: boolean
//       + name: includeRolls
//         in: query
//         description: Include rolls in messages. Default is true
//         required: false
//         type: boolean
//       + name: includeChat
//         in: query
//         description: Include general chat messages. Default is true
//         required: false
//         type: boolean
// responses:
//  200: []Message Complete list of players for the requested game
//	400: ErrorTemplate Missing or invalid QS provided
//  500: ErrorTemplate Configuration error, either env variables missing or provided roll20 credentials invalid
func Handle(req handler2.Request) (handler2.Response, error) {
	log.Println("Get messages handler has been woken up")
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

	// Parse limit. This is an optional argument, default is UINT_MAX
	limit := ^uint(0)
	if qs.Has(LIMIT_URL_NAME) {
		limitQs := qs.Get(LIMIT_URL_NAME)
		// Cannot parse directly to an uint
		var limitInt int
		if limitInt, err = strconv.Atoi(limitQs); len(limitQs) == 0 || err != nil || limitInt < 0 {
			log.Printf("Wrong limit provided: %d. Error : %s \n", limitInt, err)
			errMessage := fmt.Sprintf("The provided limit is invalid %s. It should be a postive integer.\n", gameId)
			return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError(errMessage)}, err
		}
		limit = uint(limitInt)
	}

	// options
	opt := scrapper.NewMessageOptions()
	if qs.Has(WHISPER_URL_NAME) {
		value := qs.Get(WHISPER_URL_NAME)
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			log.Printf("Wrong value provided: %s. Should be eitehr true or false \n", value)
			errMessage := fmt.Sprintf("Wrong value provided: %s. Should be eitehr true or false \n", value)
			return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError(errMessage)}, err
		}
		opt.IncludeWhispers = enabled
	}
	if qs.Has(CHAT_URL_NAME) {
		value := qs.Get(CHAT_URL_NAME)
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			log.Printf("Wrong value provided: %s. Should be eitehr true or false \n", value)
			errMessage := fmt.Sprintf("Wrong value provided: %s. Should be eitehr true or false \n", value)
			return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError(errMessage)}, err
		}
		opt.IncludeChat = enabled
	}
	if qs.Has(ROLLS_URL_NAME) {
		value := qs.Get(ROLLS_URL_NAME)
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			log.Printf("Wrong value provided: %s. Should be eitehr true or false \n", value)
			errMessage := fmt.Sprintf("Wrong value provided: %s. Should be eitehr true or false \n", value)
			return handler2.Response{StatusCode: http.StatusBadRequest, Body: http_helpers.FormatError(errMessage)}, err
		}
		opt.IncludeRolls = enabled
	}

	log.Println("Now fetching messages for campaign " + gameId)

	// Scrap the messages from the game
	s, err := scrapper.NewScrapper(values["ROLL20_BASE_URL"], &scrapper.Roll20Account{Login: values["ROLL20_USERNAME"], Password: values["ROLL20_PASSWORD"]}, nil)
	if err != nil {
		log.Printf("The scrapper instance couldn't be initialized. Error %s\n", err)
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: http_helpers.FormatError("Unexpected error")}, err
	}
	messages, err := s.GetMessages(gameId, limit, opt)
	if err != nil {
		log.Printf("Unexpected error : %s\n", err.Error())
		return handler2.Response{StatusCode: http.StatusInternalServerError, Body: http_helpers.FormatError(err.Error())}, err
	}
	log.Println("All messages have been successfully scrapped from campaign " + gameId)
	// If all messages have been picked up, send them back with a 200
	messagesJson, err := json.Marshal(messages)
	return handler2.Response{
		StatusCode: http.StatusOK,
		Body:       messagesJson,
		Header: map[string][]string{
			"Content-type": {"application/json"},
		},
	}, err
}
