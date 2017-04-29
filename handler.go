package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// ServeFile serves a given file back to the requester, and determines content type by algorithm only.
// It does not use the file's extension to determine the content type.
func ServeFile(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(filepath.Join(basePath, r.URL.Path))
		if err != nil {
			http.Error(w, fmt.Sprintf("not found: %s", r.URL.Path), http.StatusNotFound)
			log.Printf("404 - could not open file: %s - %s", filepath.Join(basePath, r.URL.Path), err)
			return
		}

		stat, err := f.Stat()
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusForbidden)
			log.Printf("403 - could not stat file: %s - %s", filepath.Join(basePath, r.URL.Path), err)
			return
		}

		chunk := make([]byte, 512)

		_, err = f.Read(chunk)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusForbidden)
			log.Printf("403 - could not read from file: %s - %s", filepath.Join(basePath, r.URL.Path), err)
			return
		}

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
			log.Printf("500 - could not seek within file: %s - %s", filepath.Join(basePath, r.URL.Path), err)
			return
		}

		w.Header().Set("Content-Type", http.DetectContentType(chunk))
		http.ServeContent(w, r, r.URL.Path, stat.ModTime(), f)

		return

	}
}

// ServeFolder lists all files in a directory, and passes them to template execution to build a directory listing.
func ServeFolder(basePath string, templ *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(filepath.Join(basePath, r.URL.Path))
		if err != nil {
			http.Error(w, fmt.Sprintf("not found: %s", r.URL.Path), http.StatusNotFound)
			log.Printf("404 - could not find file: %s - %s", filepath.Join(basePath, r.URL.Path), err)
			return
		}

		stat, err := f.Stat()
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot read target: %s", r.URL.Path), http.StatusInternalServerError)
			log.Printf("403 - could not stat file: %s - %s", filepath.Join(basePath, r.URL.Path), err)
			return
		}

		if !stat.IsDir() {
			ServeFile(basePath)(w, r)
			return
		}

		contents, err := f.Readdir(0)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot read directory: %s", r.URL.Path), http.StatusForbidden)
			log.Printf("403 - could not read file: %s - %s", filepath.Join(basePath, r.URL.Path), err)
			return
		}

		var data struct {
			files []string
			dirs  []string
			all   []string
		}

		for _, each := range contents {
			data.all = append(data.all, each.Name())
			switch each.IsDir() {
			case true:
				data.dirs = append(data.dirs, each.Name())
			default:
				data.files = append(data.files, each.Name())
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = templ.Execute(w, data)
		if err != nil {
			http.Error(w, fmt.Sprintf("error building response: %s", r.URL.Path), http.StatusInternalServerError)
			log.Printf("500 - error responding: %s", err)
			return
		}
	}
}
