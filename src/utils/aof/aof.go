package aof

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"go.uber.org/zap"
)

type AOF struct {
	aofFileDir string
	log        *zap.Logger
}

func New(aofFileDir string, log *zap.Logger) *AOF {
	return &AOF{
		aofFileDir: aofFileDir,
		log:        log,
	}
}

func (a *AOF) RecoverData(port string) error {
	addr := fmt.Sprintf("localhost:%s", port)
	file, err := os.Open(a.aofFileDir)
	if err != nil {
		return fmt.Errorf("RecoverData: failed open isaRedis file: %w", err)
	}
	defer file.Close()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("RecoverData: failed create connection to isaRedis: %w", err)
	}

	defer conn.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		conn.Write([]byte(line + "\n"))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("RecoverData: failed scan aof file: %w", err)
	}

	return nil
}

func (a *AOF) AppendOperation(method string, key string, value string) error {
	file, err := a.openFile()
	if err != nil {
		return fmt.Errorf("AppendOperation: failed open aof file: %w", err)
	}

	operation := fmt.Sprintf("%s %s %s\n", method, key, value)
	if _, err := file.WriteString(operation); err != nil {
		return fmt.Errorf("AppendOperation: failed append operation to aof file: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("AppendOperation: failed sync aof file: %w", err)
	}
	return nil
}

func (a *AOF) openFile() (*os.File, error) {
	file, err := os.OpenFile(a.aofFileDir,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(os.O_RDWR))
	if err != nil {
		return nil, fmt.Errorf("openFile: failed open aof file: %w", err)
	}

	return file, nil
}
