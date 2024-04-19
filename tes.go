package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

type URLQueue struct {
	linkQueue     []string
	visited       []string
	neighborLinks []string
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
	for _, visitedLink := range q.visited {
		if visitedLink == link {
			return true
		}
	}
	return false
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

func BFS(src string, dest string) {
	urlQueue := URLQueue{}

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		colly.AllowURLRevisit(),
	)

	c.OnRequest(func(r *colly.Request) {
		currentLink := r.URL.String()
		if !urlQueue.HasVisited(currentLink) {
			fmt.Println("Visiting", currentLink)
			urlQueue.visited = append(urlQueue.visited, currentLink)
			urlQueue.Enqueue(currentLink)
		}
	})

	c.OnHTML("table.infobox "+"a[href]", func(e *colly.HTMLElement) {
		// c.OnHTML("div#mw-content-text " + "a[href]", func(e *colly.HTMLElement) {
		// c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if urlQueue.HasDequeued(e.Request.URL.String()) {
			neighborLink := e.Attr("href")
			if validLink(neighborLink) {
				urlQueue.neighborLinks = append(urlQueue.neighborLinks, e.Request.AbsoluteURL(neighborLink))
			}
		}
	})

	c.Visit(src)

	distance := 0
	for len(urlQueue.linkQueue) != 0 {
		distance++
		currentLink := urlQueue.Dequeue()
		c.Visit(currentLink)

		currentNeighborLinks := urlQueue.neighborLinks
		urlQueue.neighborLinks = nil

		for _, neighborLink := range currentNeighborLinks {
			if !urlQueue.HasVisited(neighborLink) {
				c.Visit(neighborLink)
				urlQueue.neighborLinks = nil
			} else {
				fmt.Println(neighborLink, "already visited")
			}
		}
	}
}

func main() {
	BFS("https://en.wikipedia.org/wiki/Si_Ronda", "https://en.wikipedia.org/wiki/Jakarta")
}
