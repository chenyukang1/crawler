package internal

import (
	"context"
	"net/http"
	"sync"

	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/spf13/viper"
)

type App struct {
	Pool      *process.CrawlerPool
	TaskQueue process.TaskQueue

	config *Config
	ctx    context.Context
	cancel context.CancelFunc
}

type Config struct {
	Crawler struct {
		Parallelism int
	}
}

var (
	container *App
	once      sync.Once
)

func Get() *App {
	once.Do(func() {
		var conf Config
		v := viper.New()
		v.SetConfigFile("config.yaml")
		if err := v.ReadInConfig(); err != nil {
			log.Errorf("read in config fail %v", err)
			panic(err)
		}
		if err := v.Unmarshal(&conf); err != nil {
			log.Errorf("parse config fail %v", err)
			panic(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		taskQueue := process.NewTaskQueue(ctx)
		pool := process.NewCrawlerPool(ctx, conf.Crawler.Parallelism)

		container = &App{
			Pool:      pool,
			TaskQueue: taskQueue,
			config:    &conf,
			ctx:       ctx,
			cancel:    cancel,
		}
	})
	return container
}

func (a *App) Run() {
	if len(spider.GlobalRegistry) == 0 {
		panic("no spider registered, register spiders first")
	}
	a.TaskQueue.Init()
	go a.observe()
	go a.run()
	a.Pool.Wait()
}

func (a *App) Submit(t *process.CrawlTask) {
	a.TaskQueue.Push(t)
}

func (a *App) Stop() {
	a.cancel()
}

func (a *App) run() {
	for {
		crawler := a.Pool.Alloc()
		go func() {
			defer func() {
				a.Pool.Free(crawler)
			}()
			crawler.Start()
		}()
	}
}

func (a *App) observe() {
	err := http.ListenAndServe("localhost:6060", nil)
	if err != nil {
		log.Errorf("start at 6060 fail, %v", err)
		a.cancel()
	}
}
