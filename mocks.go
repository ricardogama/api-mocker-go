package mocker

//go:generate mockgen -destination=testdata/http_mock.go -package=testdata net/http Handler
