// Command tspdata serves Thrift Savings Plan fund prices as a JSON API.
// Data is currently seeded in memory and will eventually be stored in SQLite.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	const addr = "localhost:8080"

	log.Printf("listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, newMux()))
}

// Bounds on the price-history endpoint. defaultPriceLimit applies when the
// caller omits limit; maxPriceLimit clamps anything larger.
const (
	defaultPriceLimit = 100
	maxPriceLimit     = 1000
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/funds", getFunds)
	mux.HandleFunc("GET /api/v1/prices/latest", getLatestPrices)
	mux.HandleFunc("GET /api/v1/funds/{code}/prices", getFundPrices)
	return mux
}

// writeJSON sends v as a compact JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	// Headers must be set before the status or body is written.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		// The status is already committed, so only logging remains possible.
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

// getFundPrices responds with one fund's history. Unknown funds return 404;
// known funds without prices return an empty array.
//
// TODO: validate from, to, and limit query parameters and enforce a default limit.
func getFundPrices(w http.ResponseWriter, r *http.Request) {
	// Normalize user input at the HTTP boundary.
	code := strings.ToUpper(r.PathValue("code"))

	if _, ok := fundByCode(code); !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown fund " + code})
		return
	}

	params := r.URL.Query()
	q := priceQuery{Code: code, Limit: defaultPriceLimit}

	if s := params.Get("from"); s != "" {
		d, err := parseDate(s)
		if err != nil {
			badRequest(w, http.StatusBadRequest, "invalid from date "+s)
			return
		}
		q.From = &d
	}

	if s := params.Get("to"); s != "" {
		d, err := parseDate(s)
		if err != nil {
			badRequest(w, http.StatusBadRequest, "invalid to date "+s)
			return
		}
		q.To = &d
	}

	if s := params.Get("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			badRequest(w, http.StatusBadRequest, "invalid limit "+s)
			return
		}
		q.Limit = min(n, maxPriceLimit)
	}

	writeJSON(w, http.StatusOK, fundPrices(q))
}

func badRequest(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
