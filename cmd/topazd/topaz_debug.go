package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func init() {
	if os.Getenv("TOPAZ_DEBUG") == "1" {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
}
