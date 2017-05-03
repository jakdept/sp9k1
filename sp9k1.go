package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rakyll/statik/fs"
)

func main() {

	fs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	templateFile, err := fs.Open("/page.html")
	if err != nil {
		log.Fatal(err)
	}

	templateData, err := ioutil.ReadAll(templateFile)
	if err != nil {
		log.Fatal(err)
	}

	templ, err := template.New("page.template").Parse(string(templateData))
	if err != nil {
		log.Fatal(err)
	}

	basePath := "./"

	mux := http.NewServeMux()

	mux.Handle("/", SplitHandler(
		IndexHandler(basePath, templ),
		ContentTypeHandler(basePath)))

	mux.Handle("/static/", http.StripPrefix("/static/", SplitHandler(
		http.RedirectHandler("/", 302),
		InternalHandler(fs))))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
