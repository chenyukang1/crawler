package scheduler

import (
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/tasks"
	"sync"
)

// Global 全局唯一实例
var Global = NewScheduler(10)

type Scheduler struct {
	queue    tasks.TaskQueue
	crawlers *process.CrawlerPool
	wg       sync.WaitGroup
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
			s.wg.Add(1)
			crawler := s.crawlers.Alloc()
			go func() {
				defer func() {
					s.crawlers.Free(crawler)
				}()
				crawler.Start()
			}()
		}
	}()
	s.wg.Wait()
}

func (s *Scheduler) Submit(task tasks.CrawlTask) {
	s.queue.Push(task)
	s.wg.Add(1)
}
