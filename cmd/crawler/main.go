package main

import (
	global "github.com/chenyukang1/crawler/internal/app"
	"github.com/chenyukang1/crawler/pkg/log"
)

func main() {
	log.Info("Start crawler...")
	app := global.Get()
	app.Run()
	// process.GlobalScheduler.Submit(process.DefaultCrawlTask("https://m.douban.com"))
}
