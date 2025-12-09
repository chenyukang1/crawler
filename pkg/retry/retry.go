package retry

import (
	"context"
	"errors"
	"github.com/chenyukang1/crawler/pkg/log"
	"math"
	"time"
)

type Logic func() (any, error)

type Retry interface {
	DoRetry(ctx context.Context, logic Logic) (any, error)
}

type FixedRetry struct {
	ReTryTimes int
	Interval   time.Duration
}

func (f *FixedRetry) DoRetry(ctx context.Context, logic Logic) (any, error) {
	for i := 0; i < f.ReTryTimes; i++ {
		if res, err := logic(); err == nil {
			return res, nil
		}
		log.Errorf("fail at %d, retrying...", i)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(f.Interval):
		}
	}
	return nil, errors.New("retry failed")
}

type BackoffRetry struct {
	ReTryTimes int
	Interval   time.Duration
}

func (f *BackoffRetry) DoRetry(ctx context.Context, logic Logic) (any, error) {
	for i := 0; i < f.ReTryTimes; i++ {
		if res, err := logic(); err == nil {
			return res, nil
		}
		log.Errorf("fail at %d, retrying...", i)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(float64(f.Interval) * math.Pow(2, float64(i)))):
		}
	}
	return nil, errors.New("retry failed")
}
