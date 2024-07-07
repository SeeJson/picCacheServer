package main

import (
	"flag"
	"fmt"
	"github.com/SeeJson/picCacheServer/pictureCache"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	// base HTTP paths.
	apiVersion  = "v1"
	apiBasePath = "/api/" + apiVersion + "/"

	// path to cache.
	cachePath      = apiBasePath + "cache"
	statsPath      = apiBasePath + "stats"
	cacheClearPath = apiBasePath + "cache/clear"
	// server version.
	version                  = "1.0.0"
	maxPicCacheMemory uint64 = 1024 * 4
)

var (
	port      int
	logfile   string
	ver       bool
	maxMemory uint64
	cache     *pictureCache.PictureCache
)

func init() {
	flag.IntVar(&port, "port", 9090, "The port to listen on.")
	flag.StringVar(&logfile, "logfile", "", "Location of the logfile.")
	flag.BoolVar(&ver, "version", false, "Print server version.")
	flag.Uint64("maxMemory", maxMemory, "Print server version.")
	if maxMemory == 0 {
		maxMemory = maxPicCacheMemory
	}
	cache = pictureCache.GetPictureCache()
	cache.Init(maxMemory)
}

func main() {
	flag.Parse()

	if ver {
		fmt.Printf("BigCache HTTP Server v%s", version)
		os.Exit(0)
	}

	var logger *log.Logger

	if logfile == "" {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		f, err := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		logger = log.New(f, "", log.LstdFlags)
	}

	logger.Print("cache initialised.")

	// let the middleware log.
	http.Handle(cacheClearPath, serviceLoader(cacheClearHandler(), requestMetrics(logger)))
	http.Handle(cachePath, serviceLoader(cacheIndexHandler(), requestMetrics(logger)))
	http.Handle(statsPath, serviceLoader(statsIndexHandler(), requestMetrics(logger)))

	logger.Printf("starting server on :%d", port)

	strPort := ":" + strconv.Itoa(port)
	log.Fatal("ListenAndServe: ", http.ListenAndServe(strPort, nil))
}
