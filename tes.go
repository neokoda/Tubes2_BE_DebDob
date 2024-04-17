package main

import (
	"fmt"
	"strings"
	"github.com/gocolly/colly"
)

type URLQueue struct {
	links []string;
	visited []string;
	currentLink string;
	neighborLinks[] string;
}

func (q* URLQueue) Enqueue(link string) {
	q.links = append(q.links, link);
}

func (q* URLQueue) Dequeue() string {
	if (len(q.links) != 0) {
		link := q.links[0];
		q.links = q.links[1:];
		return link;
	}
	return "";
}

func (q* URLQueue) HasVisited(link string) bool {
    for _, visitedLink := range q.visited {
        if visitedLink == link {
            return true
        }
    }
    return false
}

func validLink(link string) bool {
	invalidPrefixes := []string{"/wiki/Special:", "/wiki/Talk:", "/wiki/User:", "/wiki/Portal:", "/wiki/Wikipedia:", "/wiki/File:", "/wiki/Category:", "/wiki/Help:"};
	for _, prefix := range invalidPrefixes {
		if (strings.HasPrefix(link, prefix)) {
			return false;
		}
	}
	return strings.HasPrefix(link, "/wiki/");
}

func BFS(link string) {
	urlQueue := URLQueue{}

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
	);

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String());
		urlQueue.visited = append(urlQueue.visited, r.URL.String());
		urlQueue.Enqueue(r.URL.String());
	})

	c.OnHTML("div#mw-content-text " + "a[href]", func(e *colly.HTMLElement) {
		neighborLink := e.Attr("href");
		if (validLink(neighborLink)) {
			urlQueue.neighborLinks = append(urlQueue.neighborLinks, e.Request.AbsoluteURL(neighborLink));
		}
	})

	c.Visit(link);

	for (len(urlQueue.links) != 0) {
		urlQueue.Dequeue();

		for _, neighborLink := range urlQueue.neighborLinks {
			if (!urlQueue.HasVisited(neighborLink)) {
				c.Visit(neighborLink);
			}
		}
	}
}

func main() {
	// c := colly.NewCollector(
	// 	colly.AllowedDomains("en.wikipedia.org"),
	// );

	// c.OnHTML("div#mw-content-text " + "a[href]", func(e *colly.HTMLElement) {
	// 	link := e.Attr("href");

	// 	if (validLink(link)) {
	// 		c.Visit(e.Request.AbsoluteURL(link));
	// 	}
	// })

	// c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting", r.URL.String());
	// })

	// c.Visit("https://en.wikipedia.org/wiki/Horse");

	// while (something()) {
	// 	somethingElse();
	// }
	BFS("https://en.wikipedia.org/wiki/Rawer_than_Raw");
}