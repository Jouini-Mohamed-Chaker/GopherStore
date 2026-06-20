package store

import "sync"

type Store struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewStore() *Store {
	return &Store{data: make(map[string][]byte)}
}

func (s *Store) Set(key string, value []byte) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return []byte{}, true
}

func (s *Store) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *Store) Delete(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.data[key]
	delete(s.data, key)
	return []byte{}, exists
}
