package scrapper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

type Scrapper struct {
	baseUrl string
	routes  *roll20Routes
	client  *http.Client
	account *Roll20Account
	options *Options
}

// NewScrapper Creates a new Roll20 Scrapper instance, login it in immediately
func NewScrapper(baseUrl string, account *Roll20Account, options *Options) (*Scrapper, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	if options == nil {
		options = &Options{IgnoreSelf: true}
	}
	s := &Scrapper{baseUrl: baseUrl, routes: getRoutes(), client: &http.Client{Jar: jar}, account: account, options: options}
	err = s.login()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// JoinGame Join a Roll 20 game instance given the campaign id and the joincode
func (s *Scrapper) JoinGame(gameId string, gameCode string) error {
	gameUrl, err := url.Parse(fmt.Sprintf("%s/join/%s/%s", s.baseUrl, gameId, gameCode))
	if err != nil {
		return fmt.Errorf("invalid parsed url: %s. Error info:  %s\n", gameUrl, err.Error())
	}
	res, err := s.client.Get(gameUrl.String())
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("Could not join game.: %s. Error info:  %s\n", gameUrl, err.Error())
	}
	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("Could not join game. Status :  %d. Message:  %s\n", res.StatusCode, body)
	}
	return nil
}

// GetPlayers Retrieve all players of a Roll20 game given the id of a joined campaign
func (s *Scrapper) GetPlayers(campaignId string) (*[]Player, error) {
	route := s.routes.campaignDetails(campaignId)
	doc, err := s.getDomOfRoute(route)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve the DOM of %s : %s", route, err)
	}
	if err != nil {
		return nil, err
	}
	var players []Player
	var ignoredPlayers []string

	// First, search for all players, this excludes the Game Creator.
	// Normal players can be GM
	doc.Find(".playerlisting .pclisting").Each(func(_ int, playerDom *goquery.Selection) {
		player, err := getPlayerFromDom(playerDom)
		if err != nil {
			ignoredPlayers = append(ignoredPlayers, err.Error())
			return
		}
		players = append(players, *player)
	})

	gm := Player{IsGm: true, AvatarUrl: ""}
	doc.Find(".playerlisting .profilemeta > .userprofile").Each(func(index int, gmDiv *goquery.Selection) {
		// Username
		gm.Username = strings.TrimSpace(gmDiv.Find(".name").Text())

		// Roll2O ID
		gmUrl, exists := gmDiv.Attr("href")
		if !exists {
			ignoredPlayers = append(ignoredPlayers, gm.Username)
			return
		}
		gmRoll20Id, err := getPlayerIdFromURL(gmUrl)
		if err != nil {
			ignoredPlayers = append(ignoredPlayers, gm.Username)
			return
		}
		gm.Roll20Id = gmRoll20Id
	})

	// GM's avatar has to be retrieved in another div
	doc.Find(".playerlisting .userprofile .avatar img").Each(func(index int, avatarDiv *goquery.Selection) {
		avatar, exists := avatarDiv.Attr("src")
		// We don't really mind about the avatar
		if !exists {
			return
		}
		gm.AvatarUrl = getAbsoluteUrlToImage(avatar)
	})
	// No GM defined
	if len(gm.Username) == 0 || gm.Roll20Id <= 0 {
		return nil, fmt.Errorf("No GM found for this game. Has the game been joined yet ?")
	}
	players = append(players, gm)

	// Remove the scrapper bot account from the list of scrapped players
	if s.options.IgnoreSelf {
		ownId, err := retrieveOwnRoll20ID(doc)
		if err != nil {
			return nil, fmt.Errorf("The scrapper couldn't retrieve its own ID, something not right with the DOM")
		}

		n := 0
		for _, player := range players {
			if player.Roll20Id != ownId {
				players[n] = player
				n++
			}
		}
		players = players[:n]

	}

	if len(ignoredPlayers) > 0 {
		err = &IncompleteError{
			Err: fmt.Errorf("The following players have been ignored : %s", strings.Join(ignoredPlayers, ",")),
		}
	}
	return &players, err
}

// GetSummary Retrieve a short overview of a Roll20 campaign
func (s *Scrapper) GetSummary(campaignId string) (*Summary, error) {
	route := s.routes.campaignDetails(campaignId)
	doc, err := s.getDomOfRoute(route)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve the DOM of %s : %s", route, err)
	}
	if err != nil {
		return nil, err
	}
	summary := Summary{}

	doc.Find(".campaign_details").Each(func(_ int, campaignDetailsDiv *goquery.Selection) {
		// Search for the ID
		id, exists := campaignDetailsDiv.Attr("data-campaignid")
		if !exists {
			return
		}
		idNumber, err := strconv.Atoi(id)
		if err != nil {
			return
		}
		summary.Id = idNumber

		// The name
		cName := strings.TrimSpace(campaignDetailsDiv.Find(".campaignname span").Text())
		summary.Name = cName

		// And finnaly the image
		cImg, exists := campaignDetailsDiv.Find(".campaignicon img").Attr("src")
		// We don't really mind if the image doesn't exist
		if !exists {
			return
		}
		summary.Image = cImg

	})

	return &summary, nil
}

// GetMessages Retrieve all messages from the chat
func (s *Scrapper) GetMessages(campaignId string, limit uint, options *MessageOptions) (*[]Message, error) {
	var messages []Message

	// Why ?
	if limit == 0 {
		return &messages, nil
	}

	if options == nil {
		options = NewMessageOptions()
	}

	// Fetching all pages, only stopping when we ran up of messages to parse, or if we got to the requested limit
	for currentPage, oldMessagesLen := 1, -1; uint(len(messages)) < limit && oldMessagesLen != len(messages); currentPage++ {
		oldMessagesLen = len(messages)
		var messageTemp []Message
		err := s.getMessagesOfPage(campaignId, currentPage, &messageTemp)
		if err != nil {
			return nil, fmt.Errorf("while parsing page %d : %s", currentPage, err)
		}
		// Filter message with user inputs
		for _, m := range messageTemp {
			if options.isAllowing(m) {
				messages = append(messages, m)
			}
		}

	}

	// If too many result were parsed, truncate the array
	if uint(len(messages)) > limit {
		messages = messages[0:limit]
	}

	return &messages, nil
}

// getMessagesOfPage Retrieve all the messages from a specific page
func (s *Scrapper) getMessagesOfPage(campaignId string, page int, messagesBuffer *[]Message) error {
	route := s.routes.campaignArchives(campaignId, page)
	doc, err := s.getDomOfRoute(route)
	if err != nil || doc == nil {
		return fmt.Errorf("unable to retrieve the DOM of %s : %s", route, err)
	}
	// Checking if we requested a non-existing page
	// Something like "Page 1/100"
	pageDiv := doc.Find(".pagination div")
	if pageDiv == nil {
		return fmt.Errorf("Unable to locate pagination div. Halting")
	}
	pageText := pageDiv.Text()
	if !strings.HasPrefix(pageText, "Page") {
		return fmt.Errorf("Unable to locate pagination text, got %s", pageText)
	}
	offsets := strings.Split(strings.TrimSpace(strings.Replace(pageText, "Page", "", 1)), "/")
	pageUpperLimit, err := strconv.Atoi(offsets[1])
	if err != nil {
		return fmt.Errorf("Unable to locate pagination text, got %s", pageText)
	}
	if page > pageUpperLimit {
		return nil
	}
	// Page is valid, let's parse

	// Searching for the base54 encoded data array
	const PREFIX = "var msgdata ="
	// It's contained in a script tag
	allScripts := doc.Find("script")
	// Reverse the matches, the script containing msgData should be near the end
	for i, j := 0, len(allScripts.Nodes)-1; i < j; i, j = i+1, j-1 {
		allScripts.Nodes[i], allScripts.Nodes[j] = allScripts.Nodes[j], allScripts.Nodes[i]
	}
	var chosenScript *goquery.Selection
	allScripts.EachWithBreak(func(i int, script *goquery.Selection) bool {
		text := strings.TrimSpace(script.Text())
		if strings.HasPrefix(text, PREFIX) {
			chosenScript = script
			// Break the loop
			return false
		}
		return true
	})
	if chosenScript == nil {
		return fmt.Errorf("Couldn't retrieve the msgdata variable")
	}
	msgScript := chosenScript.Text()
	// Remove variable declaration prefix
	msgScript = strings.TrimSpace(strings.Replace(msgScript, PREFIX, "", 1))
	const SUFFIX = "\";\nO"
	lastIndex := strings.LastIndex(msgScript, SUFFIX)
	if lastIndex == -1 {
		return fmt.Errorf("msgdata variable isn't well formatted")
	}
	// Remove isolate the value of the msgData variable
	// len(suffix) -1 is to keep the "==" but not the trailing '"'
	msgScript = msgScript[1:lastIndex]

	chatMessages, err := base64.StdEncoding.DecodeString(msgScript)
	if err != nil {
		return err
	}
	// Spatial complexity is at least 2N, N < 100 messages
	// Raw JSON struct as returned by roll20
	var mappedMessages []map[string]Message
	// Actual isolated messages
	err = json.Unmarshal(chatMessages, &mappedMessages)
	if err != nil {
		return err
	}

	for _, v := range mappedMessages[0] {
		*messagesBuffer = append(*messagesBuffer, v)
	}

	return nil
}

// Given a Roll20 relative url, retrieve the DOm as a goquery document
func (s *Scrapper) getDomOfRoute(path string) (*goquery.Document, error) {
	campaignArchivesUrl, err := url.Parse(s.baseUrl + path)
	if err != nil {
		return nil, err
	}
	res, err := s.client.Get(campaignArchivesUrl.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 || res.ContentLength == 0 {
		return nil, fmt.Errorf("invalid response received. Status is %d, Content length is %d", res.StatusCode, res.ContentLength)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// Log in to roll20 using the defined account
// This will initialize the cookies needed to make an authorized request to roll20
func (s *Scrapper) login() error {
	loginUrl, err := url.Parse(s.baseUrl + s.routes.login)
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Set("email", s.account.Login)
	params.Set("password", s.account.Password)
	r, err := http.NewRequest("POST", loginUrl.String(), strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Origin", strings.TrimSuffix(s.baseUrl, "/"))
	res, err := s.client.Do(r)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(res.Body)
		fmt.Printf(string(b))
		return fmt.Errorf("invalid credentials provided. u : %s, p: %s", s.account.Login, s.account.Password)
	}
	return nil
}

// Fetch the bot account own ID, this is to allow the bot to ignore itself on the players fetching
func retrieveOwnRoll20ID(doc *goquery.Document) (int, error) {
	href, exist := doc.Find(".topbarlogin .simple a[href*=\"wishlists\"]").First().Attr("href")
	if !exist {
		return -1, fmt.Errorf("The scrapper couldn't retrieve its own ID, something not right with the DOM")
	}
	hrefArray := strings.Split(strings.TrimSpace(href), "/")
	ownId, err := strconv.Atoi(hrefArray[len(hrefArray)-1])
	if err != nil {
		return -1, err
	}
	return ownId, nil
}

// From a player division, parse a Player object.
// In cases in which some vital info could not be retrieved (such as roll20ID)
// Ony the username of the player is returned as an error
func getPlayerFromDom(playerDom *goquery.Selection) (*Player, error) {
	avatarUrl := ""
	avatar, exists := playerDom.Find(".circleavatar").Attr("src")
	// If the avatar url exists
	if exists {
		avatarUrl = getAbsoluteUrlToImage(avatar)
	}
	// Retrieve playerDom username from the dom own text
	// Sometime, (GM) is added to the playerDom name, so we have to get rid of it
	username := ""
	if usernameArray := strings.Split(strings.TrimSpace(playerDom.Text()), "\n"); len(usernameArray) > 0 {
		username = usernameArray[0]
	}
	// ID is mandatory, if we can't retrieve it, we have to ignore the playerDom
	playerProfileUrl, exists := playerDom.Find("a").Attr("href")
	if !exists {
		return nil, fmt.Errorf(username)
	}
	playerRoll20Id, err := getPlayerIdFromURL(playerProfileUrl)
	if err != nil {
		return nil, fmt.Errorf(username)
	}
	return &Player{
		AvatarUrl: avatarUrl,
		IsGm:      playerDom.Find(".gmbadge").Length() > 0,
		Roll20Id:  playerRoll20Id,
		Username:  username,
	}, nil
}
