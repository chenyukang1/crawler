package scheduler

import (
	"sync"
	"time"
)

type Scheduler struct {
	pq      TaskQueue
	in      chan CrawlTask // 生产者输入
	out     chan CrawlTask // 分发给 worker
	stopped chan struct{}
	mu      sync.Mutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		pq:      queue,
		in:      make(chan CrawlTask, 100),
		out:     make(chan CrawlTask, 100),
		stopped: make(chan struct{}),
	}
}

func (s *Scheduler) Run() {
	s.pq.Init()
	for {
		select {
		case task := <-s.in:
			s.mu.Lock()
			s.pq.Push(task)
			s.mu.Unlock()
		default:
			s.mu.Lock()
			if s.pq.Len() > 0 {
				task := s.pq.Pop()
				s.mu.Unlock()
				s.out <- task
			} else {
				s.mu.Unlock()
				time.Sleep(10 * time.Millisecond) // 防止空转
			}
		}
	}
}

func (s *Scheduler) Push(task CrawlTask) {
	s.mu.Lock()
	s.pq.Push(task)
	s.mu.Unlock()
}
