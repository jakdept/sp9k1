// go:generate statik

package main

import (
	"context"
	"crypto/tls"
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
	// common options
	enableTLS = kingpin.Flag("tls", "enables auto-TLS and push to https").Default("false").Bool()
	port      = kingpin.Flag("port", "port to listen on").Default("80").Int()
	domain    = kingpin.Flag("domain", "domain/ip to listen on").Default("localhost").String()
	imgDir    = kingpin.Flag("images", "directory of images to serve").Short('i').Default("./").ExistingDir()

	// less common options
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
	h = dandler.Split(http.RedirectHandler("/", 301), h)
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
	h = dandler.Split(http.RedirectHandler("/", 301), h)
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

	h = handlers.CombinedLoggingHandler(os.Stdout, h)

	// compress responses
	h = handlers.CompressHandler(h)

	return h
}

func launchServers(h http.Handler, d chan struct{}, e chan<- error) {
	if !*enableTLS {
		// start other http listeners
		s := &http.Server{
			Addr:    fmt.Sprintf(":%d", *port),
			Handler: h,
		}
		if *domain == "localhost" {
			s.Addr = fmt.Sprintf("%s:%d", *domain, *port)
		}
		go func() {
			defer close(d)
			e <- s.ListenAndServe()
		}()

		go func() {
			<-d
			s.Shutdown(context.Background())
		}()
		return
	}

	cert := &autocert.Manager{
		Cache:      autocert.DirCache("secret-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(*domain),
	}

	canonicalURL := fmt.Sprint("https://", *domain, ":", *port)
	// start a port 80 listener for cert
	// create the http server on port 80
	srvHTTP := &http.Server{
		Addr:    ":http",
		Handler: cert.HTTPHandler(http.RedirectHandler(canonicalURL, 301)),
	}
	go func() {
		defer close(d)
		e <- srvHTTP.ListenAndServe()
	}()

	// create the https server
	srvHTTPS := &http.Server{
		Addr:      fmt.Sprintf(":%d", *port),
		Handler:   h,
		TLSConfig: &tls.Config{GetCertificate: cert.GetCertificate},
	}
	if *domain == "localhost" {
		srvHTTPS.Addr = fmt.Sprintf("%s:%d", *domain, *port)
	}
	go func() {
		defer close(d)
		e <- srvHTTPS.ListenAndServeTLS("", "")
	}()

	go func() {
		<-d
		srvHTTP.Shutdown(context.Background())
		srvHTTPS.Shutdown(context.Background())
	}()
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
	var errChan chan error
	go func() {
		for e := range errChan {
			logger.Fatal(e)
		}
	}()

	done := make(chan struct{})
	defer close(done)

	fs := createStaticFS(logger, *staticDir)

	templ := parseTemplate(logger, fs)

	srvHandlers := buildMuxer(logger, fs, templ, done)

	fmt.Println("launching servers")
	launchServers(srvHandlers, done, errChan)
	fmt.Println("launched servers")
	<-done
}
