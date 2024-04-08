package main

import (
	"log"
	"net/http"
	_ "net/http/pprof" //nolint: gosec
	"os"
	"time"
)

// nolint: gochecknoinits
func init() {
	if os.Getenv("TOPAZ_DEBUG") == "1" {
		go func() {
			srv := &http.Server{
				Addr:              "localhost:6060",
				ReadTimeout:       5 * time.Second,
				ReadHeaderTimeout: 5 * time.Second,
				WriteTimeout:      5 * time.Second,
				IdleTimeout:       30 * time.Second,
			}
			log.Println(srv.ListenAndServe())
		}()
	}
}
