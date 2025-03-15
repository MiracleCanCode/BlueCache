package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type parseData struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type storage struct {
	storage    map[string]any
	mu         sync.Mutex
	storageDir string
}

func NewWithLoadData(storageDir string) (*storage, error) {
	s := &storage{
		storage:    make(map[string]any),
		storageDir: storageDir,
	}
	var parsedData []parseData
	file, err := os.ReadFile(storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("LoadDataForStorage: failed load data, error open file: %w",
			err)
	}
	if len(file) > 0 {
		if err := json.Unmarshal(file, &parsedData); err != nil {
			return nil, fmt.Errorf("LoadDataForStorage: failed parse data: %w", err)
		}

		for _, val := range parsedData {
			s.storage[val.Key] = val.Value
		}
	}

	return s, nil
}

func (s *storage) Add(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if key == "" {
		return fmt.Errorf("Add: key does not is empty")
	}

	_, exist := s.storage[key]
	if exist {
		return fmt.Errorf("Add: key is busy")
	}
	file, err := os.OpenFile(s.storageDir, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("Add: failed open storage file: %w", err)
	}
	defer file.Close()

	s.storage[key] = value

	var mapToDataForStorage []parseData
	for key, value := range s.storage {
		mapToDataForStorage = append(mapToDataForStorage, parseData{
			Key:   key,
			Value: value,
		})
	}

	writingData, err := json.Marshal(mapToDataForStorage)
	if err != nil {
		return fmt.Errorf("Add: failed marshal data: %w", err)
	}

	if _, err := file.Write(writingData); err != nil {
		return fmt.Errorf("Add: failed write data to file: %w", err)
	}

	return nil
}

func (s *storage) Get(key string) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.storage[key]
	if !ok {
		return nil, fmt.Errorf("Get: there is no data for this key %s", key)
	}

	return data, nil
}
