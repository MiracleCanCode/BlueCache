package aof

import (
	"encoding/json"
	"fmt"
	"os"
)

type operationJson struct {
	Method string `json:"method"`
	Value  string `json:"value"`
	Key    string `json:"key"`
}

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

	var operation *operationJson
	if len(value) > 0 {
		operation = &operationJson{
			Method: method,
			Key:    key,
			Value:  value[0],
		}
	} else {
		operation = &operationJson{
			Method: method,
			Key:    key,
		}
	}
	messageMarshaledToJson, err := json.Marshal(&operation)
	if err != nil {
		return fmt.Errorf("AppendOperation: failed marshaling data to json: %w", err)
	}
	if _, err := file.WriteString(string(messageMarshaledToJson) + "\n"); err != nil {
		return fmt.Errorf("AppendOperation: failed append operation to aof file: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("AppendOperation: failed sync aof file: %w", err)
	}
	return nil
}
