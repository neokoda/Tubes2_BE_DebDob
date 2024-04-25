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

// func main() {
// 	start := time.Now()
// 	urlStore := BFSMulti("https://en.wikipedia.org/wiki/Bandung_Institute_of_Technology", "https://en.wikipedia.org/wiki/Joko_Widodo")
// 	elapsed := time.Since(start)

// 	fmt.Println("Search result:", urlStore.resultPaths, len(urlStore.resultPaths))
// 	fmt.Println("Articles visited:", urlStore.numVisited)
// 	fmt.Println("Time taken:", elapsed)
// }

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		src := c.Query("src")
		dest := c.Query("dest")
		search := c.Query("search")
		resultAmount := c.Query("resultAmount")

		if src == "" || dest == "" || search == "" || resultAmount == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Source, destination, search, resultAmount is required"})
			return
		} else if search != "BFS" && search != "IDS" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search algorithmn"})
			return
		} else if resultAmount != "Single" && resultAmount != "Multi" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resultAmount value"})
		}

		start := time.Now()
		urlStore := NewURLStore()
		if search == "BFS" {
			if resultAmount == "Single" {
				urlStore = BFS(getWikiArticle(src), getWikiArticle(dest))
			} else {
				urlStore = BFSMulti(getWikiArticle(src), getWikiArticle(dest))
			}
		} else {
			if resultAmount == "Single" {
				urlStore = IDS(getWikiArticle(src), getWikiArticle(dest))
			} else {
				urlStore = IDS(getWikiArticle(src), getWikiArticle(dest))
			}
		}
		elapsed := time.Since(start).Milliseconds()

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

	r.Run(":" + port)
}
