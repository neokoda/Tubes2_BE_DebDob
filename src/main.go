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
// 	urlStore := IDS("https://en.wikipedia.org/wiki/Rawer_than_Raw", "https://en.wikipedia.org/wiki/Porcupine_Meat")
// 	elapsed := time.Since(start)

// 	fmt.Println("Search result:", urlStore.resultPath)
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

		if src == "" || dest == "" || search == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Source, destination, search is required"})
			return
		} else if search != "BFS" && search != "IDS" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search algorithmn"})
			return
		}

		start := time.Now()
		urlStore := NewURLStore()
		if search == "BFS" {
			urlStore = BFS(getWikiArticle(src), getWikiArticle(dest))
		} else if search == "IDS" {
			urlStore = IDS(getWikiArticle(src), getWikiArticle(dest))
		}
		elapsed := time.Since(start).Milliseconds()
		
		c.JSON(http.StatusOK, gin.H{"paths": [][]string{urlStore.resultPath}, "visited": urlStore.numVisited, "timeTaken": elapsed})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default to port 8000 if PORT is not set
	}

	r.Run(":" + port)
}
