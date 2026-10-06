package server

import (
	"slices"
	"strings"
	"sync"
)

type DB struct {
	mu   sync.RWMutex
	data []string
}

func (db *DB) Add(s string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	i, _ := slices.BinarySearch(db.data, s)
	db.data = slices.Insert(db.data, i, s)
}

func (db *DB) Get() string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return strings.Join(db.data, "\n")
}
