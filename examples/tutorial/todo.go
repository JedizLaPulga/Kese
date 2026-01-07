package main

import (
	"sync"
	"time"
)

// Todo represents a todo item
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// TodoStore is our in-memory database
type TodoStore struct {
	mu     sync.RWMutex
	todos  map[int]*Todo
	nextID int
}

// NewTodoStore creates a new todo store
func NewTodoStore() *TodoStore {
	return &TodoStore{
		todos:  make(map[int]*Todo),
		nextID: 1,
	}
}

// Create adds a new todo
func (s *TodoStore) Create(title string) *Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo := &Todo{
		ID:        s.nextID,
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	s.todos[s.nextID] = todo
	s.nextID++

	return todo
}

// GetAll returns all todos
func (s *TodoStore) GetAll() []*Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todos := make([]*Todo, 0, len(s.todos))
	for _, todo := range s.todos {
		todos = append(todos, todo)
	}

	return todos
}

// Get returns a todo by ID
func (s *TodoStore) Get(id int) (*Todo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todo, exists := s.todos[id]
	return todo, exists
}

// Update updates a todo
func (s *TodoStore) Update(id int, title string, completed bool) (*Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, false
	}

	todo.Title = title
	todo.Completed = completed

	return todo, true
}

// Delete removes a todo
func (s *TodoStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.todos[id]
	if !exists {
		return false
	}

	delete(s.todos, id)
	return true
}
