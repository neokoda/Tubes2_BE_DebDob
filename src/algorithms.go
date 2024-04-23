package main

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

func validLink(link string) bool {
	invalidPrefixes := []string{"/wiki/Special:", "/wiki/Talk:", "/wiki/User:", "/wiki/Portal:", "/wiki/Wikipedia:", "/wiki/File:", "/wiki/Category:", "/wiki/Help:", "/wiki/Template:", "/wiki/Template_talk:"}
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

func BFS(src string, dest string) *URLStore {
	urlQueue := NewURLStore()

	var mutex sync.Mutex

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10})

	c.OnRequest(func(r *colly.Request) {
		currentLink := r.URL.String()
		urlQueue.visited.Store(currentLink, true)
		urlQueue.numVisited++
	})

	c.OnHTML("div#mw-content-text "+"a[href]", func(e *colly.HTMLElement) {
		currentLink := e.Request.URL.String()
		neighborLink, _ := url.QueryUnescape(e.Attr("href"))
		if validLink(neighborLink) {
			mutex.Lock()
			if urlQueue.predecessors[e.Request.AbsoluteURL(neighborLink)] == "" && e.Request.AbsoluteURL(neighborLink) != src {
				urlQueue.predecessors[e.Request.AbsoluteURL(neighborLink)] = currentLink
			}
			urlQueue.neighborLinks = append(urlQueue.neighborLinks, e.Request.AbsoluteURL(neighborLink))
			mutex.Unlock()
		}
	})

	urlQueue.predecessors[src] = ""
	c.Visit(src)

	found := false
	for !found {
		currentNeighborLinks := urlQueue.neighborLinks
		urlQueue.neighborLinks = nil

		for _, neighborLink := range currentNeighborLinks {
			if !urlQueue.HasVisited(neighborLink) {
				c.Visit(neighborLink)
			}
			if neighborLink == dest {
				found = true
				break
			}
		}

		if !found {
			c.Wait()
		}
	}

	urlQueue.resultPath = getPath(urlQueue.predecessors, dest)

	return urlQueue
}

func DLS(src string, dest string, maxDepth int) *URLStore {
	urlStore := NewURLStore()

	var mutex sync.Mutex
	timer := time.NewTimer(3 * time.Second)
	noVisits := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		colly.MaxDepth(maxDepth),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10})

	scraper := func() {
		defer c.Wait()

		c.OnRequest(func(r *colly.Request) {
			timer.Reset(3 * time.Second)
			currentLink := r.URL.String()
			urlStore.visited.Store(currentLink, true)
			urlStore.numVisited++
			if currentLink == dest {
				urlStore.resultPath = getPath(urlStore.predecessors, dest)
				cancel()
			}
		})

		c.OnHTML("div#mw-content-text "+"a[href]", func(e *colly.HTMLElement) {
			currentLink := e.Request.URL.String()
			neighborLink := e.Attr("href")
			if validLink(neighborLink) {
				mutex.Lock()
				if urlStore.predecessors[e.Request.AbsoluteURL(neighborLink)] == "" && e.Request.AbsoluteURL(neighborLink) != src {
					urlStore.predecessors[e.Request.AbsoluteURL(neighborLink)] = currentLink
				}
				mutex.Unlock()

				e.Request.Visit(e.Request.AbsoluteURL(neighborLink))
			}
		})

		c.Visit(src)
	}

	go scraper()

	go func() {
		<-timer.C
		noVisits <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return urlStore
	case <-noVisits:
		return urlStore
	}
}

func IDS(src string, dest string) *URLStore {
	depth := 1
	urlStore := NewURLStore()
	for {
		urlStore = DLS(src, dest, depth)
		if len(urlStore.resultPath) > 0 {
			break
		}
		depth++
	}
	return urlStore
}
