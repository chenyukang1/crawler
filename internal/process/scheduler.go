package process

import (
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

// GlobalScheduler 全局唯一实例
var GlobalScheduler = NewScheduler()

type Scheduler struct {
	queue    TaskQueue
	crawlers *CrawlerPool
	spiders  map[string]*spider.Spider
	route    map[string]chan *CrawlTask

	wg   sync.WaitGroup
	mu   sync.RWMutex
	stop chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		queue:   NewTaskQueue(),
		spiders: make(map[string]*spider.Spider),
		route:   make(map[string]chan *CrawlTask),
		stop:    make(chan struct{}, 1),
	}
}

func (s *Scheduler) Register(spider *spider.Spider) {
	s.spiders[spider.Name] = spider
	s.route[spider.Name] = make(chan *CrawlTask, 100)
}

func (s *Scheduler) Run() {
	if s.spiders == nil || len(s.spiders) == 0 {
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

func (s *Scheduler) Fetch(spiderName string) *CrawlTask {
	select {
	case task := <-s.route[spiderName]:
		return task
	case <-time.After(time.Second):
		return nil
	}
}

func (s *Scheduler) Stop() {
	s.crawlers.Stop()
	s.stop <- struct{}{}
}

func (s *Scheduler) run() {
	n := len(s.spiders)
	s.crawlers = NewCrawlerPool(n)
	//s.wg.Add(n)
	for _, v := range s.spiders {
		crawler := s.crawlers.Alloc(v)
		go func() {
			defer func() {
				s.crawlers.Free(crawler)
				//s.wg.Done()
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
