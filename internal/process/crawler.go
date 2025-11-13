package process

import (
	"github.com/chenyukang1/crawler/internal/status"
	"github.com/chenyukang1/crawler/internal/tasks"
	"sync"
)

type ICrawler interface {
	Start(task tasks.CrawlTask, finish chan bool)
	Stop()
	CanStop() bool
}

type Crawler struct {
	Fetcher *Fetcher
	Parser  *Parser

	status int // 执行状态
	lock   sync.RWMutex
}

func NewCrawler() *Crawler {
	return &Crawler{
		Fetcher: nil,
		Parser:  nil,
		status:  status.INITIAL,
	}
}

func (c *Crawler) Start(task tasks.CrawlTask, finish chan bool) {
	c.lock.Lock()
	c.status = status.RUN
}

func (c *Crawler) Stop() {
	//TODO implement me
	panic("implement me")
}

func (c *Crawler) CanStop() bool {
	//TODO implement me
	panic("implement me")
}
