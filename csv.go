package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// parseCSV reads TSP's wide share-price CSV into one SharePrice per populated
// cell. An empty cell means the fund did not exist yet and is skipped; a
// malformed one is an error.
func parseCSV(r io.Reader) ([]SharePrice, error) {
	input := csv.NewReader(r)

	records, err := input.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parseCSV: %w", err)
	}

	if len(records) == 0 {
		return nil, errors.New("parseCSV: no records")
	}

	// get header line of csv
	header := records[0]
	// format codes how we want them
	codes := make([]string, len(header))
	for i, h := range header {
		codes[i] = fundCode(h)
	}

	output := []SharePrice{}
	for _, row := range records[1:] {

		// if date is empty skip it
		if row[0] == "" {
			continue
		}
		date, err := parseDate(row[0])
		if err != nil {
			return nil, fmt.Errorf("parseCSV: %w", err)
		}

		for col := 1; col < len(row); col++ {
			code := codes[col]

			// if price is empty skip it
			if row[col] == "" {
				continue
			}
			price, err := strconv.ParseFloat(row[col], 64)
			if err != nil {
				return nil, fmt.Errorf("parseCSV: %s %s: %w", row[0], header[col], err)
			}

			item := SharePrice{
				FundCode: code,
				Date:     date,
				Price:    price,
			}

			output = append(output, item)
		}
	}

	return output, nil
}

// fundCode converts a CSV header label to a canonical code: "C Fund" becomes
// "C" and "L Income" becomes "LINCOME". Uppercasing runs last so that the
// " Fund" suffix still matches.
func fundCode(header string) string {
	result := strings.ReplaceAll(
		strings.TrimSuffix(header, " Fund"),
		" ",
		"",
	)

	return strings.ToUpper(result)
}
