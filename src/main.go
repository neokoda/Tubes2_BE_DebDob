package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getWikiArticle(title string) string {
	article := "https://en.wikipedia.org/wiki/" + title
	return article
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	// endpoint is in the form of a GET request with params src, dest, search, and resultAmount
	r.GET("/", func(c *gin.Context) {
		src := c.Query("src")
		dest := c.Query("dest")
		search := c.Query("search")
		resultAmount := c.Query("resultAmount")

		// check for missing or invalid parameters
		if src == "" || dest == "" || search == "" || resultAmount == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Source, destination, search, resultAmount is required"})
			return
		} else if search != "BFS" && search != "IDS" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search algorithmn"})
			return
		} else if resultAmount != "Single" && resultAmount != "Multi" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resultAmount value"})
		}

		if (src == dest){
			c.JSON(http.StatusOK, gin.H{"paths": [][]string{{src}}, "visited": 1, "timeTaken": 0})
			return
		}

		if (src == dest){
			c.JSON(http.StatusOK, gin.H{"paths": [][]string{{src}}, "visited": 1, "timeTaken": 0})
			return
		}
		// perform search and calculate time taken
		start := time.Now()
		urlStore := NewURLStore()
		if search == "BFS" {
			if resultAmount == "Single" {
				urlStore = BFS(getWikiArticle(src), getWikiArticle(dest), "cache.json")
			} else {
				urlStore = BFSMulti(getWikiArticle(src), getWikiArticle(dest), "cache.json")
			}
		} else {
			if resultAmount == "Single" {
				urlStore = IDS(getWikiArticle(src), getWikiArticle(dest))
			} else {
				urlStore = IDSMulti(getWikiArticle(src), getWikiArticle(dest))
			}
		}
		elapsed := time.Since(start).Milliseconds()

		// return json
		if resultAmount == "Single" {
			c.JSON(http.StatusOK, gin.H{"paths": [][]string{urlStore.resultPath}, "visited": urlStore.numVisited, "timeTaken": elapsed})
		} else {
			c.JSON(http.StatusOK, gin.H{"paths": urlStore.resultPaths, "visited": urlStore.numVisited, "timeTaken": elapsed})
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default to port 8000 if PORT is not set
	}

	// run file
	r.Run(":" + port)
}
