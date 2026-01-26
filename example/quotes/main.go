// Package quotes
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	crawler "github.com/chenyukang1/crawler/internal"
	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/chenyukang1/crawler/pkg/retry"
)

func main() {
	app := crawler.Get()
	s := &spider.Spider{
		Name:        "quotes",
		Description: "quotes测试",
		Rules: map[string]*spider.Rule{
			"Login": {
				Name: "登录页面",
				Run: func(ctx *spider.Context) {
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("get dom fail for url %s, %v", ctx.URL, err)
					}
					csrfToken, exists := dom.Find("input[name='csrf_token']").Attr("value")
					if !exists {
						log.Error("csrf token not found")
						panic("csrf token not found")
					}

					postData := url.Values{}
					postData.Set("username", "admin")
					postData.Set("password", "123456")
					postData.Set("csrf_token", csrfToken)

					headers := make(http.Header)
					headers.Add("Origin", "https://quotes.toscrape.com")
					headers.Add("Referer", "https://quotes.toscrape.com/login")
					headers.Add("Content-Type", "application/x-www-form-urlencoded")

					task := process.CrawlTask{
						URL:          "https://quotes.toscrape.com/login",
						Method:       "POST",
						Header:       headers,
						EnableCookie: true,
						PostData:     postData.Encode(),
						DialTimeout:  time.Second,
						ConnTimeout:  3 * time.Second,
						Retry: &retry.BackoffRetry{
							ReTryTimes: 3,
							Interval:   time.Second,
						},
						RedirectTimes: -1,
						SpiderName:    "quotes",
						RuleName:      "Home",
						ShouldFilter:  false,
					}

					app.Submit(&task)
				},
			},
			"Home": {
				Name: "登录页面",
				Run: func(ctx *spider.Context) {
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("get dom fail for url %s, %v", ctx.URL, err)
					}
					dom.Find(".row .quote").Each(func(i int, s *goquery.Selection) {
						data := collect.NewDataCell()
						data.Set("text", s.Find("span.text").Text())
						ctx.StructuredData = append(ctx.StructuredData, data)
					})
					for i := range 10 {
						newURL := fmt.Sprintf("https://%s/page/%d", "quotes.toscrape.com", i+1)
						task := process.DefaultCrawlTask(newURL, "quotes", "Page")
						app.Submit(task)
					}
				},
			},
			"Page": {
				Name: "分页页面",
				Run: func(ctx *spider.Context) {
					log.Info("let's parse the page")
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("get dom fail for url %s, %v", ctx.URL, err)
						return
					}
					dom.Find(".row .quote").Each(func(i int, s *goquery.Selection) {
						data := collect.NewDataCell()
						data.Set("text", s.Find("span.text").Text())
						ctx.StructuredData = append(ctx.StructuredData, data)
					})
				},
			},
		},
		EntryRule: "Login",
	}

	task := process.DefaultCrawlTask("https://quotes.toscrape.com/login", "quotes", "Login")
	task.EnableCookie = true

	if err := spider.GlobalRegistry.Register("quotes", s); err != nil {
		panic(err)
	}

	app.Submit(task)
	app.Run()
}
