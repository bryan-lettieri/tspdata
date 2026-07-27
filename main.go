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

// Bounds on the price-history endpoint. defaultPriceLimit applies when the
// caller omits limit; maxPriceLimit clamps anything larger.
const (
	defaultPriceLimit = 100
	maxPriceLimit     = 1000
)

func main() {
	const addr = "localhost:8080"

	log.Printf("listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, newMux()))
}

// newMux registers the API routes.
func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/funds", getFunds)
	mux.HandleFunc("GET /api/v1/funds/{code}/prices", getFundPrices)
	mux.HandleFunc("GET /api/v1/prices/latest", getLatestPrices)
	return mux
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
func getFundPrices(w http.ResponseWriter, r *http.Request) {
	// Normalize user input at the HTTP boundary.
	code := strings.ToUpper(r.PathValue("code"))

	if _, ok := fundByCode(code); !ok {
		writeError(w, http.StatusNotFound, "unknown fund "+code)
		return
	}

	params := r.URL.Query()
	q := priceQuery{Code: code, Limit: defaultPriceLimit}

	if s := params.Get("from"); s != "" {
		d, err := parseDate(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid from date "+s)
			return
		}
		q.From = &d
	}

	if s := params.Get("to"); s != "" {
		d, err := parseDate(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid to date "+s)
			return
		}
		q.To = &d
	}

	// Both bounds present and inverted: almost certainly a typo, not a request
	// for an empty range. Safe to dereference — && short-circuits on nil.
	if q.From != nil && q.To != nil && q.From.Compare(*q.To) > 0 {
		writeError(w, http.StatusBadRequest,
			"from date "+q.From.Format(dateLayout)+" is after to date "+q.To.Format(dateLayout))
		return
	}

	if s := params.Get("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "invalid limit "+s)
			return
		}
		q.Limit = min(n, maxPriceLimit)
	}

	writeJSON(w, http.StatusOK, fundPrices(q))
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

// writeError sends an error in the one shape every endpoint uses.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
