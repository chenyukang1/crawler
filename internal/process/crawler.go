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
	Collector collect.Collector

	fetcher  *fetch.Fetcher
	parser   *parse.Parser
	status   int           // 执行状态
	finish   chan int      //结束channel
	idleTime time.Duration // 空闲时间，超过被回收
	lock     sync.RWMutex
}

func NewCrawler() *Crawler {
	return &Crawler{
		Collector: collect.Log,
		fetcher:   fetch.Default,
		parser:    parse.Default,
		status:    status.INITIAL,
		finish:    make(chan int, 1),
		idleTime:  60 * time.Second,
	}
}

func (c *Crawler) Start() {
	c.setStatus(status.RUN)

	// 开始收集数据
	c.Collector.Pipeline()
	go func() {
		c.run()
		c.finish <- 1
	}()
	<-c.finish
}

func (c *Crawler) Pause() {
	c.setStatus(status.PAUSE)
}

func (c *Crawler) Stop() {
	if c.Status() == status.STOP {
		return
	}
	if c.CanStop() {
		c.Collector.Stop()
		close(c.finish)
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
	idled := 0 * time.Second
loop:
	for idled < c.idleTime {
		if c.isPause() {
			time.Sleep(time.Second)
			goto loop
		}
		task := tasks.GlobalQueue.Pop()
		if task == nil {
			time.Sleep(time.Second)
			idled++
			continue
		}
		req, err := fetch.BuildRequest(task)
		if err != nil {
			logger.Errorf("fetch url %s fail %v ", task.Url, err)
			return
		}
		ctx := context.Background()
		resp, err := c.fetcher.Fetch(ctx, req)
		if err != nil {
			logger.Errorf("fetch url %s fail %v ", task.Url, err)
			return
		}
		result := c.parser.Parse(resp)
		for _, cell := range result.GetStructuredData() {
			c.Collector.Push(cell)
		}
		idled = 0
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
