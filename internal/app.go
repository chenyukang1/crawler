package crawler

import (
	"context"
	"net/http"
	"sync"

	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
)

type App struct {
	Pool      *process.CrawlerPool
	TaskQueue process.TaskQueue

	config *Config
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	container *App
	once      sync.Once
	wg        sync.WaitGroup
)

func Get() *App {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		taskQueue := process.NewTaskQueue(ctx)
		pool := process.NewCrawlerPool(ctx, Conf.Crawler.Worker)
		taskQueue.Init()

		container = &App{
			Pool:      pool,
			TaskQueue: taskQueue,
			config:    Conf,
			ctx:       ctx,
			cancel:    cancel,
		}
	})
	return container
}

func (a *App) Run() {
	log.Info("start crwal...")
	if len(spider.GlobalRegistry) == 0 {
		panic("no spider registered, register spiders first")
	}
	go a.observe()
	a.run()
	wg.Wait()
}

func (a *App) Submit(t *process.CrawlTask) {
	a.TaskQueue.Push(t)
}

func (a *App) Stop() {
	a.cancel()
}

func (a *App) run() {
	for i := 0; i < a.config.Crawler.Parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := a.Pool.Alloc(a.TaskQueue, a.config.Crawler.IdleTime)
			defer a.Pool.Free(c)
			c.Start()
		}()
	}
}

func (a *App) observe() {
	err := http.ListenAndServe("localhost:6060", nil)
	if err != nil {
		log.Errorf("start at 6060 fail, %v", err)
		panic(err)
	}
}
