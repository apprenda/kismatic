package retry

import "time"

// WithBackoff will retry a function specified number of times
func WithBackoff(fn func() error, retries uint) error {
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
		time.Sleep((1 << attempts) * time.Second)
		attempts++
	}
	return err
}
