package process

import (
	"github.com/chenyukang1/crawler/internal/fetch"
	"github.com/chenyukang1/crawler/internal/status"
	"github.com/chenyukang1/crawler/internal/tasks"
	"sync"
	"time"
)

type ICrawler interface {
	Start()
	Stop()
	CanStop() bool
}

type Crawler struct {
	Fetcher *fetch.Fetcher
	Parser  *Parser

	status      int // 执行状态
	idleSeconds int // 空闲时间，超过被回收
	lock        sync.RWMutex
}

func NewCrawler() *Crawler {
	return &Crawler{
		Fetcher: nil,
		Parser:  nil,
		status:  status.INITIAL,
	}
}

func (c *Crawler) Start() {
	c.lock.Lock()
	c.status = status.RUN
	c.lock.Unlock()

	idleTime := 0
	for idleTime < c.idleSeconds {
		task := tasks.GlobalQueue.Pop()
		if task == nil {
			idleTime++
			time.Sleep(time.Second)
			continue
		}

	}
}

func (c *Crawler) Stop() {
	//TODO implement me
	panic("implement me")
}

func (c *Crawler) CanStop() bool {
	return true
}
