// go:generate statik -src=./static

package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/jakdept/dandler"
	"github.com/jakdept/flagTrap"
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

const (
	defaultListen         = ":8080"
	defaultImgDir         = "./"
	defaultWidth          = 310
	defaultHeight         = 200
	defaultCacheDays      = 30
	defaultCacheVariation = 7
	defaultURL            = "localhost:80"
)

var (
	listenAddress       string
	imgDir              string
	thumbWidth          int
	thumbHeight         int
	cacheMinDays        int
	cacheVariation      int
	staticDir           flagTrap.StringTrap
	templateFile        flagTrap.StringTrap
	canonicalURL        string
	canonicalForceHost  bool
	canonicalForcePort  bool
	canonicalDisableTLS bool
	canonicalForceTLS   bool
)

func flags() {
	usage := "address to listen for incoming traffic"
	flag.StringVar(&listenAddress, "listen", defaultListen, usage)
	flag.StringVar(&listenAddress, "l", defaultListen, usage+" (shorthand)")

	usage = "directory of images to serve"
	flag.StringVar(&imgDir, "images", defaultImgDir, usage)
	flag.StringVar(&imgDir, "i", defaultImgDir, usage+" (shorthand)")

	usage = "cache length"
	flag.IntVar(&cacheMinDays, "cacheTime", defaultCacheDays, usage)

	usage = "cache variation"
	flag.IntVar(&cacheVariation, "cacheSkew", defaultCacheVariation, usage)

	usage = "thumbnail width"
	flag.IntVar(&thumbWidth, "width", defaultWidth, usage)
	flag.IntVar(&thumbWidth, "w", defaultWidth, usage+" (shorthand)")

	usage = "thumbnail height"
	flag.IntVar(&thumbHeight, "height", defaultHeight, usage)
	flag.IntVar(&thumbHeight, "h", defaultHeight, usage+" (shorthand)")

	usage = "alternate static directory to serve"
	flag.Var(&staticDir, "static", usage)
	flag.Var(&staticDir, "s", usage+" (shorthand)")

	usage = "alternate index template to serve"
	flag.Var(&templateFile, "template", usage)
	flag.Var(&templateFile, "t", usage+" (shorthand)")

	flag.StringVar(&canonicalURL, "canonicalURl", defaultURL, "canonical host to force")
	flag.BoolVar(&canonicalDisableTLS, "canonicalDisableTLS", false, "force unencrypted protocol")
	flag.BoolVar(&canonicalForceTLS, "canonicalForceTLS", false, "force encrypted protocol")
	flag.BoolVar(&canonicalForceHost, "canonicalForceHost", false, "force a specific hostname")
	flag.BoolVar(&canonicalForcePort, "canonicalForcePort", false, "force a specific port")

	flag.Parse()
}

func parseTemplate(logger *log.Logger, fs http.FileSystem) *template.Template {
	if templateFile.IsSet() {
		// if an alternate template was provided, i can use that instead
		return template.Must(template.ParseFiles(templateFile.String()))
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

func createStaticFS(logger *log.Logger, path flagTrap.StringTrap) http.FileSystem {
	if path.IsSet() {
		return http.Dir(path.String())
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
	h = dandler.ExpiresRange(day*time.Duration(cacheMinDays),
		day*time.Duration(cacheVariation), h)
	// add the static handler to the muxer
	mux.Handle("/static/", h)

	// create a caching handler
	h = dandler.ThumbCache(logger, thumbWidth, thumbHeight, 32<<20, imgDir, "thumbs", "jpg")
	// split the folder itself into a redirect
	h = dandler.Split(http.RedirectHandler("/", 302), h)
	// strip the prefix
	h = http.StripPrefix("/thumb/", h)
	// add some expiration
	h = dandler.ExpiresRange(day*time.Duration(cacheMinDays),
		day*time.Duration(cacheVariation), h)
	// add the thumbnail handler to the muxer
	mux.Handle("/thumb/", h)

	h = dandler.DirSplit(logger, imgDir, done,
		dandler.Index(logger, imgDir, done, templ),
		dandler.ContentType(logger, imgDir),
	)
	mux.Handle("/", h)

	h = dandler.ASCIIHeader("shit\nposting\n9001", serverBanner, " ", mux)
	h = handlers.CombinedLoggingHandler(os.Stdout, h)

	// add canonical header if required
	if canonicalForceHost ||
		canonicalForcePort ||
		canonicalForceTLS ||
		canonicalDisableTLS {
		options := 0
		if canonicalForceHost {
			options += dandler.ForceHost
		}
		if canonicalForcePort {
			options += dandler.ForcePort
		}
		if canonicalForceTLS {
			options += dandler.ForceHTTPS
		} else if canonicalDisableTLS {
			options += dandler.ForceHTTP
		}

		h = dandler.CanonicalHost(canonicalURL, options, h)
	}

	// compress responses
	h = handlers.CompressHandler(h)

	return h
}

func main() {

	flags()

	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Llongfile)

	fs := createStaticFS(logger, staticDir)

	templ := parseTemplate(logger, fs)

	done := make(chan struct{})
	defer close(done)

	srvHandlers := buildMuxer(logger, fs, templ, done)

	logger.Fatal(http.ListenAndServe(listenAddress, srvHandlers))
}
