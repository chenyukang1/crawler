package main

import (
	crawler "github.com/chenyukang1/crawler/internal"
	"github.com/chenyukang1/crawler/pkg/log"
)

func main() {
	log.Info("Start crawler...")
	app := crawler.Get()
	app.Run()
}
