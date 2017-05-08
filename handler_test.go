// built with goldie
// if golden files in fixture dir are manually verified, you can update with
// go test -update

package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	_ "github.com/jakdept/sp9k1/statik"
	"github.com/rakyll/statik/fs"
	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

var testFS http.FileSystem

func init() {
	var err error
	testFS, err = fs.New()
	if err != nil {
		log.Fatalf("Failed to load statik fs, aborting tests: %s", err)
	}
}

func TestInternalHandler(t *testing.T) {
	var testData = []struct {
		uri           string
		code          int
		md5           string
		contentLength int64
		contentType   string
	}{
		{
			uri:           "/component.css",
			code:          200,
			md5:           "6929ee1f5b86c6e5669334b34e8fea65",
			contentLength: 3548,
			contentType:   "text/css; charset=utf-8",
		}, {
			uri:           "/default.css",
			code:          200,
			md5:           "b1cf11f4d2cda79f08a58383863346a7",
			contentLength: 1868,
			contentType:   "text/css; charset=utf-8",
		}, {
			uri:           "/grid.js",
			code:          200,
			md5:           "c1b9a03d47a42720891989a5844e9e3c",
			contentLength: 14173,
			contentType:   "application/javascript",
		}, {
			uri:           "/modernizr.custom.js",
			code:          200,
			md5:           "3d025169b583ce5c3af13060440e2277",
			contentLength: 8281,
			contentType:   "application/javascript",
		}, {
			uri:           "/page.html",
			code:          200,
			md5:           "9676bd8257ddcd3aa6a4e50a6068a3f8",
			contentLength: 5607,
			contentType:   "text/html; charset=utf-8",
		}, {
			uri:           "/bad.target",
			code:          404,
			md5:           "",
			contentLength: 0,
			contentType:   "",
		}, {
			uri:           "/page.template",
			code:          403,
			md5:           "23115a2a2e7d25f86bfb09392986681d",
			contentLength: 0,
			contentType:   "text/html; charset=utf-8",
		},
	}
	ts := httptest.NewServer(InternalHandler(testFS))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range testData {

		uri, err := url.Parse(test.uri)
		if err != nil {
			t.Errorf("bad URI path: [%s]", test.uri)
			continue
		}

		res, err := http.Get(baseURL.ResolveReference(uri).String())
		if err != nil {
			t.Error(err)
			continue
		}

		assert.Equal(t, test.code, res.StatusCode,
			"#%d [%s] - status code does not match", testID, test.uri)
		if test.code != 200 {
			if res.StatusCode != test.code {
				t.Logf("the response returned: \n%#v\n", res)
			}
			continue
		}
		assert.Equal(t, test.contentLength, res.ContentLength,
			"#%d [%s] - ContentLength does not match", testID, test.uri)
		assert.Equal(t, test.contentType, res.Header.Get("Content-Type"),
			"#%d [%s] - Content-Type does not match", testID, test.uri)

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Error(err)
			continue
		}
		assert.Equal(t, test.md5, fmt.Sprintf("%x", md5.Sum(body)),
			"#%d [%s] - mismatched body returned", testID, test.uri)
	}
}

func TestContentTypeHandler(t *testing.T) {
	var testData = []struct {
		uri           string
		code          int
		md5           string
		contentLength int64
		contentType   string
	}{
		{
			uri:           "/component.css",
			code:          200,
			md5:           "6929ee1f5b86c6e5669334b34e8fea65",
			contentLength: 3548,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/default.css",
			code:          200,
			md5:           "b1cf11f4d2cda79f08a58383863346a7",
			contentLength: 1868,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/grid.js",
			code:          200,
			md5:           "c1b9a03d47a42720891989a5844e9e3c",
			contentLength: 14173,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/modernizr.custom.js",
			code:          200,
			md5:           "3d025169b583ce5c3af13060440e2277",
			contentLength: 8281,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/page.html",
			code:          200,
			md5:           "9676bd8257ddcd3aa6a4e50a6068a3f8",
			contentLength: 5607,
			contentType:   "text/html; charset=utf-8",
		}, {
			uri:           "/bad.target",
			code:          404,
			md5:           "",
			contentLength: 0,
			contentType:   "",
		}, {
			uri:           "/page.template",
			code:          200,
			md5:           "23115a2a2e7d25f86bfb09392986681d",
			contentLength: 1503,
			contentType:   "text/html; charset=utf-8",
		}, {
			uri:           "/lemur_pudding_cups.jpg",
			code:          200,
			md5:           "f805ae46588af757263407301965c6a0",
			contentLength: 41575,
			contentType:   "image/jpeg",
		}, {
			uri:           "/spooning_a_barret.png",
			code:          200,
			md5:           "09d8be7d937b682447348acdc38c5895",
			contentLength: 395134,
			contentType:   "image/png",
		}, {
			uri:           "/whats_in_the_case.gif",
			code:          200,
			md5:           "2f43a5317fa2f60dbf32276faf3f139a",
			contentLength: 32933853,
			contentType:   "image/gif",
			// it's vitally important that we can serve files with the WRONG extension correctly
		}, {
			uri:           "/accidentally_save_file.gif",
			code:          200,
			md5:           "a305f39d197dce79acae597e81e22bf4",
			contentLength: 187967,
			contentType:   "image/png",
		}, {
			uri:           "/blocked_us.png",
			code:          200,
			md5:           "bc2272b02e6fab9c0c48d4743d4aae7e",
			contentLength: 45680,
			contentType:   "image/jpeg",
		}, {
			uri:           "/carlton_pls.jpg",
			code:          200,
			md5:           "7c0dc59a6ebad1645fca205f701edb39",
			contentLength: 871029,
			contentType:   "image/gif",
		},
	}
	ts := httptest.NewServer(ContentTypeHandler("./testdata/"))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range testData {

		uri, err := url.Parse(test.uri)
		if err != nil {
			t.Errorf("bad URI path: [%s]", test.uri)
			continue
		}

		res, err := http.Get(baseURL.ResolveReference(uri).String())
		if err != nil {
			t.Error(err)
			continue
		}

		assert.Equal(t, test.code, res.StatusCode,
			"#%d [%s] - status code does not match", testID, test.uri)
		if test.code != 200 {
			if res.StatusCode != test.code {
				t.Logf("the response returned: \n%#v\n", res)
			}
			continue
		}
		assert.Equal(t, test.contentLength, res.ContentLength,
			"#%d [%s] - ContentLength does not match", testID, test.uri)
		assert.Equal(t, test.contentType, res.Header.Get("Content-Type"),
			"#%d [%s] - Content-Type does not match", testID, test.uri)

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Error(err)
			continue
		}
		assert.Equal(t, test.md5, fmt.Sprintf("%x", md5.Sum(body)),
			"#%d [%s] - mismatched body returned", testID, test.uri)
	}
}

func TestIndexHandler_successful(t *testing.T) {
	templateString := `{
	"all":{
	{{- range $index, $value := .All -}}
	{{- if $index }}, {{ end -}}
	"{{- . -}}"
	{{- end -}}
	},
	"files":{
	{{- range $index, $value := .Files -}}
	{{- if $index }}, {{ end -}}
	"{{- . -}}"
	{{- end -}}
	},
	"dirs":{
	{{- range $index, $value := .Dirs -}}
	{{- if $index }}, {{ end -}}
	"{{- . -}}"
	{{- end -}}
	}
}`

	testTempl := template.Must(template.New("test").Parse(templateString))

	ts := httptest.NewServer(IndexHandler("testdata", testTempl))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read response: [%s]", err)
	}
	res.Body.Close()

	goldie.Assert(t, "templateHandler", actual)
}

// test to make sure a bad folder kicks a 404
func TestIndexHandler_badpath(t *testing.T) {
	templateString := ""
	testTempl := template.Must(template.New("test").Parse(templateString))

	ts := httptest.NewServer(IndexHandler("not-a-folder", testTempl))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 404, res.StatusCode, "got wrong response")
}

// test to make sure that a bad template kicks a 500
func TestIndexHandler_badtemplate(t *testing.T) {
	templateString := "{{ .ValueNotPresent }}"
	testTempl := template.Must(template.New("test").Parse(templateString))

	ts := httptest.NewServer(IndexHandler("testdata", testTempl))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 500, res.StatusCode, "got wrong response")
}

// handler will always response with 200 ok and the given body
func foundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ohai mr client")
	})
}

func TestSplitHandler(t *testing.T) {
	testData := []struct {
		uri  string
		code int
	}{
		{uri: "/", code: 200},
		{uri: "/other", code: 404},
		{uri: "", code: 200},
		{uri: "./", code: 200},
		{uri: "bad/url", code: 404},
	}

	// setup a handler that returns one thing on the main path, and another on other paths
	ts := httptest.NewServer(SplitHandler(foundHandler(), http.NotFoundHandler()))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range testData {

		uri, err := url.Parse(test.uri)
		if err != nil {
			t.Errorf("bad URI path: [%s]", test.uri)
			continue
		}

		res, err := http.Get(baseURL.ResolveReference(uri).String())
		if err != nil {
			t.Error(err)
			continue
		}
		if res.StatusCode == 200 {
			res.Body.Close()
		}

		assert.Equal(t, test.code, res.StatusCode, "#%d - not routed properly", testID)
	}
}

func TestGenerateThumbnail(t *testing.T) {
	testImages := []string{"accidentally_save_file.gif", "spooning_a_barret.png"}

	h := thumbnailHandler{x: 100, y: 100}
	buf := bytes.NewBuffer([]byte{})

	for id, inputImage := range testImages {
		img, _, err := h.openImage("testdata/" + inputImage)
		if err != nil {
			t.Logf("#%d - failed to open test image [%s]: %s", id, inputImage, err)
			t.Fail()
			continue
		}
		thumb, err := h.generateThumbnail(img)
		if err != nil {
			t.Logf("#%d - failed to generate thumbnail [%s]: %s", id, inputImage, err)
			t.Fail()
			continue
		}
		buf.Reset()
		err = png.Encode(buf, thumb)
		if err != nil {
			t.Logf("#%d - failed to encode thumbnail [%s]: %s", id, inputImage, err)
			t.Fail()
			continue
		}

		goldie.Assert(t, inputImage, buf.Bytes())
	}
}

func TestGeneratePaths(t *testing.T) {
	testData := []struct {
		imageName string
		rawPath   string
		thumbPath string
	}{
		{
			imageName: "accidentally_save_file.gif",
			rawPath:   "testdata/accidentally_save_file.gif",
			thumbPath: "output/accidentally_save_file.gif.jpg",
		}, {
			imageName: "blocked_us.png",
			rawPath:   "testdata/blocked_us.png",
			thumbPath: "output/blocked_us.png.jpg",
		}, {
			imageName: "carlton_pls.jpg",
			rawPath:   "testdata/carlton_pls.jpg",
			thumbPath: "output/carlton_pls.jpg.jpg",
		}, {
			imageName: "lemur_pudding_cups.jpg",
			rawPath:   "testdata/lemur_pudding_cups.jpg",
			thumbPath: "output/lemur_pudding_cups.jpg.jpg",
		}, {
			imageName: "spooning_a_barret.png",
			rawPath:   "testdata/spooning_a_barret.png",
			thumbPath: "output/spooning_a_barret.png.jpg",
		}, {
			imageName: "whats_in_the_case.gif",
			rawPath:   "testdata/whats_in_the_case.gif",
			thumbPath: "output/whats_in_the_case.gif.jpg",
		},
	}

	h := thumbnailHandler{raw: "testdata", thumbs: "output", thumbExt: "jpg"}

	for id, test := range testData {
		assert.Equal(t, test.rawPath, h.generateRawPath(test.imageName), "#%d - wrong raw path", id)
		assert.Equal(t, test.thumbPath, h.generateThumbPath(test.imageName), "#%d - wrong thumb path", id)
	}
}

func TestLoadThumbnail(t *testing.T) {
	testData := []struct {
		imageName string
		size      int64
	}{
		{
			imageName: "accidentally_save_file.gif",
			size:      17861,
		}, {
			imageName: "blocked_us.png",
			size:      44940,
		}, {
			imageName: "carlton_pls.jpg",
			size:      22806,
		}, {
			imageName: "lemur_pudding_cups.jpg",
			size:      72840,
		}, {
			imageName: "spooning_a_barret.png",
			size:      47306,
		}, {
			imageName: "whats_in_the_case.gif",
			size:      48763,
		},
	}

	tempdir, err := ioutil.TempDir("", "sp9k1-")
	if err != nil {
		t.Fatalf("failed creating test directory: %s", err)
	}

	h := thumbnailHandler{x: 200, y: 200, raw: "testdata", thumbExt: "png", thumbs: tempdir}

	for id, test := range testData {
		h.loadThumbnail(test.imageName)
		info, err := os.Stat(h.generateThumbPath(test.imageName))
		if err != nil {
			t.Logf("#%d - failed to stat thumbnail [%s] tempdir [%s]: %s",
				id, test.imageName, tempdir, err)
			t.Fail()
			continue
		}
		assert.Equal(t, test.size, info.Size(),
			"#%d [%s] - size does not match - tempDir [%s]", id, test.size, tempdir)
	}
}

func TestThumbnailHandler(t *testing.T) {
	var testData = []struct {
		uri           string
		code          int
		md5           string
		contentLength int64
		contentType   string
	}{
		{
			uri:           "/accidentally_save_file.gif.png",
			code:          200,
			md5:           "6929ee1f5b86c6e5669334b34e8fea65",
			contentLength: 3548,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/blocked_us.png.png",
			code:          200,
			md5:           "b1cf11f4d2cda79f08a58383863346a7",
			contentLength: 1868,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/carlton_pls.jpg.png",
			code:          200,
			md5:           "c1b9a03d47a42720891989a5844e9e3c",
			contentLength: 14173,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/lemur_pudding_cups.jpg.png",
			code:          200,
			md5:           "3d025169b583ce5c3af13060440e2277",
			contentLength: 8281,
			contentType:   "text/plain; charset=utf-8",
		}, {
			uri:           "/spooning_a_barret.png.png",
			code:          200,
			md5:           "9676bd8257ddcd3aa6a4e50a6068a3f8",
			contentLength: 5607,
			contentType:   "text/html; charset=utf-8",
		}, {
			uri:           "/whats_in_the_case.gif.png",
			code:          200,
			md5:           "9676bd8257ddcd3aa6a4e50a6068a3f8",
			contentLength: 5607,
			contentType:   "text/html; charset=utf-8",
		}, {
			uri:           "/bad.target.png",
			code:          404,
			md5:           "",
			contentLength: 0,
			contentType:   "",
		},
	}

	tempdir, err := ioutil.TempDir("", "sp9k1-")
	if err != nil {
		t.Fatalf("failed creating test directory: %s", err)
	}

	ts := httptest.NewServer(ThumbnailHandler(300, 250, "./testdata/", tempdir, "png"))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range testData {
		for _, run := range []string{"a", "b"} {
			uri, err := url.Parse(test.uri)
			if err != nil {
				t.Errorf("bad URI path: [%s]", test.uri)
				continue
			}

			res, err := http.Get(baseURL.ResolveReference(uri).String())
			if err != nil {
				t.Error(err)
				continue
			}

			assert.Equal(t, test.code, res.StatusCode,
				"#%d-%s [%s] tempdir [%s] - status code does not match: ", testID, run, test.uri, tempdir)
			if test.code != 200 {
				if res.StatusCode != test.code {
					t.Logf("the response returned: \n%#v\n", res)
				}
				continue
			}
			// assert.Equal(t, test.contentLength, res.ContentLength,
			// 	"#%d-%s [%s] tempdir [%s] - ContentLength does not match: ", testID, run, test.uri, tempdir)
			// assert.Equal(t, test.contentType, res.Header.Get("Content-Type"),
			// 	"#%d-%s [%s] tempdir [%s]- Content-Type does not match: ", testID, run, test.uri, tempdir)

			// body, err := ioutil.ReadAll(res.Body)
			// res.Body.Close()
			// if err != nil {
			// 	t.Error(err)
			// 	continue
			// }
			// assert.Equal(t, test.md5, fmt.Sprintf("%x", md5.Sum(body)),
			// 	"#%d-%s [%s] tempdir [%s]- mismatched body returned: ", testID, run, test.uri, tempdir)
		}
	}
}
