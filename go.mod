module roll20-scrapper

go 1.17

require (
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/joho/godotenv v1.4.0
	github.com/openfaas/templates-sdk v0.0.0-20200723110415-a699ec277c12
	github.com/stretchr/testify v1.7.1
	handler/function v0.0.0-00010101000000-000000000000
)

require (
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace handler/function => ./

replace github.com/ => ./vendor
