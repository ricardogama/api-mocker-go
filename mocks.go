package mocker

//go:generate mockgen -destination=mocks/http.go -package=mocks  net/http Handler
