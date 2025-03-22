package recoverdatafromaof

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type storage interface {
	Set(key string, value string) error
	Del(key string) error
}

type message struct {
	Method string `json:"method"`
	Value  string `json:"value"`
	Key    string `json:"key"`
}

type recoverData struct {
	store storage
}

func New(store storage) *recoverData {
	return &recoverData{
		store: store,
	}
}

func (r *recoverData) Recover(aofFilePath string) error {
	file, err := os.Open(aofFilePath)
	if err != nil {
		return fmt.Errorf("RecoverData: failed open isaRedis file: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var msg message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return fmt.Errorf("Recover: failed unmarshaling line from aof file: %w", err)
		}
		if len(line) == 0 || line == "EOF" {
			continue
		}
		if err := r.distributeData(msg); err != nil {
			return fmt.Errorf("Recover: failed recover data to storage: %w", err)
		}

	}
	return nil
}

func (r *recoverData) distributeData(msg message) error {
	if msg.Method == "SET" {
		if err := r.store.Set(msg.Key, msg.Value); err != nil {
			return fmt.Errorf("distributeData: failed set data: %w", err)
		}
	}

	if msg.Method == "DEL" {
		if err := r.store.Del(msg.Key); err != nil {
			return fmt.Errorf("distributeData: failed delete data: %w", err)
		}
	}
	return nil
}
