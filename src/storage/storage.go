package storage

import "fmt"

type aofInterface interface {
	AppendOperation(method string, key string, value string) error
}

type Storage struct {
	store      map[string]string
	aofManager aofInterface
}

func New(aof aofInterface) *Storage {
	return &Storage{
		store:      make(map[string]string),
		aofManager: aof,
	}
}

func (s *Storage) Get(key string) (string, error) {
	data, ok := s.store[key]
	if !ok {
		return "", fmt.Errorf("Get: failed get data by key")
	}

	return data, nil
}

func (s *Storage) Set(key string, value string) error {
	_, ok := s.store[key]
	if ok {
		return fmt.Errorf("Set: failed set data to storage, key is busy")
	}

	if err := s.aofManager.AppendOperation("SET", key, value); err != nil {
		return fmt.Errorf("Set: failed append operation to aof: %w", err)
	}

	s.store[key] = value

	return nil
}

func (s *Storage) Del(key string) error {
	data, ok := s.store[key]
	if !ok {
		return fmt.Errorf("Del: the key does not exist")
	}

	if err := s.aofManager.AppendOperation("DEL", key, data); err != nil {
		return fmt.Errorf("Del: failed append operation to aof file: %w", err)
	}

	delete(s.store, key)
	return nil
}
