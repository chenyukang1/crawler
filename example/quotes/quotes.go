// Package quotes
package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/chenyukang1/crawler/pkg/retry"
)

func main() {
	quotesSpider := &spider.Spider{
		Name:        "quotes",
		Description: "quotes测试",
		Rules: map[string]*spider.Rule{
			"Login": {
				Name: "登录页面",
				Run: func(ctx *spider.Context) {
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("get dom fail for url %s, %v", ctx.Url, err)
					}
					csrfToken, exists := dom.Find("input[name='csrf_token']").Attr("value")
					if !exists {
						log.Error("cstf token not found")
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
						Url:          "https://quotes.toscrape.com/login",
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
						Priority:      0,
						Reloadable:    false,
						SpiderName:    "quotes",
						RuleName:      "Home",
					}

					process.GlobalScheduler.Submit(&task)
				},
			},
			"Home": {
				Name: "登录页面",
				Run: func(ctx *spider.Context) {
					html, err := ctx.GetHtml()
					if err != nil {
						log.Errorf("get html fail for url %s, %v", ctx.Url, err)
					}
					log.Info(string(html))
				},
			},
		},
		EntryRule: "Login",
	}

	task := process.DefaultCrawlTask("https://quotes.toscrape.com/login", "quotes", "Login")

	scheduelr := process.GlobalScheduler
	scheduelr.Register(quotesSpider)
	scheduelr.Run()
	scheduelr.Submit(task)
	scheduelr.Wait()
}
