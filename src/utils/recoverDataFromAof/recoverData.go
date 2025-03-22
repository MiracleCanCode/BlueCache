package recoverdatafromaof

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type storage interface {
	Set(key string, value string) error
	Del(key string) error
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
		if len(line) == 0 || line == "EOF" {
			continue
		}
		if err := r.distributeData(line); err != nil {
			return fmt.Errorf("Recover: failed recover data to storage: %w", err)
		}

	}
	return nil
}

func (r *recoverData) distributeData(message string) error {
	if strings.HasPrefix(message, "SET") {
		parts := strings.SplitN(message, " ", 3)
		if len(parts) != 3 {
			return fmt.Errorf("distributeData: incrorect set data string")
		}
		if err := r.store.Set(parts[1], parts[2]); err != nil {
			return fmt.Errorf("distributeData: failed set data: %w", err)
		}
	}

	if strings.HasPrefix(message, "DEL") {
		parts := strings.SplitN(message, " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("distributeData: incorect delete data string")
		}
		if err := r.store.Del(parts[1]); err != nil {
			return fmt.Errorf("distributeData: failed delete data: %w", err)
		}
	}
	return nil
}
