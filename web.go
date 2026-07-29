package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates/index.html
var templateFS embed.FS

var indexTmpl = template.Must(template.ParseFS(templateFS, "templates/index.html"))

type homePage struct {
	AsOf string
	Rows []homePageRow
}

type homePageRow struct {
	Fund      string
	Price     float64
	Change    float64
	HasChange bool
}

func makeHomePage() homePage {
	prices := latestPrices()

	data := homePage{
		Rows: make([]homePageRow, 0, len(prices)),
	}

	if len(prices) > 0 {
		data.AsOf = prices[0].Date.Format(dateLayout)
	}

	for _, price := range prices {
		fund, ok := fundByCode(price.FundCode)
		if !ok {
			// This indicates inconsistent internal data, not bad user input.
			log.Printf("getHome: missing metadata for fund %q", price.FundCode)
			continue
		}

		row := homePageRow{
			Fund:  fund.ShortName,
			Price: price.Price,
		}

		// fundPrices returns the newest records first.
		history := fundPrices(priceQuery{
			Code:  price.FundCode,
			Limit: 2,
		})

		if len(history) == 2 && history[1].Price != 0 {
			previous := history[1].Price
			row.Change = ((price.Price / previous) - 1) * 100
			row.HasChange = true
		}

		data.Rows = append(data.Rows, row)
	}
	return data
}

func getHome(w http.ResponseWriter, r *http.Request) {
	page := makeHomePage()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTmpl.Execute(w, page); err != nil {
		log.Printf("getHome: %v", page)
	}
}
