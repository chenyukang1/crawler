package process

import (
	"sync"
)

// GlobalScheduler 全局唯一实例
var GlobalScheduler = NewScheduler(10)

type Scheduler struct {
	queue    TaskQueue
	crawlers *CrawlerPool
	wg       sync.WaitGroup
	stopped  chan struct{}
}

func NewScheduler(cap int) *Scheduler {
	return &Scheduler{
		queue:    GlobalQueue,
		crawlers: NewCrawlerPool(cap),
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

func (s *Scheduler) Submit(task *CrawlTask) {
	s.queue.Push(task)
	s.wg.Add(1)
}
