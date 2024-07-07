package main

import (
	"github.com/SeeJson/picCacheServer/pictureCache"
	"net/http"
)

// index for stats handle
func statsIndexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCacheStatsHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

// returns the cache's statistics.
func getCacheStatsHandler(w http.ResponseWriter, r *http.Request) {

	cache := pictureCache.GetPictureCache()
	var target []byte
	cache.GetPicCacheInfo().MarshalTo(target)

	// since we're sending a struct, make it easy for consumers to interface.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(target)
}
