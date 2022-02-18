package main

import (
	"fmt"
	"log"
	"mycache"
	"mycache/eliminationstrategy"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *mycache.Group {
	return mycache.NewGroup("scores", eliminationstrategy.LRU, 2<<10, mycache.GetterFunc( // Type conversion, convert func to GetterFunc
		func(key string) ([]byte, error) {
			log.Println("search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, g *mycache.Group) {
	peers := mycache.NewHTTPPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Println("cache server is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, g *mycache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("api server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	apiAddr := "http://localhost:9999"
	addrs := []string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}

	g := createGroup()

	go startAPIServer(apiAddr, g)

	startCacheServer(addrs[0], addrs, g)
}
