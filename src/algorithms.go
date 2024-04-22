package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

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

func BFS(src string, dest string) [][]string {
	urlQueue := NewURLStore()

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		colly.AllowURLRevisit(),
	)

	c.OnRequest(func(r *colly.Request) {
		currentLink := r.URL.String()
		if !urlQueue.HasVisited(currentLink) {
			fmt.Println("Visiting", currentLink)
			// urlQueue.visited[currentLink] = true
			urlQueue.Enqueue(currentLink)
		}
	})

	c.OnHTML("table.infobox "+"a[href]", func(e *colly.HTMLElement) {
		// c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// c.OnHTML("div#mw-content-text "+"a[href]", func(e *colly.HTMLElement) {
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

	urlQueue.resultPath = getPath(urlQueue.predecessors, dest)

	return [][]string{urlQueue.resultPath}
}

func BFS2(src string, dest string) [][]string {
	urlQueue := NewURLStore()

	var mutex sync.Mutex
	var wg sync.WaitGroup

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		// colly.AllowedDomains("informatika.stei.itb.ac.id"),
	)

	c.OnRequest(func(r *colly.Request) {
		currentLink := r.URL.String()
		fmt.Println("Visiting", currentLink)
		urlQueue.visited.Store(currentLink, true)
	})

	// c.OnHTML("table.infobox "+"a[href]", func(e *colly.HTMLElement) {
	// c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	c.OnHTML("div#mw-content-text "+"a[href]", func(e *colly.HTMLElement) {
		currentLink := e.Request.URL.String()
		neighborLink := e.Attr("href")
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

	maxConcurrency := 200
	semaphore := make(chan struct{}, maxConcurrency)

	found := false
	for !found {
		currentNeighborLinks := urlQueue.neighborLinks
		urlQueue.neighborLinks = nil

		for _, neighborLink := range currentNeighborLinks {
			semaphore <- struct{}{}
			wg.Add(1)
			go func(link string) {
				defer wg.Done()
				defer func() { <-semaphore }()

				if !urlQueue.HasVisited(link) {
					c.Visit(link)
				}
			}(neighborLink)

			if neighborLink == dest {
				found = true
				break
			}
		}

		wg.Wait()
	}

	urlQueue.resultPath = getPath(urlQueue.predecessors, dest)

	return [][]string{urlQueue.resultPath}
}

func DLS(src string, dest string, maxDepth int) [][]string {
	urlStore := NewURLStore()

	var mutex sync.Mutex
	var wg sync.WaitGroup

	maxConcurrency := 200
	semaphore := make(chan struct{}, maxConcurrency)

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		// colly.AllowedDomains("informatika.stei.itb.ac.id"),
		colly.MaxDepth(maxDepth),
	)

	c.OnRequest(func(r *colly.Request) {
		currentLink := r.URL.String()
		fmt.Println("Visiting", currentLink)
		urlStore.visited.Store(currentLink, true)
		if currentLink == dest {
			urlStore.resultPath = getPath(urlStore.predecessors, dest)
		}
	})

	// c.OnHTML("table.infobox "+"a[href]", func(e *colly.HTMLElement) {
	// c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	c.OnHTML("div#mw-content-text "+"a[href]", func(e *colly.HTMLElement) {
		currentLink := e.Request.URL.String()
		neighborLink := e.Attr("href")
		if validLink(neighborLink) {
			mutex.Lock()
			if urlStore.predecessors[e.Request.AbsoluteURL(neighborLink)] == "" && e.Request.AbsoluteURL(neighborLink) != src {
				urlStore.predecessors[e.Request.AbsoluteURL(neighborLink)] = currentLink
			}
			mutex.Unlock()

			semaphore <- struct{}{}
			wg.Add(1)
			go func(link string) {
				defer wg.Done()
				defer func() { <-semaphore }()

				e.Request.Visit(link)
			}(e.Request.AbsoluteURL(neighborLink))
		}
	})

	c.Visit(src)

	wg.Wait()

	return [][]string{urlStore.resultPath}
}

func IDS(src string, dest string) [][]string {
	depth := 1
	paths := [][]string{}
	for {
		paths = DLS(src, dest, depth)
		if len(paths[0]) > 0 {
			break
		}
		depth++
	}
	return paths
}
