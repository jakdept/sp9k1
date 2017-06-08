// go:generate statik -src=./static

package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/jakdept/dandler"
	_ "github.com/jakdept/sp9k1/statik"
	"github.com/rakyll/statik/fs"
)

// now would this be shitposting if there were _tests_?

var serverBanner = `
                 '.'                           .-:::::::::::::::::::::-'        
                 -///-'                      '/+++++++++++++++++++++++++-       
                ':+++++/-                    -++//////////////////////+++       
                /++++++//:.                  -+/----------------------:++       
               '/++++///:::.                 -+/---:dddddddds+ymmms----++       
             .://+////:::::-                 -+/...:mmmmmdyoymNNNNy...-++       
           '://////::::::::::-.              -++::-:mmmdyoymNNNNNNy--:/++       
           ://:::::::::::/+++//:'            -+++++ommmhoymNNNNNNNh++++++       
           :/syys+:::///++syhyo:.            -+++++ommmmdsodNNNNNNh++++++       
         '.+mdo+hmh++++/ommo+yNh-''          -+++++ommmmmmhosmNNNNh++++++       
       .//+mN.'''hNs///:mN:'''sNy//-         -+++++ommmmmmmdyohNNNh++++++       
      '///+NN.'''hNs::::mM-'''sMy/::-        -++++++dmmmmmmmmdssdNh++++++       
      '::::yNd+/yMd::://sNmo/sMm/:::-        -++++++oyyyyyyyyyyo+s+++++++       
      '-::::+hmmds//++++/+ydmds/::::-'       -+++++++++++++++++++++++++++       
   '-///++++++++++++++++++///:::::////:.     '/+++++++++++++++++++++++++-       
  .///////////+hmmmmmmmmmmmh+::://+///::-      .:::::::+++++++++/:::::-'        
  ::////////::::+shmmNmmhs+:://++//::::::.             +++++++++:               
  -:::::::::::::::::::::///+++///:::::::-'             +++++++++:               
   .-:::::::::::::::////+////:::::::::-.'              +++++++++:               
      '''''''    ''''''''''''''''''''                  .........'               
`

func main() {

	listenAddress := flag.String("listen", ":8080", "address to liste")
	imageDir := flag.String("images", "./", "location of images to host")
	// thumbDir := flag.String("thumbs", "", "if set, location to hold thumbnails")
	staticDir := flag.String("static", "", "if set, alternate location to serve as /static/")
	templateFile := flag.String("template", "", "if set, alternate template to use")
	thumbWidth := flag.Int("thumbWidth", 310, "width of thumbnails to create")
	thumbHeight := flag.Int("thumbHeight", 200, "width of thumbnails to create")

	flag.Parse()

	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Llongfile)

	fs, err := fs.New()
	if err != nil {
		logger.Fatal(err)
	}

	var templ *template.Template
	if *templateFile == "" {
		// have to do it the hard way because it comes from fs
		templFile, err := fs.Open("/page.template")
		if err != nil {
			logger.Fatal(err)
		}
		templData, err := ioutil.ReadAll(templFile)
		if err != nil {
			logger.Fatal(err)
		}
		templ, err = template.New("page.template").Parse(string(templData))
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		// if an alternate template was provided, i can use that instead
		templ, err = template.ParseFiles(*templateFile)
		if err != nil {
			logger.Fatal(err)
		}
	}

	var thumbnailPath string
	// if *thumbDir != "" {
	// 	if *imageDir != "" {
	// 		thumbnailPath = fmt.Sprintf("%s-%s", *imageDir, *thumbDir)
	// 		err = os.MkdirAll(thumbnailPath, 0750)
	// 		if err != nil {
	// 			logger.Fatalf("Could not create tempoary thumbnail directory - %s", err)
	// 		}
	// 	} else {
	thumbnailPath, err = ioutil.TempDir("", "thumbnailcache-")
	if err != nil {
		logger.Fatalf("Could not create tempoary thumbnail directory - %s", err)
	}
	// 	}
	// }

	mux := http.NewServeMux()
	done := make(chan struct{})
	defer close(done)

	var staticHandler http.Handler
	if *staticDir == "" {
		staticHandler = dandler.InternalHandler(logger, fs)
	} else {
		staticHandler = http.FileServer(http.Dir(*staticDir))
	}

	mux.Handle(
		"/", dandler.DirSplitHandler(logger, *imageDir, done,
			dandler.IndexHandler(logger, *imageDir, done, templ),
			dandler.ContentTypeHandler(logger, *imageDir),
		),
	)

	mux.Handle("/static/",
		http.StripPrefix("/static/",
			dandler.SplitHandler(
				http.RedirectHandler("/", 302),
				staticHandler,
			),
		),
	)

	mux.Handle("/thumb/",
		http.StripPrefix("/thumb/",
			dandler.SplitHandler(
				http.RedirectHandler("/", 302),
				dandler.ThumbnailHandler(logger, *thumbWidth, *thumbHeight, *imageDir, thumbnailPath, "jpg"),
			),
		),
	)

	allHandlers := dandler.HeaderHandler("Server", serverBanner, mux)
	allHandlers = dandler.HeaderHandler("Cache-control", "public max-age=2592000", allHandlers)
	allHandlers = handlers.CombinedLoggingHandler(os.Stdout, allHandlers)

	// compress responses
	allHandlers = handlers.CompressHandler(allHandlers)

	logger.Fatal(http.ListenAndServe(*listenAddress, allHandlers))
}
