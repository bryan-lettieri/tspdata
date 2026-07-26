package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/funds", getFunds)
	mux.HandleFunc("GET /api/v1/prices/latest", getLatestPrices)
	mux.HandleFunc("GET /api/v1/funds/{code}/prices", getFundPrices)

	log.Fatal(http.ListenAndServe("localhost:8080", mux))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	if err := enc.Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func getFunds(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, allFunds())
}

func getLatestPrices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, latestPrices())
}

func getFundPrices(w http.ResponseWriter, r *http.Request) {
	//todo
}
