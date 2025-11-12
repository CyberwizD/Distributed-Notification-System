package retry

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// Config describes the retry behavior.
type Config struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	JitterFactor   float64
}

// Do executes fn and retries with exponential backoff until it succeeds or the
// context is cancelled.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = 500 * time.Millisecond
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 10 * time.Second
	}
	if cfg.JitterFactor <= 0 {
		cfg.JitterFactor = 0.2
	}

	backoff := cfg.InitialBackoff
	var err error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		if err = fn(); err == nil {
			return nil
		}

		if attempt == cfg.MaxAttempts {
			break
		}

		sleep := applyJitter(backoff, cfg.JitterFactor)
		if sleep > cfg.MaxBackoff {
			sleep = cfg.MaxBackoff
		}

		timer := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.Join(err, ctx.Err())
		case <-timer.C:
		}

		if backoff < cfg.MaxBackoff {
			backoff *= 2
			if backoff > cfg.MaxBackoff {
				backoff = cfg.MaxBackoff
			}
		}
	}
	return err
}

func applyJitter(duration time.Duration, factor float64) time.Duration {
	if factor <= 0 {
		return duration
	}
	delta := int64(float64(duration) * factor)
	return duration + time.Duration(rand.Int63n(2*delta)-delta)
}
