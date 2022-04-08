package scrapper

type roll20Routes struct {
	login           string
	loginRedirect   string
	campaignDetails func(id string) string
}

func getRoutes() *roll20Routes {
	return &roll20Routes{
		login:         "/sessions/create",
		loginRedirect: "/sessions/new",
		campaignDetails: func(id string) string {
			return "/campaigns/details/" + id
		},
	}
}
