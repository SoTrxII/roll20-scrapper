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
		if strings.Contains(r.URL.Path, "/campaigns/chatarchive/") {
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

// Get all messages, ignoring whispers and without a limit
func TestDefaultValues(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_chat_archive.html")
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
	var messages []scrapper.Message
	err = json.Unmarshal(res.Body, &messages)
	assert.Nil(t, err)
	assert.Len(t, messages, 195)
	mockServer.Close()

}
func TestWithLimit(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_chat_archive.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1&limit=3",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	var messages []scrapper.Message
	err = json.Unmarshal(res.Body, &messages)
	assert.Nil(t, err)
	assert.Len(t, messages, 3)
	mockServer.Close()

}

func TestWithInvalidLimit(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_chat_archive.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1&limit=-3",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	var messages []scrapper.Message
	err = json.Unmarshal(res.Body, &messages)
	assert.Error(t, err)
	mockServer.Close()

}

func TestExcludingAll(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_chat_archive.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1&includeChats=false&includeWhispers=false&includeRolls=false",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	var messages []scrapper.Message
	err = json.Unmarshal(res.Body, &messages)
	assert.Len(t, messages, 0)
	mockServer.Close()

}
func TestIncludingAll(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_chat_archive.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1&includeChats=true&includeWhispers=true&includeRolls=true",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	var messages []scrapper.Message
	err = json.Unmarshal(res.Body, &messages)
	assert.Len(t, messages, 210)
	mockServer.Close()

}

func TestWrongOptionsChats(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_chat_archive.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1&includeChats=dd",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	mockServer.Close()

}
func TestWrongOptionsWhispers(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_chat_archive.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1&includeWhispers=dd",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	mockServer.Close()

}
func TestWrongOptionsRolls(t *testing.T) {
	mockServer := SetupTestServer("assets/sample_campaign_chat_archive.html")
	req := handler2.Request{
		Body:        nil,
		Header:      nil,
		QueryString: "gameId=1&includeRolls=dd",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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
