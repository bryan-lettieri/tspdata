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
	"strings"
)

func main() {
	const addr = "localhost:8080"

	// Logged before ListenAndServe, which blocks until the server stops. Note
	// this prints even if the bind then fails — the Fatal below is what tells
	// you it did.
	log.Printf("listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, newMux()))
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/funds", getFunds)
	mux.HandleFunc("GET /api/v1/prices/latest", getLatestPrices)
	mux.HandleFunc("GET /api/v1/funds/{code}/prices", getFundPrices)
	return mux
}

// writeJSON encodes v as compact JSON alongside the given status code.
//
// The order of the first two calls is load-bearing: every header must be set
// before WriteHeader, and WriteHeader before any body is written. A header set
// afterwards is discarded silently — no error, no panic, it simply never
// reaches the client.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
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

// getFundPrices responds with the price history for a single fund, or 404 when
// the code is not one we know.
//
// The path value is uppercased here rather than inside the store: a URL is
// untrusted input, so normalizing at the boundary lets everything downstream
// assume a canonical code. Note the 404 branch returns — falling through would
// write a second body and status.
//
// A missing fund is a 404; a known fund with no matching rows is a 200 with an
// empty array. Those are different answers and callers need to tell them apart.
//
// TODO: from/to/limit are still ignored. They come from r.URL.Query(), since
// ServeMux patterns match method and path only. The response also needs a
// stable sort and a bounded default limit — one fund's full history is roughly
// 6k rows, about 300 KB of compact JSON.
func getFundPrices(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(r.PathValue("code"))

	if _, ok := fundByCode(code); !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown fund " + code})
		return
	}

	writeJSON(w, http.StatusOK, fundPrices(code))
}
