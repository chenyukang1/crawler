package main

import (
	"github.com/chenyukang1/crawler/internal/logger"
	"github.com/chenyukang1/crawler/internal/scheduler"
	"github.com/chenyukang1/crawler/internal/tasks"
)

func main() {
	logger.Info("Start crawler...")
	scheduler.Global.Run()
	scheduler.Global.Submit(tasks.DefaultCrawlTask("https://m.douban.com"))
}
