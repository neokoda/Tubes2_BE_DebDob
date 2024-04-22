package main

import "sync"

type URLStore struct {
	predecessors  map[string]string
	linkQueue     []string
	linkStack     []string
	visited       sync.Map
	neighborLinks []string
	resultPath    []string
}

func NewURLStore() *URLStore {
	return &URLStore{
		predecessors: make(map[string]string),
	}
}

func (q *URLStore) Enqueue(link string) {
	q.linkQueue = append(q.linkQueue, link)
}

func (q *URLStore) Dequeue() string {
	if len(q.linkQueue) != 0 {
		link := q.linkQueue[0]
		q.linkQueue = q.linkQueue[1:]
		return link
	}
	return ""
}

func (q *URLStore) Push(url string) {
	q.linkStack = append(q.linkStack, url)
}

func (q *URLStore) Pop() string {
	if len(q.linkStack) != 0 {
		topIndex := len(q.linkStack) - 1
		topURL := q.linkStack[topIndex]
		q.linkStack = q.linkStack[:topIndex]
		return topURL
	}
	return ""
}

func (q *URLStore) HasVisited(link string) bool {
	// return q.visited[link]
	_, ok := q.visited.Load(link)
	return ok
}

func (q *URLStore) HasDequeued(link string) bool {
	for _, queueLink := range q.linkQueue {
		if queueLink == link {
			return false
		}
	}
	return true
}
