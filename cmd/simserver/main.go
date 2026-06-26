package main

import (
	"log"
	"net/http"
	"os"

	"github.com/make-smart-products/todo-list/internal/api"
	"github.com/make-smart-products/todo-list/internal/sim"
)

func main() {
	address := os.Getenv("OIL_WORKER_API_ADDR")
	if address == "" {
		address = ":8080"
	}

	server := api.NewServer(sim.NewDefaultSimulation())

	log.Printf("oil-worker simulation API listening on %s (browser client at /)", address)
	if err := http.ListenAndServe(address, server); err != nil {
		log.Fatal(err)
	}
}
