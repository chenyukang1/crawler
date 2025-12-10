package main

import (
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/pkg/log"
)

func main() {
	log.Info("Start crawler...")
	process.GlobalScheduler.Run()
	//process.GlobalScheduler.Submit(process.DefaultCrawlTask("https://m.douban.com"))
}
