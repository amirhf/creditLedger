package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Handle("/metrics", promhttp.Handler())
	// TODO: Kafka consumers + GET /v1/accounts/:id/balance
	addr := ":7104"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("read-model listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
