package main

import "sync"

// Data structure for wiki traversal.
type URLStore struct {
	predecessors      map[string]string
	predecessorsMulti map[string][]string
	visited           sync.Map
	numVisited        int
	neighborLinks     []string
	resultPath        []string
	resultPaths       [][]string
}

// Create a new URL Store
func NewURLStore() *URLStore {
	return &URLStore{
		predecessors:      make(map[string]string),
		predecessorsMulti: make(map[string][]string),
	}
}

// Check if article has been visited
func (q *URLStore) HasVisited(link string) bool {
	_, ok := q.visited.Load(link)
	return ok
}
