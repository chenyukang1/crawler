package process

import (
	"sync"
	"time"
)

type Config struct {
	Concurrency    int
	MaxRetries     int
	RequestTimeout time.Duration
}

type ICrawler interface {
	Start()
	Stop()
	CanStop()
}

type Crawler struct {
	Fetcher *Fetcher
	Parser  *Parser
	Config  *Config

	status int // 执行状态
	lock   sync.RWMutex
}

func (c *Crawler) Start() {
	c.lock.Lock()
}

func (c *Crawler) Stop() {
	//TODO implement me
	panic("implement me")
}

func (c *Crawler) CanStop() {
	//TODO implement me
	panic("implement me")
}
