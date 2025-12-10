package main

import (
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
)

func main() {
	log.Info("开始爬虫任务...")
	scheduler := process.GlobalScheduler
	spider1 := &spider.Spider{
		Name:        "豆瓣网",
		Description: "解析首页",
		Rules: map[string]*spider.Rule{
			"Home": {
				Name: "解析首页",
				Run: func(ctx *spider.Context) {
					html, err := ctx.GetHtml()
					if err != nil {
						log.Errorf("解析 html 出错: %v", err)
					}
					log.Infof("首页html: %s", html)
				},
			},
		},
	}
	scheduler.Register(spider1)
	scheduler.Run()

	scheduler.Submit(process.DefaultCrawlTask("https://m.douban.com", "豆瓣网", "Home"))
	scheduler.Wait()
}
