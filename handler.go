package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// ServeFile serves a given file back to the requester, and determines content type by algorithm only.
// It does not use the file's extension to determine the content type.
func ServeFile(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Clean(r.URL.Path))
	if err != nil {
		http.Error(w, fmt.Sprintf("%s not found", r.URL.Path), http.StatusNotFound)
		return
	}

	chunk := make([]byte, 512)

	_, err = f.Read(chunk)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file %s", r.URL.Path), http.StatusForbidden)
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file %s", r.URL.Path), http.StatusInternalServerError)
		return
	}

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file %s", r.URL.Path), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(chunk))
	http.ServeContent(w, r, r.URL.Path, stat.ModTime(), f)

	return
}

func main() {

}
