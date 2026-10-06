package server

import (
	"slices"
	"sync"
)

type DB struct {
	mu   sync.RWMutex
	data []string
}

func (db *DB) Add(s string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data = append(db.data, s)
	slices.Sort(db.data)
}

func (db *DB) Get() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.data
}
