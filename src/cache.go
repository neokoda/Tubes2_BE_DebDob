package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

type URLCache struct {
	Links map[string][]string `json:"links"`
}

func NewURLCache() *URLCache {
	return &URLCache{
		Links: make(map[string][]string),
	}
}

func validLinkCache(link string) bool {
	invalidPrefixes := []string{"/wiki/Special:", "/wiki/Talk:", "/wiki/User:", "/wiki/Portal:", "/wiki/Wikipedia:", "/wiki/File:", "/wiki/Category:", "/wiki/Help:", "/wiki/Template:", "/wiki/Template_talk:"}
	for _, prefix := range invalidPrefixes {
		if strings.HasPrefix(link, prefix) {
			return false
		}
	}
	return strings.HasPrefix(link, "/wiki/")
}

func reverseSliceCache(slice []string) {
	for i := 0; i < len(slice)/2; i++ {
		j := len(slice) - i - 1
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func getPathCache(predecessors map[string]string, dest string) []string {
	path := make([]string, 0)
	node := dest

	for node != "" {
		path = append(path, node)
		node = predecessors[node]
	}

	reverseSliceCache(path)
	return path
}

func saveCacheToFile(cache *URLCache, filename string) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func loadCacheFromFile(filename string) (*URLCache, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cache URLCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func BFSCache(src, dest, cacheFilename string) *URLStore {
	// Load the cache if it exists
	cache, err := loadCacheFromFile(cacheFilename)
	if err != nil {
		fmt.Println("Error loading cache:", err)
		cache = NewURLCache() // Start fresh if cache load fails
	}


	urlQueue := NewURLStore()
	var mutex sync.Mutex

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10})

	// Function to check if a link is in the cache


	// Set up the scraping logic
	c.OnRequest(func(r *colly.Request) {
		currentLink := r.URL.String()
		urlQueue.visited.Store(currentLink, true)
		urlQueue.numVisited++
	})

	c.OnHTML("div#mw-content-text a[href]", func(e *colly.HTMLElement) {
		currentLink := e.Request.URL.String()
		neighborLink, _ := url.QueryUnescape(e.Attr("href"))

		if validLinkCache(neighborLink) {
			absoluteURL := e.Request.AbsoluteURL(neighborLink)

			mutex.Lock()
			// Store the link in the cache
			cache.Links[currentLink] = append(cache.Links[currentLink], absoluteURL)

			// Update the predecessors map
			if urlQueue.predecessors[absoluteURL] == "" && absoluteURL != src {
				urlQueue.predecessors[absoluteURL] = currentLink
			}

			urlQueue.neighborLinks = append(urlQueue.neighborLinks, absoluteURL)
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

	urlQueue.resultPath = getPathCache(urlQueue.predecessors, dest)

	// Save the cache to file
	if err := saveCacheToFile(cache, cacheFilename); err != nil {
		fmt.Println("Error saving cache:", err)
	}

	return urlQueue
}
