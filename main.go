// Command tspdata serves Thrift Savings Plan fund share prices as a small JSON API.
//
// TSP publishes a share price CSV each business night. This program exposes the
// most recent prices plus historical data for every fund. Storage is currently a
// pair of in-memory seed slices (see store.go); SQLite replaces them once the
// nightly ingest job exists.
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe("localhost:8080", newMux()))
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/funds", getFunds)
	mux.HandleFunc("GET /api/v1/prices/latest", getLatestPrices)
	mux.HandleFunc("GET /api/v1/funds/{code}/prices", getFundPrices)
	return mux
}

// writeJSON encodes v as indented JSON alongside the given status code.
//
// The order of the first two calls is load-bearing: every header must be set
// before WriteHeader, and WriteHeader before any body is written. A header set
// afterwards is discarded silently — no error, no panic, it simply never
// reaches the client.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	if err := enc.Encode(v); err != nil {
		// The status line and headers are already on the wire, so the response
		// can no longer be corrected to report this. Logging is all that is left.
		log.Printf("writeJSON: %v", err)
	}
}

// getFunds responds with metadata for every fund.
func getFunds(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, allFunds())
}

// getLatestPrices responds with one price per fund for the most recent date on record.
func getLatestPrices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, latestPrices())
}

// getFundPrices responds with the price history for a single fund.
//
// TODO: unimplemented — currently returns 200 with an empty body, because a
// handler that writes nothing still gets a default 200 from net/http.
//
// The fund code comes from the path wildcard via r.PathValue("code"); the
// from/to/limit filters come from r.URL.Query(), since ServeMux patterns match
// method and path only. An unknown code should be a 404, and the response needs
// a bounded default limit — the full history is roughly 90k rows.
func getFundPrices(w http.ResponseWriter, r *http.Request) {
}
