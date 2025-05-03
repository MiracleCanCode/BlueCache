package storage

import (
	"fmt"
	"strings"
  "time"
	"github.com/minikeyvalue/src/utils/constants"
)

type aofInterface interface {
	AppendOperation(method string, key string, ttl time.Time, value ...string) error
}

var store map[string]string = make(map[string]string)
var ttlStore map[string]time.Time = make(map[string]time.Time)

type Storage struct {
	aofManager  aofInterface
	recoverData bool
  mu sync.Mutex
}

func New(aof aofInterface, recoverData bool) *Storage {
	return &Storage{
		aofManager:  aof,
		recoverData: recoverData,
	}
}

func (s *Storage) Get(key string) (string, error) {
  s.mu.Lock()
  defer s.mu.Unlock()
	data, ok := store[key]
	if !ok {
		return "", fmt.Errorf("Get: failed get data by key item is exist")
	}

  ttl, ok := ttlStore[key]
  if !ok {
    return "", fmt.Errorf("Get: failed get data ttl")
  }

  if !s.checkRecordLifetime(ttl) {
    delete(store, key)
    delete(ttlStore, key)
    return "", fmt.Errorf("Get: failed get data life's time is up")
  }

	return data, nil
}

func (s *Storage) Set(key string, ttl time.Time, value string) error {
  s.mu.Lock()
  defer s.mu.Unlock()
	_, ok := store[key]
	if ok {
		return fmt.Errorf("Set: failed set data to storage, key is busy: %s", key)
	}

  ttlStore[key] = ttl
 
	if !s.recoverData {
		if err := s.aofManager.AppendOperation(constants.SET_COMMAND, key, ttl, value); err != nil {
			return fmt.Errorf("Set: failed append operation to aof: %w", err)
		}
	}

	store[key] = value

	return nil
}

func (s *Storage) Del(key string) error {
  s.mu.Lock()
  defer s.mu.Unlock()
	modifyKey := strings.TrimSpace(key)
	if !s.recoverData {
		if err := s.aofManager.AppendOperation(constants.DEL_COMMAND, key, time.Time{}); err != nil {
			return fmt.Errorf("Del: failed append operation to aof file: %w", err)
		}
	}

	delete(store, modifyKey)
  delete(ttlStore, modifyKey)
	return nil
}

func (s *Storage) checkRecordLifetime(recordLifetime time.Time) bool {
  now := time.Now()
  if recordLifetime.Before(now) {
    return false 
  }   

  return true
}
