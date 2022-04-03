package function

import (
	"encoding/json"
	"fmt"
	handler2 "github.com/openfaas/templates-sdk/go-http"
	"github.com/stretchr/testify/assert"
	"handler/function/pkg/scrapper"
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
	var sample *os.File
	// Open provided path
	sample, err := os.Open(path.Join(dir, campaignDataPath))
	// On CI, the path may be wrong because the import path is different
	if err != nil {
		sample, err = os.Open(path.Join(dir, "../", campaignDataPath))
	}
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

// Server only allowing the scrapper to log in
func SetupLoginOnlyServer() *httptest.Server {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/campaigns/details/") {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	}))
	os.Setenv("ROLL20_BASE_URL", mockServer.URL)
	os.Setenv("ROLL20_USERNAME", "mock")
	os.Setenv("ROLL20_PASSWORD", "mock")
	return mockServer
}

//wrong user provider argument
func TestWrongGameID(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_page.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=sss",
		Method:      "GET",
		Host:        "",
	}
	_, err := Handle(req)
	assert.Error(t, err)
	mockServer.Close()

}

//No user provided argument
func TestMissingGameID(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_page.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "r=dd",
		Method:      "GET",
		Host:        "",
	}
	_, err := Handle(req)
	assert.Error(t, err)
	mockServer.Close()

}

// Not all players were parsed
func TestIncompleteAnswer(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_missing_id.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusMultiStatus, res.StatusCode)
	var players []scrapper.Player
	err = json.Unmarshal(res.Body, &players)
	assert.Nil(t, err)
	assert.Len(t, players, 6)
	mockServer.Close()

}

// All players were parsed
func TestCompleteAnswer(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_page.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	var players []scrapper.Player
	err = json.Unmarshal(res.Body, &players)
	assert.Nil(t, err)
	assert.Len(t, players, 8)
	mockServer.Close()

}

// No env variables defined
func TestNoEnv(t *testing.T) {
	os.Unsetenv("ROLL20_BASE_URL")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1",
		Method:      "GET",
		Host:        "",
	}
	res, _ := Handle(req)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	fmt.Println(string(res.Body))

}

// QS missing
func TestNoQS(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_missing_id.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "",
		Method:      "GET",
		Host:        "",
	}
	res, _ := Handle(req)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	fmt.Println(string(res.Body))
	mockServer.Close()
}

// QS is somehow wrong. Fuzz attack ?
func TestInvalidQS(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_missing_id.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "wrong=;;;",
		Method:      "GET",
		Host:        "",
	}
	res, _ := Handle(req)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	fmt.Println(string(res.Body))
	mockServer.Close()
}

// End gracefully when the scrapper itself fails
func TestScrapperLoginError(t *testing.T) {
	// env defined but wrong, the scrapper won't be able to login
	os.Setenv("ROLL20_BASE_URL", "wrong")
	os.Setenv("ROLL20_USERNAME", "mock")
	os.Setenv("ROLL20_PASSWORD", "mock")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	fmt.Println(string(res.Body))

}

// End gracefully when the scrapper itself fails
func TestScrappingError(t *testing.T) {
	mockServer := SetupLoginOnlyServer()
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	fmt.Println(string(res.Body))
	mockServer.Close()

}
