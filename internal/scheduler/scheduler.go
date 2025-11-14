package scheduler

import (
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/tasks"
)

// Global 全局唯一实例
var Global = NewScheduler(10)

type Scheduler struct {
	queue    tasks.TaskQueue
	crawlers *process.CrawlerPool
	stopped  chan struct{}
}

func NewScheduler(cap int) *Scheduler {
	return &Scheduler{
		queue:    tasks.GlobalQueue,
		crawlers: process.NewCrawlerPool(cap),
		stopped:  make(chan struct{}),
	}
}

func (s *Scheduler) Run() {
	s.queue.Init()
	go func() {
		for i := 0; i < s.crawlers.Cores(); i++ {
			crawler := s.crawlers.Alloc()
			go func() {
				crawler.Start()
			}()
			s.crawlers.Free(crawler)
		}
	}()
}
