package main

import (
	"fmt"
	cache "github.com/caratpine/geecache/src"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom": "640",
	"Jack": "589",
	"Sam": "567",
}

func main() {
	cache.NewGroup("scores", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
	}))
	addr := "localhost:9999"
	pool := cache.NewHTTPPool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, pool))
}