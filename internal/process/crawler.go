package process

import (
	"context"
	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/internal/fetch"
	"github.com/chenyukang1/crawler/internal/logger"
	"github.com/chenyukang1/crawler/internal/parse"
	"github.com/chenyukang1/crawler/internal/status"
	"github.com/chenyukang1/crawler/internal/tasks"
	"sync"
	"time"
)

type ICrawler interface {
	Start()
	Pause()
	Stop()
	CanStop() bool
	Status() int
}

type Crawler struct {
	Fetcher   *fetch.Fetcher
	Parser    *parse.Parser
	Collector *collect.Collector

	status      int // 执行状态
	idleSeconds int // 空闲时间，超过被回收
	lock        sync.RWMutex
}

func NewCrawler() *Crawler {
	return &Crawler{
		Fetcher: fetch.Default,
		Parser:  parse.Default,
		status:  status.INITIAL,
	}
}

func (c *Crawler) Start() {
	c.setStatus(status.RUN)

	finish := make(chan int, 1)
	go func() {
		c.run()
		finish <- 1
	}()
	<-finish
}

func (c *Crawler) Pause() {
	c.setStatus(status.PAUSE)
}

func (c *Crawler) Stop() {
	if c.Status() == status.STOP {
		return
	}
	if c.CanStop() {
		c.setStatus(status.STOP)
	}
}

func (c *Crawler) CanStop() bool {
	return true
}

func (c *Crawler) Status() int {
	c.lock.RLock()
	s := c.status
	c.lock.Unlock()
	return s
}

func (c *Crawler) run() {
	idleTime := 0
loop:
	for idleTime < c.idleSeconds {
		if c.isPause() {
			time.Sleep(time.Second)
			goto loop
		}
		task := tasks.GlobalQueue.Pop()
		if task == nil {
			time.Sleep(time.Second)
			idleTime++
			continue
		}
		req, err := fetch.BuildRequest(task)
		if err != nil {
			logger.Errorf("fetch url %s fail %v ", task.Url, err)
			return
		}
		ctx := context.Background()
		resp, err := c.Fetcher.Fetch(ctx, req)
		if err != nil {
			logger.Errorf("fetch url %s fail %v ", task.Url, err)
			return
		}
		c.Parser.Parse(resp)
		idleTime = 0
	}
}

func (c *Crawler) isPause() bool {
	return c.Status() == status.PAUSE
}

func (c *Crawler) setStatus(status int) {
	c.lock.Lock()
	c.status = status
	c.lock.Unlock()
}
