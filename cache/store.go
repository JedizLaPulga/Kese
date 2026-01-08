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

// MemoryStore is an in-memory cache implementation with LRU eviction.
type MemoryStore struct {
	mu      sync.RWMutex
	items   map[string]*item
	maxSize int
}

type item struct {
	data       []byte
	expiry     time.Time
	lastAccess time.Time
}

// NewMemoryStore creates a new in-memory cache store with default max size (1000 items).
func NewMemoryStore() *MemoryStore {
	return NewMemoryStoreWithSize(1000)
}

// NewMemoryStoreWithSize creates a cache store with specified max size.
// When max size is reached, least recently used items are evicted.
func NewMemoryStoreWithSize(maxSize int) *MemoryStore {
	if maxSize <= 0 {
		maxSize = 1000 // sensible default
	}
	store := &MemoryStore{
		items:   make(map[string]*item),
		maxSize: maxSize,
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Get retrieves a value from the cache and updates last access time.
func (s *MemoryStore) Get(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	now := time.Now()
	if now.After(item.expiry) {
		return nil, false
	}

	// Update last access for LRU
	item.lastAccess = now

	return item.data, true
}

// Set stores a value in the cache with TTL.
// If cache is full, evicts least recently used item first.
func (s *MemoryStore) Set(key string, value []byte, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Evict LRU item if at capacity and key doesn't exist
	if _, exists := s.items[key]; !exists && len(s.items) >= s.maxSize {
		s.evictLRU()
	}

	now := time.Now()
	s.items[key] = &item{
		data:       value,
		expiry:     now.Add(ttl),
		lastAccess: now,
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

// evictLRU removes the least recently used item from cache.
// Caller must hold the lock.
func (s *MemoryStore) evictLRU() {
	if len(s.items) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range s.items {
		if first || item.lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.lastAccess
			first = false
		}
	}

	delete(s.items, oldestKey)
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
