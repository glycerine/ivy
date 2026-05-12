package main

import (
	"log"
	"net/http"

	"github.com/glycerine/ivy/goivy/svk/server"
)

func main() {
	cfg, err := server.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ivysvk listening on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, srv))
}
