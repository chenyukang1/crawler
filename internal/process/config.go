package process

import "time"

type Config struct {
	Concurrency    int
	MaxRetries     int
	RequestTimeout time.Duration
}
