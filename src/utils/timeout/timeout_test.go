package timeout_test

import (
	"testing"
	"time"

	"github.com/minikeyvalue/src/utils/timeout"
)

func TestTimeout(t *testing.T) {
	timeoutFn := func() error {
		operationTime := 4000 * time.Millisecond
		time.Sleep(operationTime)
		t.Fatal("Operation is complete, test fail")
		return nil
	}

	operationTime := 2000
	if err := timeout.Operation(int(operationTime), timeoutFn); err != nil {
		t.Log("Test success")
		return
	}
	t.Fatal("Timeout did not trigger, test failed")
}
