package timeout

import (
	"context"
	"fmt"
	"time"
)

func Operation(delayMiliseconds int, fn func() error) error {
	delay := time.Millisecond * time.Duration(delayMiliseconds)
	ctx, cancel := context.WithTimeout(context.Background(), delay)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("Operation failed: %w", err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("Operation timeout!")
	}
}
