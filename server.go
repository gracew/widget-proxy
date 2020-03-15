package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/handlers"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/model"
	"github.com/gracew/widget-proxy/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// individual API routes
	r := mux.NewRouter()
	r.HandleFunc("/", instrumentedHandler(handlers.CreateHandler, model.OperationTypeCreate.String())).Methods("POST", "OPTIONS")
	r.HandleFunc("/{id}", instrumentedHandler(handlers.ReadHandler, model.OperationTypeRead.String())).Methods("GET", "OPTIONS")
	r.HandleFunc("/", instrumentedHandler(handlers.ListHandler, model.OperationTypeList.String())).Methods("GET", "OPTIONS")
	// TODO(gracew): remove cors later
	r.Use(mux.CORSMethodMiddleware(r))
	http.Handle("/", r)

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("api ready at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	db := pg.Connect(&pg.Options{User: "postgres"})
	defer db.Close()
	s := store.Store{DB: db}
	s.CreateSchema()
}


type handler = func(w http.ResponseWriter, r *http.Request)

func instrumentedHandler(handler handler, label string) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.RequestCounter.WithLabelValues(label).Inc()
		start := time.Now()
		handler(w, r)
		end := time.Now()
		metrics.RequestSummary.WithLabelValues(label).Observe(end.Sub(start).Seconds())
	}
}
