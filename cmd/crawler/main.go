package main

import (
	"github.com/chenyukang1/crawler/internal/logger"
	"github.com/chenyukang1/crawler/internal/scheduler"
)

func main() {
	logger.Info("Start crawler...")
	scheduler.NewScheduler().Run()
}
