package tasks

import (
	"context"
	"errors"
	"github.com/chenyukang1/crawler/internal/logger"
	"math"
	"time"
)

type Logic func() error

type Retry interface {
	DoRetry(ctx context.Context, logic Logic) error
}

type FixedRetry struct {
	ReTryTimes int
	Interval   time.Duration
}

func (f *FixedRetry) DoRetry(ctx context.Context, logic Logic) error {
	for i := 0; i < f.ReTryTimes; i++ {
		if err := logic(); err == nil {
			return nil
		}
		logger.Errorf("fail at %d, retrying...", i)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(f.Interval):
		}
	}
	return errors.New("retry failed")
}

type BackoffRetry struct {
	ReTryTimes int
	Interval   time.Duration
}

func (f *BackoffRetry) DoRetry(ctx context.Context, logic Logic) error {
	for i := 0; i < f.ReTryTimes; i++ {
		if err := logic(); err == nil {
			return nil
		}
		logger.Errorf("fail at %d, retrying...", i)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(float64(f.Interval) * math.Pow(2, float64(i)))):
		}
	}
	return errors.New("retry failed")
}
