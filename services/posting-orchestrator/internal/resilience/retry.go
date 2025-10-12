package resilience

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffMultiple float64
}

// DefaultRetryConfig returns sensible defaults for retry
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        2 * time.Second,
		BackoffMultiple: 2.0,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// Retry executes a function with exponential backoff
func Retry(ctx context.Context, config RetryConfig, fn RetryableFunc, logger *log.Logger) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn(ctx)
		if err == nil {
			// Success!
			if attempt > 1 {
				logger.Printf("Retry succeeded on attempt %d", attempt)
			}
			return nil
		}

		lastErr = err
		
		// Check if we've exhausted attempts
		if attempt >= config.MaxAttempts {
			logger.Printf("Retry exhausted after %d attempts: %v", config.MaxAttempts, err)
			break
		}

		// Log the retry
		logger.Printf("Attempt %d failed: %v, retrying in %v", attempt, err, delay)

		// Wait before retry (with context cancellation support)
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}

		// Increase delay for next attempt (exponential backoff)
		delay = time.Duration(float64(delay) * config.BackoffMultiple)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", config.MaxAttempts, lastErr)
}
