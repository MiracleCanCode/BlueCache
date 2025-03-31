package aof

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

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

	flushTimeDefault := time.Millisecond * 100
	aof := &AOF{
		file:      file,
		buffer:    make([]string, 0, 1000),
		flushTime: flushTimeDefault,
	}
	go func() {
		if err := aof.syncWorker(); err != nil {
			return
		}
	}()

	return aof, nil
}

func (a *AOF) AppendOperation(method, key string, value ...string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var val string
	if len(value) > 0 {
		val = value[0]
	}
	message := fmt.Sprintf("%s %s %s", method, key, val)

	a.buffer = append(a.buffer, message+"\n")

	if len(a.buffer) >= 100 {
		if err := a.flush(); err != nil {
			return fmt.Errorf("AppendOperation: failed flush: %w", err)
		}
	}

	return nil
}

func (a *AOF) flush() error {
	data := strings.Join(a.buffer, "")
	a.buffer = a.buffer[:0]

	if _, err := a.file.WriteString(data); err != nil {
		return fmt.Errorf("flush: failed write string to file: %w", err)
	}
	return nil
}

func (a *AOF) syncWorker() error {
	ticker := time.NewTicker(a.flushTime)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		if err := a.flush(); err != nil {
			return fmt.Errorf("syncWorker: failed flush time: %w", err)
		}
		if err := a.file.Sync(); err != nil {
			return fmt.Errorf("syncWorker: failed sync file: %w", err)
		}
		a.mu.Unlock()
	}

	return nil
}
