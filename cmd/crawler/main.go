package main

import (
	global "github.com/chenyukang1/crawler/internal/app"
	"github.com/chenyukang1/crawler/pkg/log"
)

func main() {
	log.Info("Start crawler...")
	global.Get().Scheduler.Run()
	// process.GlobalScheduler.Submit(process.DefaultCrawlTask("https://m.douban.com"))
}
