package cache

import (
	"sync"
	"time"
)

// Store is an interface for cache storage backends.
type Store interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
	Delete(key string)
	Clear()
}

// MemoryStore is an in-memory cache implementation.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]*item
}

type item struct {
	data   []byte
	expiry time.Time
}

// NewMemoryStore creates a new in-memory cache store.
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		items: make(map[string]*item),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Get retrieves a value from the cache.
func (s *MemoryStore) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expiry) {
		return nil, false
	}

	return item.data, true
}

// Set stores a value in the cache with TTL.
func (s *MemoryStore) Set(key string, value []byte, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = &item{
		data:   value,
		expiry: time.Now().Add(ttl),
	}
}

// Delete removes a value from the cache.
func (s *MemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
}

// Clear removes all items from the cache.
func (s *MemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]*item)
}

// cleanup removes expired items periodically.
func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, item := range s.items {
			if now.After(item.expiry) {
				delete(s.items, key)
			}
		}
		s.mu.Unlock()
	}
}
