package main

import (
	"context"

	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

// Check if link is wikipedia article
func validLink(link string) bool {
	invalidPrefixes := []string{"/wiki/Special:", "/wiki/Talk:", "/wiki/User:", 
	"/wiki/Portal:", "/wiki/Wikipedia:", "/wiki/File:", "/wiki/Category:", "/wiki/Help:", 
	"/wiki/Template:", "/wiki/Template_talk:"}
	for _, prefix := range invalidPrefixes {
		if strings.HasPrefix(link, prefix) {
			return false
		}
	}
	return strings.HasPrefix(link, "/wiki/")
}

// Reverses a slice
func reverseSlice(slice []string) {
	for i := 0; i < len(slice)/2; i++ {
		j := len(slice) - i - 1
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Checks whether or not a string is in a slice
func stringInSlice(str string, list []string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}

// Gets full path from map of predecessors
func getPath(predecessors map[string]string, dest string) []string {
	path := make([]string, 0)
	node := dest

	for node != "" { // source has no predecessor
		path = append(path, node)
		node = predecessors[node]
	}

	reverseSlice(path)
	return path
}

// Gets all paths from map of predecessors
func getPaths(predecessors map[string][]string, src string, dest string) [][]string {
	var paths [][]string
	var resultPaths [][]string

	found := false
	paths = append(paths, []string{dest})
	// builds paths from destination node, every iteration adds the length of the path by one
	for !found {
		currentPaths := paths
		paths = nil
		for _, path := range currentPaths {
			for _, pred := range predecessors[path[len(path)-1]] {
				if pred == src {
					found = true
				}
				newPath := append(path, pred)
				paths = append(paths, newPath)
			}
		}
	}

	// reverses and adds paths that have src as the final element
	for _, path := range paths {
		if path[len(path)-1] == src {
			reverseSlice(path)
			resultPaths = append(resultPaths, path)
		}
	}

	return resultPaths
}

// Multi solution BFS
func BFSMulti(src string, dest string, cache *URLCache) *URLStore {
	// initialize url store and mutex
	urlQueue := NewURLStore()

	var mutex sync.Mutex

	// set up colly config and On<...> functions
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
		

		
		if validLink(neighborLink) && !urlQueue.HasVisited(neighborLink) {
			mutex.Lock()
			// append to existing predecessor map if already exists, creates new one if not
			if _, ok := urlQueue.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)]; ok && !stringInSlice(currentLink, urlQueue.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)]) {
				urlQueue.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)] = append(urlQueue.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)], currentLink)
			} else if !ok {
				urlQueue.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)] = []string{currentLink}
			}
			urlQueue.neighborLinks = append(urlQueue.neighborLinks, e.Request.AbsoluteURL(neighborLink))
			mutex.Unlock()
		}
	})

	c.Visit(src)

	found := false
	for !found { // one iteration is one addition to depth
		currentNeighborLinks := urlQueue.neighborLinks
		urlQueue.neighborLinks = nil // copy neighborLinks and clear

		for _, neighborLink := range currentNeighborLinks {
			if !urlQueue.HasVisited(neighborLink) {
				_, ok := cache.Links[neighborLink]
				if ok { // present in cache
					
					go func(neighborLink string) {
						urlQueue.visited.Store(neighborLink, true)

						mutex.Lock()
						for _, neighborLink2 := range cache.Links[neighborLink] {
							if _, ok2 := urlQueue.predecessorsMulti[neighborLink2]; ok && !stringInSlice(neighborLink, urlQueue.predecessorsMulti[neighborLink2]) {
								urlQueue.predecessorsMulti[neighborLink2] = append(urlQueue.predecessorsMulti[neighborLink2], neighborLink)
							} else if !ok2 {
								urlQueue.predecessorsMulti[neighborLink2] = []string{neighborLink}
							}
						}
						urlQueue.neighborLinks = append(urlQueue.neighborLinks, cache.Links[neighborLink]...)
						mutex.Unlock()

						urlQueue.numVisited++
					}(neighborLink)
				} else { // manually scrape
					c.Visit(neighborLink)
					
				}
			}
			if neighborLink == dest {
		
				found = true
				break
			}
		}

		c.Wait()
	}

	// get paths from predecessorsMulti
	urlQueue.resultPaths = getPaths(urlQueue.predecessorsMulti, src, dest)

	return urlQueue
}

func BFS(src string, dest string, cache *URLCache) *URLStore {
	// initialize url store and mutex
	urlQueue := NewURLStore()

	// set up colly config and On<...> functions
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
	for !found { // one iteration is one addition to depth
		currentNeighborLinks := urlQueue.neighborLinks
		urlQueue.neighborLinks = nil // copy neighborlinks and clear

		for _, neighborLink := range currentNeighborLinks {
			if !urlQueue.HasVisited(neighborLink) {
				_, ok := cache.Links[neighborLink]
				if ok { // present in cache
					go func(neighborLink string) {
						urlQueue.visited.Store(neighborLink, true)

						mutex.Lock()
						for _, neighborLink2 := range cache.Links[neighborLink] {
							neighborLink2, _ := url.QueryUnescape(neighborLink2)
							if urlQueue.predecessors[neighborLink2] == "" && neighborLink2 != src {
								urlQueue.predecessors[neighborLink2] = neighborLink
							}
						}
						urlQueue.neighborLinks = append(urlQueue.neighborLinks, cache.Links[neighborLink]...)
						mutex.Unlock()

						urlQueue.numVisited++
					}(neighborLink)
				} else { // manually scrape
					c.Visit(neighborLink)
				}
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

	// get paths from predecessors
	mutex.Lock()
	urlQueue.resultPath = getPath(urlQueue.predecessors, dest)
	mutex.Unlock()

	return urlQueue
}

func DLS(src string, dest string, maxDepth int) *URLStore {
	// initialize new URLStore
	urlStore := NewURLStore()

	// concurrency tools
	var mutex sync.Mutex
	timer := time.NewTimer(2 * time.Second) // Stop DLS if no visits are being made after two seconds
	noVisits := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// initialize colly config
	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		colly.MaxDepth(maxDepth),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 20})

	// DLS function to immediately stop execution on first finding of destination
	scraper := func() {
		defer c.Wait()

		c.OnRequest(func(r *colly.Request) {
			timer.Reset(2 * time.Second)
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
			neighborLink, _ := url.QueryUnescape(e.Attr("href"))
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

	// execute scraper
	go scraper()

	// timer watcher function
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

func DLSMulti(src string, dest string, maxDepth int) *URLStore {
	// initialize new URLStore
	urlStore := NewURLStore()
	found := false

	// concurrency tools
	var mutex sync.Mutex

	// initialize colly config
	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		colly.MaxDepth(maxDepth),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10})

	c.OnRequest(func(r *colly.Request) {
		currentLink := r.URL.String()
		urlStore.visited.Store(currentLink, true)
		urlStore.numVisited++
		if currentLink == dest {
			found = true
		}
	})

	c.OnHTML("table.infobox "+"a[href]", func(e *colly.HTMLElement) {
		currentLink := e.Request.URL.String()
		neighborLink, _ := url.QueryUnescape(e.Attr("href"))
		if validLink(neighborLink) {
			mutex.Lock()
			if _, ok := urlStore.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)]; ok && !stringInSlice(currentLink, urlStore.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)]) {
				urlStore.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)] = append(urlStore.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)], currentLink)
			} else if !ok {
				urlStore.predecessorsMulti[e.Request.AbsoluteURL(neighborLink)] = []string{currentLink}
			}
			mutex.Unlock()

			e.Request.Visit(e.Request.AbsoluteURL(neighborLink))
		}
	})

	c.Visit(src)

	c.Wait() // wait until all nodes are done being visited

	if found {
		urlStore.resultPaths = getPaths(urlStore.predecessorsMulti, src, dest)
	}

	return urlStore
}

func IDS(src string, dest string) *URLStore {
	depth := 1
	urlStore := NewURLStore()
	for {
		urlStore = DLS(src, dest, depth)
		if len(urlStore.resultPath) > 0 { // solution not found, increase depth
			break
		}
		depth++
	}
	return urlStore
}

func IDSMulti(src string, dest string) *URLStore {
	depth := 1
	urlStore := NewURLStore()
	for {
		urlStore = DLSMulti(src, dest, depth)
		if len(urlStore.resultPaths) > 0 { // solution not found, increase depth
			break
		}
		depth++
	}
	return urlStore
}
