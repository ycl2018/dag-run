package dagRun

import "time"

type TaskOption func(*option)

type option struct {
	retry   int
	timeout time.Duration
}

// Retry set task max retry times
func Retry(maxTimes int) TaskOption {
	return func(o *option) {
		o.retry = maxTimes
	}
}

// Timeout set task timeout duration
func Timeout(timeout time.Duration) TaskOption {
	return func(o *option) {
		o.timeout = timeout
	}
}
