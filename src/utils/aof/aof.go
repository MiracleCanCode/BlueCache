package aof

import (
	"fmt"
	"os"
)

type AOF struct {
	aofFileDir string
	store      map[string]string
}

func New(aofFileDir string) *AOF {
	return &AOF{
		aofFileDir: aofFileDir,
	}
}

// func (a *AOF) RecoverData() error {
// file, err := os.Open(a.aofFileDir)
// if err != nil {
// return fmt.Errorf("RecoverData: failed open aof file: %w", err)
// }
// }

func (a *AOF) AppendOperation(method string, value string, key string) error {
	file, err := a.openFile()
	if err != nil {
		return fmt.Errorf("AppendOperation: failed open aof file: %w", err)
	}

	operation := fmt.Sprintf("%s %s %s\n", method, key, value)
	if _, err := file.Write([]byte(operation)); err != nil {
		return fmt.Errorf("AppendOperation: failed append operation to aof file: %w", err)
	}
	return nil
}

func (a *AOF) openFile() (*os.File, error) {
	file, err := os.OpenFile(a.aofFileDir, os.O_WRONLY|os.O_APPEND, os.FileMode(os.O_CREATE))
	if err != nil {
		return nil, fmt.Errorf("openFile: failed open aof file: %w", err)
	}

	return file, nil
}
