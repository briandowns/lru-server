package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/golang-lru"
	"github.com/tidwall/modern-server"
)

const defaultSize = 1000000

func main() {
	var size int
	var cache *lru.Cache
	var err error
	opts := &server.Options{
		Version: "0.0.1",
		Name:    "lru-server",
		Flags:   func() { flag.IntVar(&size, "s", defaultSize, "") },
		FlagsParsed: func() {
			cache, err = lru.New(size)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
		Usage: func(s string) string {
			return strings.Replace(s, "{{USAGE}}",
				fmt.Sprintf("  -s size      : size of lru (default: %d)\n", defaultSize), -1)
		},
	}
	server.Main(
		func(w http.ResponseWriter, r *http.Request) {
			key := strings.Split(r.URL.Path, "/")[1]
			switch r.Method {
			case http.MethodGet:
				if val, ok := cache.Get(key); !ok {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.Write([]byte(val.(string)))
				}
			case http.MethodDelete:
				cache.Remove(key)
			case http.MethodPut, http.MethodPost:
				if data, err := ioutil.ReadAll(r.Body); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					cache.Add(key, string(data))
				}
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}, opts,
	)
}
