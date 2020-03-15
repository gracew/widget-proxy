package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/config"
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

	db := pg.Connect(&pg.Options{User: "postgres", Addr: config.PostgresAddress})
	defer db.Close()
	s := store.InstrumentedStore{Delegate: store.PgStore{DB: db}}
	s.CreateSchema()

	r := mux.NewRouter()
	h := handlers.Handlers{Store: s}
	r.HandleFunc("/", instrumentedHandler(h.CreateHandler, model.OperationTypeCreate.String())).Methods("POST", "OPTIONS")
	r.HandleFunc("/{id}", instrumentedHandler(h.ReadHandler, model.OperationTypeRead.String())).Methods("GET", "OPTIONS")
	r.HandleFunc("/", instrumentedHandler(h.ListHandler, model.OperationTypeList.String())).Methods("GET", "OPTIONS")
	// TODO(gracew): remove cors later
	r.Use(mux.CORSMethodMiddleware(r))
	http.Handle("/", r)

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("api ready at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
