package main

import (
	"fmt"
	"github.com/SeeJson/picCacheServer/pictureCache"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func cacheIndexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCacheHandler(w, r)
		case http.MethodPost:
			postCacheHandler(w, r)
		case http.MethodDelete:
			deleteCacheHandler(w, r)
		}
	})
}

func cacheClearHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clearCache(w, r)
	})
}

func clearCache(w http.ResponseWriter, r *http.Request) {
	cache := pictureCache.GetPictureCache()
	cache.ClearPicture()
	log.Println("cache is successfully cleared")
	w.WriteHeader(http.StatusOK)
}

// handles get requests.
func getCacheHandler(w http.ResponseWriter, r *http.Request) {
	//target := r.URL.Path[len(cachePath):]
	//if target == "" {
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte("can't get a key if there is no key."))
	//	log.Print("empty request.")
	//	return
	//}
	seq := r.URL.Query().Get("seq")
	key := r.URL.Query().Get("key")

	no, _ := strconv.Atoi(seq)
	entry, err := cache.GetPicture(uint64(no), key)
	if err != nil {
		errMsg := (err).Error()
		if strings.Contains(errMsg, "not found") {
			log.Print(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(entry.Bytes())
}

type ReqPutPicCache struct {
	Url string `json:"url"`
}

func postCacheHandler(w http.ResponseWriter, r *http.Request) {

	entry, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	isSave, seq, _, pictureKey := cache.SavePicture(entry)
	if !isSave {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/cache?seq=%d&key=%s", port, seq, pictureKey)

	//respMap := make(map[string]string)
	//respMap["url"] = url
	//respByte, _ := json.Marshal(respMap)
	w.Write([]byte(url))
}

// delete cache objects.
func deleteCacheHandler(w http.ResponseWriter, r *http.Request) {
	//target := r.URL.Path[len(cachePath):]

	//if err := cache.Delete(target); err != nil {
	//	if strings.Contains((err).Error(), "not found") {
	//		w.WriteHeader(http.StatusNotFound)
	//		log.Printf("%s not found.", target)
	//		return
	//	}
	//	w.WriteHeader(http.StatusInternalServerError)
	//	log.Printf("internal cache error: %s", err)
	//}
	// this is what the RFC says to use when calling DELETE.
	w.WriteHeader(http.StatusOK)
}

func downLoad(url string) (data []byte, err error) {
	v, err := http.Get(url)
	if err != nil {
		fmt.Printf("Http get [%v] failed! %v", url, err)
		return data, err
	}
	defer v.Body.Close()

	data, err = io.ReadAll(v.Body)
	if err != nil {
		fmt.Printf("Read http response failed! %v", err)
		return data, err
	}

	return data, nil
}
