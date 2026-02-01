package recipes

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"

	crawler "github.com/chenyukang1/crawler/internal"
	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
)

type BooksToscrape struct{}

var booksToscrape BooksToscrape

func (b *BooksToscrape) Run(app *crawler.App, registry *spider.Registry) {
	s := &spider.Spider{
		Name: "books_toscrape",
		Rules: map[string]*spider.Rule{
			"Home": {
				Name: "首页",
				Run: func(ctx *spider.Context) {
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("get dom fail %v", err)
						return
					}

					dom.Find(".side_categories .nav-list li > a").Each(func(i int, s *goquery.Selection) {
						title := strings.TrimSpace(s.Text())
						if title == "Books" {
							return
						}
						href, ok := s.Attr("href")
						if !ok {
							return
						}

						u, err := url.Parse(ctx.URL)
						if err != nil {
							log.Errorf("解析url失败 %v", err)
							return
						}
						u.Path = path.Dir(u.Path)
						categoryURL := u.JoinPath(href).String()
						categoryURL = categoryURL[:len(categoryURL)-1]

						header := make(http.Header)
						header.Add("Referer", u.String())

						task := process.DefaultCrawlTask(categoryURL, "books_toscrape", "Category")
						task.Header = header

						app.Submit(task)
					})
				},
			},
			"Category": {
				Name: "分类页",
				Run: func(ctx *spider.Context) {
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("get dom fail %v", err)
						return
					}
					u, err := url.Parse(ctx.URL)
					if err != nil {
						log.Errorf("解析url失败 %v", err)
						return
					}
					// 是否有下一页
					s := dom.Find(".next a")
					if s.Length() > 0 {
						href, ok := s.Attr("href")
						if ok {
							nextPageURL, err := url.JoinPath(ctx.URL, "..", href)
							if err == nil {
								header := make(http.Header)
								header.Add("Referer", u.String())

								pageTask := process.DefaultCrawlTask(nextPageURL, "books_toscrape", "Category")
								app.Submit(pageTask)
							}
						}
					}

					category := dom.Find(".page-header h1").Text()
					crawlCtx := make(map[string]any)
					crawlCtx["category"] = category

					// 解析详情
					dom.Find(".product_pod .image_container").Each(func(i int, s *goquery.Selection) {
						href, ok := s.Find("a").Attr("href")
						if !ok {
							return
						}
						pageURL, err := url.JoinPath(ctx.URL, "../"+href)
						if err == nil {
							header := make(http.Header)
							header.Add("Referer", u.String())

							pageTask := process.DefaultCrawlTask(pageURL, "books_toscrape", "Page", process.WithContext(crawlCtx))
							app.Submit(pageTask)
						}
					})
				},
			},
			"Page": {
				Name: "详情页",
				Run: func(ctx *spider.Context) {
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("get dom fail %v", err)
						return
					}
					book := dom.Find(".product_main").Find("h1").Text()
					desc := dom.Find("#product_description + p").Text()
					data := collect.NewDataCell()
					data.Set("book", book)
					data.Set("description", desc)
					ctx.StructuredData = append(ctx.StructuredData, data)
				},
			},
		},
	}
	if err := spider.GlobalRegistry.Register(s); err != nil {
		panic(err)
	}
	task := process.DefaultCrawlTask("https://books.toscrape.com/?", "books_toscrape", "Home")
	app.Submit(task)
}

func init() {
	registry["books_toscrape"] = &booksToscrape
}
