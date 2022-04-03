package function

import (
	"fmt"
	handler2 "github.com/openfaas/templates-sdk/go-http"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

func SetupTestServer(campaignDataPath string) *httptest.Server {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Dir(filename)
	sample, _ := os.Open(path.Join(dir, campaignDataPath))
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
		if strings.Contains(r.URL.Path, "/join") {
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

//No user provided argument
func TestQsParams(t *testing.T) {
	mockServer := SetupTestServer("../assets/sample_campaign_page.html")
	var tests = []struct {
		gameId, gameCode string
		statusCode       int
	}{
		{"", "", http.StatusBadRequest},
		{"gg", "", http.StatusBadRequest},
		{"", "gg", http.StatusBadRequest},
		{"30000", "dsdsds", http.StatusNoContent},
	}

	for _, tt := range tests {
		// t.Run enables running "subtests", one for each
		// table entry. These are shown separately
		// when executing `go test -v`.
		testname := fmt.Sprintf("\n Values : %s,%s\n", tt.gameId, tt.gameCode)
		qs, _ := url.ParseQuery("")
		if len(tt.gameId) != 0 {
			qs.Add("gameId", tt.gameId)
		}
		if len(tt.gameCode) != 0 {
			qs.Add("gameCode", tt.gameCode)
		}

		t.Run(testname, func(t *testing.T) {
			req := handler2.Request{
				Body:        nil,
				Header:      nil,
				QueryString: qs.Encode(),
				Method:      "GET",
				Host:        "",
			}
			res, _ := Handle(req)
			if res.StatusCode != tt.statusCode {
				t.Errorf("got %d, want %d", res.StatusCode, tt.statusCode)
			}
		})
	}
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

// QS is somehow wrong. Fuzz attack ?
func TestInvalidQS(t *testing.T) {
	mockServer := SetupTestServer("../assets/sample_campaign_missing_id.html")
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
		QueryString: "gameId=1&gameCode=1",
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
		QueryString: "gameId=1&gameCode=1",
		Method:      "GET",
		Host:        "",
	}
	res, err := Handle(req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	fmt.Println(string(res.Body))
	mockServer.Close()

}
