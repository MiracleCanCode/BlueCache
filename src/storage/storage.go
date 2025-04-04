package storage

import (
	"fmt"
	"strings"

	"github.com/minikeyvalue/src/utils/constants"
)

type aofInterface interface {
	AppendOperation(method string, key string, value ...string) error
}

var store map[string]string = make(map[string]string)

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

	return data, nil
}

func (s *Storage) Set(key string, value string) error {
	_, ok := store[key]
	if ok {
		return fmt.Errorf("Set: failed set data to storage, key is busy: %s", key)
	}

	if !s.recoverData {
		if err := s.aofManager.AppendOperation(constants.SET_COMMAND, key, value); err != nil {
			return fmt.Errorf("Set: failed append operation to aof: %w", err)
		}
	}

	store[key] = value

	return nil
}

func (s *Storage) Del(key string) error {
	modifyKey := strings.TrimSpace(key)
	if !s.recoverData {
		if err := s.aofManager.AppendOperation(constants.DEL_COMMAND, key); err != nil {
			return fmt.Errorf("Del: failed append operation to aof file: %w", err)
		}
	}

	delete(store, modifyKey)

	return nil
}
