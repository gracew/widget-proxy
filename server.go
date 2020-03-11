package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/handlers"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// register metrics
	prometheus.MustRegister(metrics.RequestSummary)
	prometheus.MustRegister(metrics.CustomLogicSummary)
	prometheus.MustRegister(metrics.CustomLogicErrors)
	prometheus.MustRegister(metrics.DatabaseSummary)
	prometheus.MustRegister(metrics.DatabaseErrors)

	// individual API routes
	r := mux.NewRouter()
	r.HandleFunc("/", instrumentedHandler(handlers.CreateHandler, metrics.RequestSummary, model.OperationTypeCreate.String())).Methods("POST", "OPTIONS")
	r.HandleFunc("/{id}", instrumentedHandler(handlers.ReadHandler, metrics.RequestSummary, model.OperationTypeRead.String())).Methods("GET", "OPTIONS")
	r.HandleFunc("/", instrumentedHandler(handlers.ListHandler, metrics.RequestSummary, model.OperationTypeList.String())).Methods("GET", "OPTIONS")
	// TODO(gracew): remove cors later
	r.Use(mux.CORSMethodMiddleware(r))
	http.Handle("/", r)


	http.Handle("/metrics", promhttp.Handler())

	log.Printf("api ready at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type handler = func(w http.ResponseWriter, r *http.Request)

func instrumentedHandler(handler handler, summary *prometheus.SummaryVec, label string) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler(w, r)
		end := time.Now()
		summary.WithLabelValues(label).Observe(float64(end.Sub(start).Milliseconds()))
	}
}