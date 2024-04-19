package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type URLQueue struct {
	predecessors  map[string]string
	linkQueue     []string
	visited       map[string]bool
	neighborLinks []string
}

func NewURLQueue() *URLQueue {
	return &URLQueue{
		predecessors: make(map[string]string),
		visited:      make(map[string]bool),
	}
}

func (q *URLQueue) Enqueue(link string) {
	q.linkQueue = append(q.linkQueue, link)
}

func (q *URLQueue) Dequeue() string {
	if len(q.linkQueue) != 0 {
		link := q.linkQueue[0]
		q.linkQueue = q.linkQueue[1:]
		return link
	}
	return ""
}

func (q *URLQueue) HasVisited(link string) bool {
	return q.visited[link]

}

func (q *URLQueue) HasDequeued(link string) bool {
	for _, queueLink := range q.linkQueue {
		if queueLink == link {
			return false
		}
	}
	return true
}

func validLink(link string) bool {
	invalidPrefixes := []string{"/wiki/Special:", "/wiki/Talk:", "/wiki/User:", "/wiki/Portal:", "/wiki/Wikipedia:", "/wiki/File:", "/wiki/Category:", "/wiki/Help:", "/wiki/Template:"}
	for _, prefix := range invalidPrefixes {
		if strings.HasPrefix(link, prefix) {
			return false
		}
	}
	return strings.HasPrefix(link, "/wiki/")
}

func reverseSlice(slice []string) {
	for i := 0; i < len(slice)/2; i++ {
		j := len(slice) - i - 1
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func getPath(predecessors map[string]string, dest string) []string {
	path := make([]string, 0)
	node := dest

	for node != "" {
		path = append(path, node)
		node = predecessors[node]
	}

	reverseSlice(path)
	return path
}

func BFS(src string, dest string) []string {
	urlQueue := NewURLQueue()

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		colly.AllowURLRevisit(),
	)

	c.OnRequest(func(r *colly.Request) {
		currentLink := r.URL.String()
		if !urlQueue.HasVisited(currentLink) {
			urlQueue.visited[currentLink] = true
			urlQueue.Enqueue(currentLink)
		}
	})

	// c.OnHTML("table.infobox "+"a[href]", func(e *colly.HTMLElement) {
	c.OnHTML("div#mw-content-text "+"a[href]", func(e *colly.HTMLElement) {
		// c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if urlQueue.HasDequeued(e.Request.URL.String()) {
			neighborLink := e.Attr("href")
			if validLink(neighborLink) {
				urlQueue.neighborLinks = append(urlQueue.neighborLinks, e.Request.AbsoluteURL(neighborLink))
			}
		}
	})

	urlQueue.predecessors[src] = ""
	c.Visit(src)

	found := false
	for len(urlQueue.linkQueue) != 0 {
		currentLink := urlQueue.Dequeue()

		c.Visit(currentLink)

		currentNeighborLinks := urlQueue.neighborLinks
		urlQueue.neighborLinks = nil

		for _, neighborLink := range currentNeighborLinks {
			if !urlQueue.HasVisited(neighborLink) {
				urlQueue.predecessors[neighborLink] = currentLink
				if neighborLink == dest {
					found = true
					break
				}
				c.Visit(neighborLink)
				urlQueue.neighborLinks = nil
			}
		}

		if found {
			break
		}
	}

	path := getPath(urlQueue.predecessors, dest)
	return path
}

func main() {
	start := time.Now()
	result := BFS("https://en.wikipedia.org/wiki/Durian", "https://en.wikipedia.org/wiki/Fungicide")
	elapsed := time.Since(start)

	fmt.Println("BFS result:", result)
	fmt.Println("Time taken:", elapsed)
}
