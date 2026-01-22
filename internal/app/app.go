package global

import (
	"sync"

	"github.com/chenyukang1/crawler/internal/process"
)

type App struct {
	Scheduler *process.Scheduler
	Pool      *process.CrawlerPool
	TaskQueue process.TaskQueue
	once      sync.Once
}

var container *App

func Get() *App {
	container.once.Do(func() {
		container.Scheduler = process.NewScheduler()
		container.Pool = process.NewCrawlerPool(10)
		container.TaskQueue = process.NewTaskQueue()
	})
	return container
}
