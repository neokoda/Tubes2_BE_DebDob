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
// 	result := IDS("https://en.wikipedia.org/wiki/Bandung_Institute_of_Technology", "https://en.wikipedia.org/wiki/China")
// 	elapsed := time.Since(start)

// 	fmt.Println("Search result:", result)
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
		}

		start := time.Now()
		paths := [][]string{}
		if search == "BFS" {
			paths = BFS2(getWikiArticle(src), getWikiArticle(dest))
		} else if search == "IDS" {
			paths = IDS(getWikiArticle(src), getWikiArticle(dest))
		}
		elapsed := time.Since(start).Milliseconds()

		c.JSON(http.StatusOK, gin.H{"paths": paths, "timeTaken (ms)": elapsed})
	})

	r.Run(":" + os.Getenv("PORT"))
}
