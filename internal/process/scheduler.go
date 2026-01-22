package process

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
)

type Scheduler struct {
	queue   TaskQueue
	pool    *CrawlerPool
	spiders map[string]*spider.Spider
	route   map[string]chan *CrawlTask
	ctx     context.Context
	cancel  context.CancelFunc

	// wg   sync.WaitGroup
	mu   sync.RWMutex
	stop chan struct{}
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		queue:   NewTaskQueue(),
		spiders: make(map[string]*spider.Spider),
		route:   make(map[string]chan *CrawlTask),
		stop:    make(chan struct{}, 1),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *Scheduler) Register(spider *spider.Spider) {
	s.spiders[spider.Name] = spider
	s.route[spider.Name] = make(chan *CrawlTask, 100)
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
	go s.dispatch()
}

func (s *Scheduler) Submit(task *CrawlTask) {
	s.queue.Push(task)
}

func (s *Scheduler) Stop() {
	s.pool.Stop()
	s.stop <- struct{}{}
}

func (s *Scheduler) run() {
	n := len(s.spiders)
	s.pool = NewCrawlerPool(n)
	for _, v := range s.spiders {
		crawler := s.pool.Alloc(v)
		go func() {
			defer func() {
				s.pool.Free(crawler)
			}()
			crawler.Start()
		}()
	}
}

func (s *Scheduler) dispatch() {
	for {
		select {
		case task := <-s.queue.Chan():
			s.mu.RLock()
			targetChan, ok := s.route[task.SpiderName]
			s.mu.RUnlock()
			if ok {
				select {
				case targetChan <- task:
					// 发送成功
				default:
					log.Warnf("Spider %s 的Channel已满, 丢弃任务", task.SpiderName)
				}
			} else {
				log.Errorf("没有Spider %s 的爬虫, 丢弃任务", task.SpiderName)
			}
		case <-s.stop:
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

func (s *Scheduler) Wait() {
	<-s.stop
}
