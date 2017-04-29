package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {

	templ, err := template.New("page.html").ParseFiles("page.html")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", ServeFolder("sample_images", templ))
	mux.Handle("/static/", http.StripPrefix("/static/", ServeFile("static")))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
