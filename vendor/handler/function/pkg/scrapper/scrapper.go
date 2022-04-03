package scrapper

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type Scrapper struct {
	baseUrl string
	routes  *roll20Routes
	client  *http.Client
	account *Roll20Account
}

// Player A Roll20 Player as listed on the campaign page
type Player struct {
	// Player avatar URL
	AvatarUrl string
	// Is the player a GM of the parsed game.
	// There are multiple GM per games
	IsGm bool
	// Roll20 Id of the player
	Roll20Id int
	// Username (not character name)
	Username string
}

// A Roll20Account is a basic creds user in roll20
//
// As there is so such thing as a service account in Roll20,
// this has to be a real account
type Roll20Account struct {
	Login    string
	Password string
}

type IncompleteError struct {
	Err error
}

func (r *IncompleteError) Error() string {
	return r.Err.Error()
}

// NewScrapper Creates a new Roll20 Scrapper instance, login it in immediately
func NewScrapper(baseUrl string, account *Roll20Account) (*Scrapper, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	s := &Scrapper{baseUrl: baseUrl, routes: getRoutes(), client: &http.Client{Jar: jar}, account: account}
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

func (s *Scrapper) GetPlayers(campaignId string) (*[]Player, error) {
	campaignDomUrl, err := url.Parse(s.baseUrl + s.routes.campaignDetails(campaignId))
	if err != nil {
		return nil, err
	}
	res, err := s.client.Get(campaignDomUrl.String())
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
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

	if len(ignoredPlayers) > 0 {
		err = &IncompleteError{
			Err: fmt.Errorf("The following players have been ignored : %s", strings.Join(ignoredPlayers, ",")),
		}
	}
	return &players, err
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
