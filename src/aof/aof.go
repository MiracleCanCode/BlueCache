package aof

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
  "github.com/minikeyvalue/src/utils/constants"
	"go.uber.org/zap"
)

type AOF struct {
	file      *os.File
	mu        sync.Mutex
	buffer    []string
	flushTime time.Duration
	logger    *zap.Logger
}

const flushTimeDefault = time.Second * 10
const maxBufferSize = 100

func NewAOF(aofFileDir string, log *zap.Logger) (*AOF, error) {
	file, err := os.OpenFile(aofFileDir, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("NewAOF: failed to open aof file: %w", err)
	}

	aof := &AOF{
		file:      file,
		buffer:    make([]string, 0, maxBufferSize),
		flushTime: flushTimeDefault,
		logger:    log,
	}

	go func() {
		if err := aof.syncWorker(); err != nil {
			log.Error("Failed aof sync worker", zap.Error(err))
		}
	}()

	return aof, nil
}

func (a *AOF) AppendOperation(method, key string, ttl time.Time, value ...string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
  
	var val string
	if len(value) > 0 {
		val = value[0]
	}
  
  var message string
  if method == constants.SET_COMMAND {
	  message = fmt.Sprintf("%s %s %s %s",  method, ttl, key, val)
  } else {
	  message = fmt.Sprintf("%s %s %s", method, key, val)
  }

	a.buffer = append(a.buffer, message+"\n")

	if len(a.buffer) >= maxBufferSize {
		if err := a.flush(); err != nil {
			return fmt.Errorf("AppendOperation: failed flush: %w", err)
		}
	}

	return nil
}

func (a *AOF) flush() error {
	data := strings.Join(a.buffer, "")

	if _, err := a.file.WriteString(data); err != nil {
		return fmt.Errorf("flush: failed write string to file: %w", err)
	}

	a.buffer = a.buffer[:0]
	return nil
}

func (a *AOF) syncWorker() error {
	ticker := time.NewTicker(a.flushTime)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		if err := a.flush(); err != nil {
			a.logger.Error("Failed flush time", zap.Error(err))
		}
		if err := a.file.Sync(); err != nil {
			a.logger.Error("Failed file sync", zap.Error(err))
		}
		a.mu.Unlock()
	}

	return nil
}
