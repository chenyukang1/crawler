package process

import (
	"context"
	"sync"
	"time"
)

type CrawlerPool struct {
	count    int
	capacity int
	pool     chan *Crawler
	all      []*Crawler
	mu       sync.Mutex
}

func NewCrawlerPool(capacity int) *CrawlerPool {
	return &CrawlerPool{
		capacity: capacity,
		pool:     make(chan *Crawler, capacity),
		all:      make([]*Crawler, capacity),
	}
}

func (p *CrawlerPool) Alloc(ctx context.Context) *Crawler {
	for {
		select {
		case crawler := <-p.pool:
			return crawler
		default:
			p.mu.Lock()
			if p.count < p.capacity {
				crawler := NewCrawler(ctx)
				p.all = append(p.all, crawler)
				p.count++
				p.mu.Unlock()
				return crawler
			}
		}
		time.Sleep(time.Second)
	}
}

func (p *CrawlerPool) Free(crawler *Crawler) {
	if !crawler.CanStop() {
		return
	}
	p.mu.Lock()
	for i, c := range p.all {
		if c == crawler {
			p.all[i] = p.all[len(p.all)-1]
			p.all = p.all[:len(p.all)-1]
			break
		}
	}
	p.count--
	p.mu.Unlock()
	p.pool <- crawler
}

func (p *CrawlerPool) Stop() {
	for _, crawler := range p.all {
		crawler.Stop()
	}
}
