package process

import (
	"context"
	"sync"

	"github.com/chenyukang1/crawler/pkg/log"
)

type CrawlerPool struct {
	seq      int
	capacity int
	pool     chan *Crawler
	ctx      context.Context
	mu       sync.Mutex
}

func NewCrawlerPool(c context.Context, p int) *CrawlerPool {
	return &CrawlerPool{
		capacity: p,
		pool:     make(chan *Crawler, p),
		ctx:      c,
	}
}

func (p *CrawlerPool) Alloc(q TaskQueue) *Crawler {
	select {
	case c := <-p.pool:
		return c
	default:
	}

	p.mu.Lock()
	if p.seq < p.capacity {
		crawler := NewCrawler(p.ctx, p.seq, q)
		p.seq++
		p.mu.Unlock()
		return crawler
	}
	p.mu.Unlock()

	return <-p.pool
}

func (p *CrawlerPool) Free(c *Crawler) {
	log.Infof("[crawler-%d] free", c.seq)
	p.pool <- c
}
