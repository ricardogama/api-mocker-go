# api-mocker-go

Go SDK for [api-mocker](https://github.com/ricardogama/api-mocker)

## Status

[![Build Status](https://travis-ci.com/ricardogama/api-mocker-go.svg?branch=master)](https://travis-ci.com/ricardogama/api-mocker-go) [![Coverage Status](https://coveralls.io/repos/github/ricardogama/api-mocker-go/badge.svg?branch=master)](https://coveralls.io/github/ricardogama/api-mocker-go?branch=master) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](/LICENSE)

## Installation

Install the package via `go get` or add to your dependency manager file:

```sh
$ go get github.com/ricardogama/api-mocker-go
```

## Usage

Create a new mocker and use its sugar functions to communicate with the mocker server:

```go
package main

import (
  "fmt"
  "net/http"

  "github.com/ricardogama/api-mocker-go"
)

func main() {
  mock := mocker.New("http://localhost:3000")
  defer mock.Clear() // DELETE /mocks

  // POST /mocks
  err := mock.Expect(&mocker.Request{
    Method: "POST",
    Path: "/foo",
    Body: map[string]interface{}{
      "foo": "bar",
    },
    Headers: map[string][]string{
      "X-User-ID": []string{"1"},
    },
    Response: &mocker.Response{
      Status: 200,
      Headers: map[string][]string{
        "Content-Range": []string{"bytes 200-1000/67589"},
      },
      Body: map[string]interface{}{
        "qux": "qix",
      },
    },
  })

  if err != nil {
    panic(err)
  }

  _, err := http.Get("http://localhost:3000/foo")
  if err != nil {
      panic(err)
  }

  // GET /mocks, returning an error if there's unexpected requests.
  err := mock.Ensure()
  fmt.Println(err)
  // missing 1 expected calls: [
  //     {
  //         "headers": {
  //             "X-User-ID": ["1"]
  //         },
  //         "body": {
  //             "foo": "bar"
  //         },
  //         "method": "POST",
  //         "path": "/foo",
  //         "response": {
  //             "headers": {
  //                 "Content-Range": ["bytes 200-1000/67589"]
  //             },
  //             "body": {
  //                 "qux": "qix"
  //             },
  //             "status": 200
  //         }
  //     }
  // ]
  // 1 unexpected calls: [
  //     {
  //         "headers": {
  //             "accept-encoding": ["gzip"],
  //             "host": ["localhost:3005"],
  //             "user-agent": ["Go-http-client/1.1"]
  //         },
  //         "body": {},
  //         "method": "GET",
  //         "path": "/foo",
  //         "response": null
  //     }
  // ]
  })
}
```

## License

[MIT](/LICENSE)
