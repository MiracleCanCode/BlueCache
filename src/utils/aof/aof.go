package aof

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type operationJson struct {
	Method string `json:"method"`
	Value  string `json:"value"`
	Key    string `json:"key"`
}

type AOF struct {
	file      *os.File
	mu        sync.Mutex
	buffer    []string
	flushTime time.Duration
}

func NewAOF(aofFileDir string) (*AOF, error) {
	file, err := os.OpenFile(aofFileDir, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("NewAOF: failed to open aof file: %w", err)
	}

	aof := &AOF{
		file:      file,
		buffer:    make([]string, 0, 100),
		flushTime: time.Millisecond * 100,
	}

	go aof.syncWorker()

	return aof, nil
}

func (a *AOF) AppendOperation(method, key string, value ...string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var operation *operationJson
	if len(value) > 0 {
		operation = &operationJson{Method: method, Key: key, Value: value[0]}
	} else {
		operation = &operationJson{Method: method, Key: key}
	}

	messageMarshaledToJson, err := json.Marshal(&operation)
	if err != nil {
		return fmt.Errorf("AppendOperation: failed marshaling data to json: %w", err)
	}

	a.buffer = append(a.buffer, string(messageMarshaledToJson)+"\n")

	if len(a.buffer) >= 100 {
		return a.flush()
	}

	return nil
}

func (a *AOF) flush() error {
	data := strings.Join(a.buffer, "")
	a.buffer = a.buffer[:0]

	_, err := a.file.WriteString(data)
	return err
}

func (a *AOF) syncWorker() {
	ticker := time.NewTicker(a.flushTime)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		if err := a.flush(); err != nil {
			log.Println("AOF syncWorker: failed to flush:", err)
		}
		if err := a.file.Sync(); err != nil {
			log.Println("AOF syncWorker: failed to sync:", err)
		}
		a.mu.Unlock()
	}
}
