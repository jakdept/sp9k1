package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	_ "github.com/jakdept/sp9k1/statik"
	"github.com/rakyll/statik/fs"
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

var handlerTestData = []struct {
	uri           string
	code          int
	md5           string
	contentLength int64
}{
	{
		uri:           "/default.css",
		code:          200,
		md5:           "b1cf11f4d2cda79f08a58383863346a7",
		contentLength: 1868,
	},
}

func TestInternalHandler(t *testing.T) {
	ts := httptest.NewServer(InternalHandler(testFS))
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse url: %s", err)
	}

	for testID, test := range handlerTestData {

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
		assert.Equal(t, test.contentLength, res.ContentLength,
			"#%d [%s] - ContentLength does not match", testID, test.uri)

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

func TestSplitHandler(t *testing.T) {

}
