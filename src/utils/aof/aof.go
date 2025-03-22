package aof

import (
	"fmt"
	"os"
)

type AOF struct {
	aofFileDir string
}

func New(aofFileDir string) *AOF {
	return &AOF{
		aofFileDir: aofFileDir,
	}
}

func (a *AOF) AppendOperation(method string, key string, value ...string) error {
	file, err := os.OpenFile(a.aofFileDir,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(os.O_RDWR))
	if err != nil {
		return fmt.Errorf("AppendOperation: failed open aof file: %w", err)
	}

	var operation string
	if len(value) > 0 {
		operation = fmt.Sprintf("%s %s %s\n", method, key, value[0])
	} else {
		operation = fmt.Sprintf("%s %s\n", method, key)
	}

	if _, err := file.WriteString(operation); err != nil {
		return fmt.Errorf("AppendOperation: failed append operation to aof file: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("AppendOperation: failed sync aof file: %w", err)
	}
	return nil
}
