//go:generate statik -src=./static

package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	_ "github.com/jakdept/sp9k1/statik"
)

// SplitHandler allows the routing of one handler at /, and another at all locations below /.
func SplitHandler(root, more http.Handler) http.Handler {
	return splitHandler{bare: root, more: more}
}

type splitHandler struct {
	bare http.Handler
	more http.Handler
}

func (p splitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if path.Clean(r.URL.Path) == "/" {
		p.bare.ServeHTTP(w, r)
	} else {
		p.more.ServeHTTP(w, r)
	}
}

// InternalHandler serves a static, in memory filesystem..
func InternalHandler(fs http.FileSystem) http.Handler {
	return internalHandler{handler: http.FileServer(fs)}
}

type internalHandler struct {
	handler http.Handler
}

func (c internalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if path.Ext(r.URL.Path) == ".template" {
		http.Error(w, fmt.Sprintf("template requested, blocked: %s", r.URL.Path), http.StatusForbidden)
		log.Printf("403 - error responding: %s", r.URL.Path)
		return
	}

	c.handler.ServeHTTP(w, r)
	return
}

// IndexHandler lists all files in a directory, and passes them to template execution to build a directory listing.
func IndexHandler(basePath string, templ *template.Template) http.Handler {
	return indexHandler{basePath, templ}
}

type IndexData struct {
	Files []string
	Dirs  []string
	All   []string
}

type indexHandler struct {
	basePath string
	templ    *template.Template
}

func (c indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Join(c.basePath, r.URL.Path))
	if err != nil {
		http.Error(w, fmt.Sprintf("not found: %s", r.URL.Path), http.StatusNotFound)
		log.Printf("404 - could not find file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read target: %s", r.URL.Path), http.StatusInternalServerError)
		log.Printf("500 - could not stat file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	if !stat.IsDir() {
		http.Error(w, fmt.Sprintf("cannot read target: %s", r.URL.Path), http.StatusForbidden)
		log.Printf("403 - could not stat file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	contents, err := f.Readdir(0)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read directory: %s", r.URL.Path), http.StatusForbidden)
		log.Printf("403 - could not read file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	var data IndexData

	for _, each := range contents {
		data.All = append(data.All, each.Name())
		switch each.IsDir() {
		case true:
			data.Dirs = append(data.Dirs, each.Name())
		default:
			data.Files = append(data.Files, each.Name())
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = c.templ.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("error building response: %s", r.URL.Path), http.StatusInternalServerError)
		log.Printf("500 - error responding: %s", err)
		return
	}
}

// ContentTypeHandler serves a given file back to the requester, and determines content type by algorithm only.
// It does not use the file's extension to determine the content type.
func ContentTypeHandler(basePath string) http.Handler {
	return contentTypeHandler{basePath}
}

type contentTypeHandler struct {
	basePath string
}

// contentTypeHandler.ServeHTTP satasifies the Handler interface.
func (c contentTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Join(c.basePath, r.URL.Path))
	if err != nil {
		http.Error(w, fmt.Sprintf("not found: %s", r.URL.Path), http.StatusNotFound)
		log.Printf("404 - could not open file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		log.Printf("500 - could not stat file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	chunk := make([]byte, 512)

	_, err = f.Read(chunk)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		log.Printf("500 - could not read from file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %s", r.URL.Path), http.StatusInternalServerError)
		log.Printf("500 - could not seek within file: %s - %s", filepath.Join(c.basePath, r.URL.Path), err)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(chunk))
	http.ServeContent(w, r, r.URL.Path, stat.ModTime(), f)

	return
}
