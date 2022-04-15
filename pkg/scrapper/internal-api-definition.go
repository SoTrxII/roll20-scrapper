package scrapper

import "fmt"

type roll20Routes struct {
	login            string
	loginRedirect    string
	campaignDetails  func(id string) string
	campaignArchives func(id string, page int) string
}

func getRoutes() *roll20Routes {
	return &roll20Routes{
		login:         "/sessions/create",
		loginRedirect: "/sessions/new",
		campaignDetails: func(id string) string {
			return "/campaigns/details/" + id
		},
		campaignArchives: func(id string, page int) string {
			return fmt.Sprintf("/campaigns/chatarchive/%s?p=%d&hiderollresults=true", id, page)
		},
	}
}
