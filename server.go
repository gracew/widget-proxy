package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/handlers"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// individual API routes
	r := mux.NewRouter()
	r.HandleFunc("/apis/{apiID}/{env}", handlers.CreateHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/apis/{apiID}/{env}/{id}", handlers.ReadHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/apis/{apiID}/{env}", handlers.ListHandler).Methods("GET", "OPTIONS")
	// TODO(gracew): remove cors later
	r.Use(mux.CORSMethodMiddleware(r))
	http.Handle("/", r)

	log.Printf("api ready at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
