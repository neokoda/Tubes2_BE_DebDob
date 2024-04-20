package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func getWikiArticle(title string) string {
	article := "https://en.wikipedia.org/wiki/" + title
	return article
}

func main() {
	start := time.Now()
	result := IDS("https://en.wikipedia.org/wiki/Nutshell", "https://en.wikipedia.org/wiki/Hamlet")
	elapsed := time.Since(start)

	fmt.Println("Search result:", result)
	fmt.Println("Time taken:", elapsed)
}

func main() {
	r := gin.Default()

	// Endpoint to handle GET requests with query parameters
	r.GET("/", func(c *gin.Context) {
		// Retrieve query parameters
		src := c.Query("src")
		dest := c.Query("dest")
		search := c.Query("search")

		// Example of validation
		if src == "" || dest == "" || search == "" {
			// Return a Bad Request response if name parameter is missing
			c.JSON(http.StatusBadRequest, gin.H{"error": "Source, destination, search is required"})
			return
		}

		// You can then use the retrieved query parameters as needed
		start := time.Now()
		paths := [][]string{}
		if search == "BFS" {
			paths = BFS(getWikiArticle(src), getWikiArticle(dest))
		} else if search == "IDS" {
			paths = IDS(getWikiArticle(src), getWikiArticle(dest))
		}
		elapsed := time.Since(start).Milliseconds()

		// Send a JSON response
		c.JSON(http.StatusOK, gin.H{"paths": paths, "timeTaken (ms)": elapsed})
	})

	r.Run(":" + os.Getenv("PORT")) // Listen and serve on 0.0.0.0:8080
}
