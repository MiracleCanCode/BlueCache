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
}

func New(aof aofInterface, recoverData bool) *Storage {
	return &Storage{
		aofManager:  aof,
		recoverData: recoverData,
	}
}

func (s *Storage) Get(key string) (string, error) {
	data, ok := store[key]
	if !ok {
		return "", fmt.Errorf("Get: failed get data by key item is exist")
	}

  ttl, ok := ttlStore[key]
  if !ok {
    return "", fmt.Errorf("Get: failed get data ttl")
  }
  
  healthRecord, err := s.checkRecordLifetime(ttl)
  if err != nil {
    return "", fmt.Errorf("Get: failed check ttl for record")
  }

  if !healthRecord {
    delete(store, key)
    delete(ttlStore, key)
    return "", fmt.Errorf("Get: failed get data life's time is up")
  }

	return data, nil
}

func (s *Storage) Set(key string, value string, ttl time.Time) error {
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
	modifyKey := strings.TrimSpace(key)
	if !s.recoverData {
		if err := s.aofManager.AppendOperation(constants.DEL_COMMAND, key, nil); err != nil {
			return fmt.Errorf("Del: failed append operation to aof file: %w", err)
		}
	}

	delete(store, modifyKey)
  delete(ttlStore, modifyKey)
	return nil
}

func (s *Storage) checkRecordLifetime(recordLifetime time.Time) (bool, error) {
  const timeLayout = "2006-01-02 15:04:05.9999999 -0700 MST"
  parseRecordLifetime, err := time.Parse(timeLayout, recordLifetime)
  if err != nil {
    return false, fmt.Errorf("checkRecordLifetime: failed parse record lifetime: %w", err)
  }

  now := time.Now()
  if parseRecordLifetime.Before(now) {
    return false, nil 
  }   

  return true, nil
}
