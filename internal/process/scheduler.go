package process

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
)

type Scheduler struct {
	queue   TaskQueue
	pool    *CrawlerPool
	spiders map[string]*spider.Spider
	ctx     context.Context
	cancel  context.CancelFunc

	// wg   sync.WaitGroup
	mu sync.RWMutex
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		queue:   NewTaskQueue(),
		spiders: make(map[string]*spider.Spider),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *Scheduler) Run() {
	if len(s.spiders) == 0 {
		log.Errorf("没有爬虫规则注册！！")
		return
	}
	s.queue.Init()
	go func() {
		err := http.ListenAndServe("localhost:6060", nil)
		if err != nil {
			log.Errorf("启动在本地6060端口失败! %v", err)
			return
		}
	}()
	go s.run()
}

func (s *Scheduler) Submit(task *CrawlTask) {
	s.queue.Push(task)
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.pool.Stop()
}

func (s *Scheduler) run() {
	n := len(s.spiders)
	s.pool = NewCrawlerPool(n)
	crawler := s.pool.Alloc(s.ctx)
	go func() {
		defer func() {
			s.pool.Free(crawler)
		}()
		crawler.Start()
	}()
}
