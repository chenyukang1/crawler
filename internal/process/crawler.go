package process

import (
	"github.com/chenyukang1/crawler/internal/fetch"
	"github.com/chenyukang1/crawler/internal/status"
	"github.com/chenyukang1/crawler/internal/tasks"
	"sync"
)

type ICrawler interface {
	Start()
	Stop()
	CanStop() bool
}

type Crawler struct {
	Fetcher *fetch.Fetcher
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

func (c *Crawler) Start() {
	c.lock.Lock()
	c.status = status.RUN
	c.lock.Unlock()
	task := tasks.GlobalQueue.Pop()
	if task == nil {
		return
	}
}

func (c *Crawler) Stop() {
	//TODO implement me
	panic("implement me")
}

func (c *Crawler) CanStop() bool {
	return true
}
