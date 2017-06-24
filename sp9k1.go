// go:generate statik -src=./static

package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/gorilla/handlers"
	"github.com/jakdept/dandler"
	_ "github.com/jakdept/sp9k1/statik"
	"github.com/rakyll/statik/fs"
)

// now would this be shitposting if there were _tests_?

var serverBanner = `
'______________________________________________________________________________
/                                                                              \
|                 '.'                           .-:::::::::::::::::::::-'      | 
|                 -///-'                      '/+++++++++++++++++++++++++-     | 
|                ':+++++/-                    -++//////////////////////+++     | 
|                /++++++//:.                  -+/----------------------:++     | 
|               '/++++///:::.                 -+/---:dddddddds+ymmms----++     | 
|             .://+////:::::-                 -+/...:mmmmmdyoymNNNNy...-++     | 
|           '://////::::::::::-.              -++::-:mmmdyoymNNNNNNy--:/++     | 
|           ://:::::::::::/+++//:'            -+++++ommmhoymNNNNNNNh++++++     | 
|           :/syys+:::///++syhyo:.            -+++++ommmmdsodNNNNNNh++++++     | 
|         '.+mdo+hmh++++/ommo+yNh-''          -+++++ommmmmmhosmNNNNh++++++     | 
|       .//+mN.'''hNs///:mN:'''sNy//-         -+++++ommmmmmmdyohNNNh++++++     | 
|      '///+NN.'''hNs::::mM-'''sMy/::-        -++++++dmmmmmmmmdssdNh++++++     | 
|      '::::yNd+/yMd::://sNmo/sMm/:::-        -++++++oyyyyyyyyyyo+s+++++++     | 
|      '-::::+hmmds//++++/+ydmds/::::-'       -+++++++++++++++++++++++++++     | 
|   '-///++++++++++++++++++///:::::////:.     '/+++++++++++++++++++++++++-     | 
|  .///////////+hmmmmmmmmmmmh+::://+///::-      .:::::::+++++++++/:::::-'      | 
|  ::////////::::+shmmNmmhs+:://++//::::::.             +++++++++:             | 
|  -:::::::::::::::::::::///+++///:::::::-'             +++++++++:             | 
|   .-:::::::::::::::////+////:::::::::-.'              +++++++++:             | 
|      '''''''    ''''''''''''''''''''                  .........'             | 
\______________________________________________________________________________/
`

var (
	listenAddress = kingpin.Flag("listen", "addresses to listen for incoming non-TLS connections").
			Short('l').Default("127.0.0.1:8080").TCP()

	hostname = kingpin.Flag("hostname", "hostname to register").Short('h').String()

	imgDir = kingpin.Flag("images", "directory of images to serve").
		Short('i').Default("./").ExistingDir()

	cacheMinDays = kingpin.Flag("cacheMin", "minimum days to cache images in browser").
			Default("30").Int()

	cacheVariation = kingpin.Flag("cacheVariation", "difference between minimum and maximum length to cache images").
			Default("7").Int()

	thumbWidth  = kingpin.Flag("width", "maximum thumbnail width").Default("310").Int()
	thumbHeight = kingpin.Flag("height", "thumbnail height").Default("200").Int()

	staticDir = kingpin.Flag("static", "alternate static directory to serve").Short('s').ExistingDir()

	templateFile = kingpin.Flag("template", "alternate index template to serve").Short('t').ExistingFile()

	canonicalURL        = kingpin.Flag("canonicalURL", "default redirect to serve").Default("localhost:80").String()
	canonicalDisableTLS = kingpin.Flag("canonicalDisableTLS", "force unencrypted protocol").Default("false").Bool()
	canonicalForceTLS   = kingpin.Flag("canonicalForceTLS", "force encrypted protocol").Default("true").Bool()
	canonicalForceHost  = kingpin.Flag("canonicalForceHost", "force a specific hostname").Default("true").Bool()
	canonicalForcePort  = kingpin.Flag("canonicalForcePort", "force a specific port").Default("false").Bool()
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

	h = dandler.ASCIIHeader("shit\nposting\n9001", serverBanner, " ", mux)
	h = handlers.CombinedLoggingHandler(os.Stdout, h)

	// add canonical header if required
	if *canonicalForceHost ||
		*canonicalForcePort ||
		*canonicalForceTLS ||
		*canonicalDisableTLS {
		options := 0
		if *canonicalForceHost {
			options += dandler.ForceHost
		}
		if *canonicalForcePort {
			options += dandler.ForcePort
		}
		if *canonicalForceTLS {
			options += dandler.ForceHTTPS
		} else if *canonicalDisableTLS {
			options += dandler.ForceHTTP
		}

		h = dandler.CanonicalHost(*canonicalURL, options, h)
	}

	// compress responses
	h = handlers.CompressHandler(h)

	return h
}

func main() {

	kingpin.Parse()

	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Llongfile)

	fs := createStaticFS(logger, *staticDir)

	templ := parseTemplate(logger, fs)

	done := make(chan struct{})
	defer close(done)

	srvHandlers := buildMuxer(logger, fs, templ, done)

	logger.Fatal(http.ListenAndServe((*listenAddress).String(), srvHandlers))
}
