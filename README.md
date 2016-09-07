# httpzip

[![GoDoc](https://godoc.org/github.com/mdigger/httpzip?status.svg)](https://godoc.org/github.com/mdigger/httpzip)
[![Build Status](https://travis-ci.org/mdigger/httpzip.svg?branch=master)](https://travis-ci.org/mdigger/httpzip)
[![Coverage Status](https://coveralls.io/repos/github/mdigger/httpzip/badge.svg)](https://coveralls.io/github/mdigger/httpzip?branch=master)


Package httzip allows you to connect a zip archive to the web server as
static files handler.

```go
package main

import (
	"log"
	"net/http"

	"github.com/mdigger/httpzip"
)

func main() {
	// open zip-file
	zipServer, err := httpzip.Open("static.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer zipServer.Close()

	// initialize http handler
	http.Handle("/static/", http.StripPrefix("/static/", zipServer))
}
```

