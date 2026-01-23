package process

import (
	"context"
	"sync"
)

type CrawlerPool struct {
	count    int
	capacity int
	pool     chan *Crawler
	ctx      context.Context
	mu       sync.Mutex
	wg       sync.WaitGroup
}

func NewCrawlerPool(c context.Context, p int) *CrawlerPool {
	return &CrawlerPool{
		capacity: p,
		pool:     make(chan *Crawler, p),
		ctx:      c,
	}
}

func (p *CrawlerPool) Alloc() *Crawler {
	p.wg.Add(1)

	select {
	case c := <-p.pool:
		return c
	default:
	}

	p.mu.Lock()
	if p.count < p.capacity {
		crawler := NewCrawler(p.ctx)
		p.count++
		p.mu.Unlock()
		return crawler
	}
	p.mu.Unlock()

	return <-p.pool
}

func (p *CrawlerPool) Free(crawler *Crawler) {
	p.mu.Lock()
	p.count--
	p.mu.Unlock()
	p.pool <- crawler
}

func (p *CrawlerPool) Wait() {
	p.wg.Wait()
}
