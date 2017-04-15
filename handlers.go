package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func ServeFile(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Clean(r.URL.Path))
	if err != nil {
		http.Error(w, fmt.Sprintf("%s not found", r.URL.Path), http.StatusNotFound)
		return
	}

	fInfo, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("%s not found", r.URL.Path), http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, "", fInfo.ModTime(), f)
}

func main() {

}
