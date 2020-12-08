package util

import (
	"time"
)

// RetryFunc ...
type RetryFunc func() error

// Retry retryable function with attempts and delay.
func Retry(rf RetryFunc, attempts int, delay time.Duration) error {
	if err := rf(); err != nil {
		if s, ok := err.(RetryableError); !ok {
			return s
		}
		if attempts--; attempts > 0 {
			time.Sleep(delay)
			return Retry(rf, attempts, delay)
		}
		return err
	}
	return nil
}

type RetryableError struct {
	err error
}

// NewRetryableError ...
func NewRetryableError(err error) RetryableError {
	return RetryableError{err: err}
}

func (re RetryableError) Error() string {
	if re.err != nil {
		return re.err.Error()
	}
	return ""
}