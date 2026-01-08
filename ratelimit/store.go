package ratelimit

import (
	"sync"
	"time"
)

// Store is an interface for rate limit storage backends.
type Store interface {
	// Get returns the current count for the given key
	Get(key string) (int, error)

	// Increment increments the count for the given key and returns the new count
	Increment(key string, window time.Duration) (int, error)

	// Reset resets the count for the given key
	Reset(key string) error
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]*entry
}

type entry struct {
	count  int
	expiry time.Time
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		data: make(map[string]*entry),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Get returns the current count for the given key.
func (s *MemoryStore) Get(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if e, exists := s.data[key]; exists {
		if time.Now().Before(e.expiry) {
			return e.count, nil
		}
	}

	return 0, nil
}

// Increment increments the count for the given key.
func (s *MemoryStore) Increment(key string, window time.Duration) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	if e, exists := s.data[key]; exists {
		if now.Before(e.expiry) {
			e.count++
			return e.count, nil
		}
	}

	// Create new entry
	s.data[key] = &entry{
		count:  1,
		expiry: now.Add(window),
	}

	return 1, nil
}

// Reset resets the count for the given key.
func (s *MemoryStore) Reset(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}

// cleanup removes expired entries every minute.
func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, e := range s.data {
			if now.After(e.expiry) {
				delete(s.data, key)
			}
		}
		s.mu.Unlock()
	}
}
