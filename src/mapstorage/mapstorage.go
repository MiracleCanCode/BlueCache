package mapstorage

import "sync"

type Storage struct {
	storage map[string]byte
	mu      sync.Mutex
}

func (s *Storage) Add() {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *Storage) Get() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}
