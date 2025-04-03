package retry

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

func RetryOperation(log *zap.Logger, operation func() error, baseDelay int,
	attempts int) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		lastErr = operation()
		if lastErr == nil {
			if i > 0 {
				log.Info("Operation complete", zap.Int("attempt", i+1))
				return nil
			}
			return nil
		}

		delay := time.Millisecond * time.Duration(baseDelay*((i+1)*i))

		log.Error("Operation failed", zap.Int("attempt", i+1),
			zap.Duration("retry_time", delay), zap.Error(lastErr))
		time.Sleep(delay)
	}

	return fmt.Errorf("RetryOperation: Operation failed after %d attempts, last error: %w",
		attempts, lastErr)
}
