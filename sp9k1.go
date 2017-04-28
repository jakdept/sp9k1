package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {

	templ, err := template.New("page.html").Parse("page.html")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8080", ServeFolder("./", templ)))
}
