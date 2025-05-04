package recoverdatafromaof

import (
	"bufio"
	"fmt"
	"os"
	"strings"
  "time"
	"github.com/minikeyvalue/src/utils/constants"
)

type storage interface {
	Set(key string, ttl time.Time, value string) error
	Del(key string) error
}

type recoverData struct {
	store storage
}

const TIME_LAYOUT = "2006-01-02 15:04:05.9999999 -0700 MST"

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
		if len(line) == 0 {
			continue
		}
		if err := r.distributeData(line); err != nil {
			return fmt.Errorf("Recover: failed recover data to storage: %w", err)
		}

	}
	return nil
}

func (r *recoverData) distributeData(message string) error {
	if strings.HasPrefix(message, constants.SET_COMMAND) {
		parts := strings.SplitN(message, " ", 7)

		if len(parts) != 7 {
			return fmt.Errorf("distributeData: incrorect set data string")
		}
    
    recordLifeTime := fmt.Sprintf("%s %s %s %s", parts[1], parts[2], parts[3], parts[4])
    parseRecordLifetime, timePass, err := r.checkValidTime(recordLifeTime)
    if err != nil {
      return fmt.Errorf("distributeData: failed check time valid: %w", err)
    }

    if !timePass {
      return nil
    }

		if err := r.store.Set(parts[5], parseRecordLifetime, parts[6]); err != nil {
			return fmt.Errorf("distributeData: failed set data: %w", err)
		}
	}

	if strings.HasPrefix(message, constants.DEL_COMMAND) {
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

func (r *recoverData) checkValidTime(lifetimeRecord string) (time.Time, bool, error) {
    parseRecordLifetime, err := time.Parse(TIME_LAYOUT, lifetimeRecord)
    if err != nil {
      return parseRecordLifetime, false, fmt.Errorf("distributeData: failed parse record lifetime: %w", err)
    }
    
    now := time.Now()
    if parseRecordLifetime.Before(now) {
      return parseRecordLifetime, false, nil
    }

    return parseRecordLifetime, true, nil
}
