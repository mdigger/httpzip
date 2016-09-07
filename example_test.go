package httpzip_test

import (
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/mdigger/httpzip"
)

func Example() {
	// open zip-file
	zipServer, err := httpzip.Open("static.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer zipServer.Close()

	// initialize http handler
	http.Handle("/static/", http.StripPrefix("/static/", zipServer))
}

func Example_2() {
	// open zip-file
	zipServer, err := httpzip.Open("static.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer zipServer.Close()

	// read file data
	data, err := zipServer.GetData("templates/default.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.New("").Parse(string(data))
	if err != nil {
		log.Fatal(err)
	}
	if err := tmpl.Execute(os.Stdout, "test"); err != nil {
		log.Fatal(err)
	}
}
