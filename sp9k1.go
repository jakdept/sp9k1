package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/rakyll/statik/fs"
)

func main() {

	fs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	templateFile, err := fs.Open("/page.template")
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
	thumbnailPath, err := ioutil.TempDir("", "thumbnailcache-")
	if err != nil {
		log.Fatalf("Could not create tempoary thumbnail directory - %s", err)
	}

	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Llongfile)

	mux := http.NewServeMux()
	done := make(chan struct{})
	defer close(done)

	mux.Handle("/", DirSplitHandler(logger, basePath, done,
		IndexHandler(logger, basePath, done, templ),
		ContentTypeHandler(logger, basePath)))

	mux.Handle("/static/", http.StripPrefix("/static/", SplitHandler(
		http.RedirectHandler("/", 302),
		InternalHandler(logger, fs))))

	mux.Handle("/thumb/", http.StripPrefix("/thumb/", SplitHandler(
		http.RedirectHandler("/", 302),
		ThumbnailHandler(logger, 310, 200, basePath, thumbnailPath, "jpg"))))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
