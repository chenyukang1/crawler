package scheduler

import (
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/tasks"
	"sync"
	"time"
)

// Global 全局唯一实例
var Global = NewScheduler(100)

type Scheduler struct {
	queue   tasks.TaskQueue
	workers *process.CrawlerPool
	in      chan tasks.CrawlTask // 生产者输入
	out     chan tasks.CrawlTask // 分发给 worker
	stopped chan struct{}
	mu      sync.Mutex
}

func NewScheduler(buffer int) *Scheduler {
	return &Scheduler{
		queue:   new(tasks.TaskQueueHeapWrapper),
		workers: process.NewCrawlerPool(10),
		in:      make(chan tasks.CrawlTask, buffer),
		out:     make(chan tasks.CrawlTask, buffer),
		stopped: make(chan struct{}),
	}
}

func (s *Scheduler) Run() {
	s.queue.Init()
	go s.consume()
	for {
		select {
		case task := <-s.in:
			s.mu.Lock()
			s.queue.Push(task)
			s.mu.Unlock()
		default:
			s.mu.Lock()
			if s.queue.Len() > 0 {
				task := s.queue.Pop()
				s.mu.Unlock()
				s.out <- task
			} else {
				s.mu.Unlock()
				time.Sleep(10 * time.Millisecond) // 防止空转
			}
		}
	}
}

func (s *Scheduler) Submit(task tasks.CrawlTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue.Push(task)
}

func (s *Scheduler) consume() {
	for {
		select {
		case task := <-s.out:
			crawler := s.workers.Alloc()
			finish := make(chan bool, 1)
			go crawler.Start(task, finish)
			<-finish
			s.workers.Free(crawler)
		}
	}
}
