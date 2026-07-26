package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/api/v1/funds", getFunds)                // -> []Fund
	router.GET("/api/v1/prices/latest", getLatestPrices) // all funds, most recent date
	// router.GET("/api/v1/funds/:code/prices?from=&to=&limit=") // -> []SharePRice

	log.Fatal(router.Run("localhost:8080"))
}

// getFunds responds with the list of all funds as JSON
func getFunds(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, allFunds())
}

func getLatestPrices(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, latestPrices())
}
