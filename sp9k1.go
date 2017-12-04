// go:generate statik

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/gorilla/handlers"
	"github.com/jakdept/dandler"
	_ "github.com/jakdept/sp9k1/statik"
	"github.com/rakyll/statik/fs"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// now would this be shitposting if there were _tests_?

var (
	listenAddress  = kingpin.Flag("listen", "non-TLS listen addresses").Short('l').Default("127.0.0.1:8080").TCPList()
	enableTLS      = kingpin.Flag("tls", "enables auto-TLS and push to https").Default("false").Bool()
	domain         = kingpin.Flag("domain", "domain to register, startup, and redirect to").String()
	imgDir         = kingpin.Flag("images", "directory of images to serve").Short('i').Default("./").ExistingDir()
	cacheMinDays   = kingpin.Flag("cacheMin", "minimum days to cache images in browser").Default("30").Int()
	cacheVariation = kingpin.Flag("cacheDayVariation", "browser cache variation").Default("7").Int()
	thumbWidth     = kingpin.Flag("width", "maximum thumbnail width").Default("310").Int()
	thumbHeight    = kingpin.Flag("height", "thumbnail height").Default("200").Int()
	staticDir      = kingpin.Flag("static", "alternate static directory to serve").Short('s').ExistingDir()
	templateFile   = kingpin.Flag("template", "alternate index template to serve").Short('t').ExistingFile()
)

func parseTemplate(logger *log.Logger, fs http.FileSystem) *template.Template {
	if *templateFile != "" {
		// if an alternate template was provided, i can use that instead
		return template.Must(template.ParseFiles(*templateFile))
	}
	// have to do it the hard way because it comes from fs
	templFile, err := fs.Open("/page.template")
	if err != nil {
		logger.Fatal(err)
	}
	templData, err := ioutil.ReadAll(templFile)
	if err != nil {
		logger.Fatal(err)
	}
	return template.Must(template.New("page.template").Parse(string(templData)))
}

func createStaticFS(logger *log.Logger, path string) http.FileSystem {
	if path != "" {
		return http.Dir(path)
	}
	filesystem, err := fs.New()
	if err != nil {
		logger.Fatal(err)
	}
	return filesystem
}

func buildMuxer(logger *log.Logger,
	fs http.FileSystem,
	templ *template.Template,
	done chan struct{},
) http.Handler {

	day := time.Hour * time.Duration(64)
	var h http.Handler
	mux := http.NewServeMux()

	// building the static handler
	h = http.FileServer(fs)
	// split the main folder off into a redirect
	h = dandler.Split(http.RedirectHandler("/", 302), h)
	// add a prefix before the handler
	h = http.StripPrefix("/static/", h)
	// add some expiration
	h = dandler.ExpiresRange(day*time.Duration(*cacheMinDays),
		day*time.Duration(*cacheVariation), h)
	// add the static handler to the muxer
	mux.Handle("/static/", h)

	// create a caching handler
	h = dandler.ThumbCache(logger, *thumbWidth, *thumbHeight, 32<<20, *imgDir, "thumbs", "jpg")
	// split the folder itself into a redirect
	h = dandler.Split(http.RedirectHandler("/", 302), h)
	// strip the prefix
	h = http.StripPrefix("/thumb/", h)
	// add some expiration
	h = dandler.ExpiresRange(day*time.Duration(*cacheMinDays),
		day*time.Duration(*cacheVariation), h)
	// add the thumbnail handler to the muxer
	mux.Handle("/thumb/", h)

	h = dandler.DirSplit(logger, *imgDir, done,
		dandler.Index(logger, *imgDir, done, templ),
		dandler.ContentType(logger, *imgDir),
	)
	mux.Handle("/", h)

	h = mux

	// add canonical header if required
	if *domain != "" {
		options := 0
		options += dandler.ForceHost
		if *enableTLS {
			options += dandler.ForceHTTPS
			h = dandler.CanonicalHost(fmt.Sprintf("%s:443", *domain), options, h)
		} else {
			h = dandler.CanonicalHost(fmt.Sprintf("%s:%d", *domain,
				(*listenAddress)[0].IP), options, h)
		}
	}

	h = handlers.CombinedLoggingHandler(os.Stdout, h)

	// compress responses
	h = handlers.CompressHandler(h)

	return h
}

func main() {

	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.CommandLine.Version("1.0")
	kingpin.CommandLine.Author("jakdept")
	kingpin.Parse()

	if *enableTLS && *domain == "" {
		log.Fatal("failed to start - if you specify --tls you must specify --domain")
	}

	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Llongfile)

	fs := createStaticFS(logger, *staticDir)

	templ := parseTemplate(logger, fs)

	done := make(chan struct{})
	defer close(done)

	srvHandlers := buildMuxer(logger, fs, templ, done)

	var errChan chan error
	go func() {
		for _, address := range *listenAddress {
			errChan <- http.ListenAndServe(address.String(), srvHandlers)
		}
	}()
	if *enableTLS {
		go func() {
			errChan <- http.Serve(autocert.NewListener(*domain), srvHandlers)
		}()
	}
	for e := range errChan {
		log.Fatal(e)
	}
}
