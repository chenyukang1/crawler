package process

import (
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

func (p *CrawlerPool) Alloc() *Crawler {
	for {
		select {
		case crawler := <-p.pool:
			return crawler
		default:
			p.mu.Lock()
			if p.count < p.capacity {
				crawler := NewCrawler()
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
	p.pool <- crawler
}

// Cores 核心爬虫数
func (p *CrawlerPool) Cores() int {
	return p.capacity << 1
}
