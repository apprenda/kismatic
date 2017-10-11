package retry

import "time"

type retryMethod int

const (
	withBackoff retryMethod = iota
	linear
)

// WithBackoff will retry a function specified number of times with an exponential backoff
func WithBackoff(retries uint, fn func() error) error {
	return retry(fn, retries, withBackoff)
}

// Linear will retry a function specified number of times
func Linear(retries uint, fn func() error) error {
	return retry(fn, retries, linear)
}

func retry(fn func() error, retries uint, method retryMethod) error {
	var attempts uint
	var err error
	for {
		err = fn()
		if err == nil {
			break
		}
		if attempts == retries {
			break
		}
		sleep := 1 * time.Second
		switch method {
		case withBackoff:
			sleep = (1 << attempts) * time.Second
		case linear:
			sleep = 1 * time.Second
		}
		time.Sleep(sleep)
		attempts++
	}
	return err
}
