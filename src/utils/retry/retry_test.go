package retry_test

import (
	"errors"
	"testing"

	"github.com/minikeyvalue/src/utils/retry"
	"go.uber.org/zap"
)

func TestRetry(t *testing.T) {
	log := zap.NewNop()

	var attempts int
	fn := func() error {
		attempts++
		return errors.New("Failed operation")
	}

	operationTime := 200
	operationAttempts := 5

	err := retry.RetryOperation(log, fn, operationTime, operationAttempts)
	if err == nil {
		t.Fatal("Expected error, but got nil")
	}

	if attempts != operationAttempts {
		t.Fatalf("Expected %d attempts, but got %d", operationAttempts, attempts)
	}
}
