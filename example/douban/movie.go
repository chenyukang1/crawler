package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
	"net/http"
	"net/url"
	"strings"
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
					doc, err := ctx.GetDom()
					if err != nil {
						log.Errorf("解析 dom 出错: %v", err)
					}

					var hrefs []string
					extractedData := collect.NewDataCell()
					extractedData.Set("标题", doc.Find("title").Text())
					extractedData.Set("主要文案", doc.Find(".splash-text-main").Text())
					doc.Find("#home_guide").Each(func(i int, s *goquery.Selection) {
						href, exists := s.Attr("href")
						if exists {
							hrefs = append(hrefs, href)
						}
					})

					for i, href := range hrefs {
						extractedData.Set(fmt.Sprintf("导航跳转_%d", i), href)
						scheduler.Submit(process.DefaultCrawlTask(fmt.Sprintf("%s/%s", ctx.Url, href),
							ctx.Spider.Name, "HOME_GUIDE"))
					}
					ctx.StructuredData = append(ctx.StructuredData, extractedData)
				},
			},

			"HOME_GUIDE": {
				Name: "解析导航页",
				Run: func(ctx *spider.Context) {
					doc, err := ctx.GetDom()
					targetText := "找电影"
					var path string
					doc.Find(".app-items.app-type1 > a").Each(func(i int, s *goquery.Selection) {
						h2Text := s.Find("h2").Text()
						if strings.TrimSpace(h2Text) == targetText {
							href, exists := s.Attr("href")
							if exists {
								path = href
								return
							}
						}
					})

					u, err := url.Parse(ctx.Url)
					if err != nil {
						log.Errorf("解析url失败")
					}
					movieUrl := fmt.Sprintf("%s/%s", u.Host, path)
					extractedData := collect.NewDataCell()
					extractedData.Set("标题", doc.Find("title").Text())
					extractedData.Set("找电影跳转", movieUrl)
					ctx.StructuredData = append(ctx.StructuredData, extractedData)

					nextTask := process.DefaultCrawlTask("https://m.douban.com/rexxar/api/v2/movie/modules?need_manual_chart_card=1&for_mobile=1",
						ctx.Spider.Name, "MOVIE")
					nextTask.Header = http.Header{}
					nextTask.Header.Add("Content-Type", "application/json")
					nextTask.Header.Add("Referer", "https://m.douban.com/movie/")
					scheduler.Submit(nextTask)
				},
			},

			"MOVIE": {
				Name: "热门电影",
				Run: func(ctx *spider.Context) {
					jsonData, err := ctx.GetHtml()
					if err != nil {
						log.Errorf("%s 获取网页文本失败: %v", ctx.Url, err)
					}
					var root map[string]any
					err = json.Unmarshal(jsonData, &root)
					if err != nil {
						log.Errorf("%s 解析json失败: %v", ctx.Url, err)
					}

					modules := root["modules"].([]any)
					moduleMap := make(map[string]map[string]any)
					for _, module := range modules {
						m := module.(map[string]any)
						moduleMap[m["module_name"].(string)] = m
					}

					// 尚未上映
					{
						data := moduleMap["movie_coming_soon"]["data"].([]any)
						for _, d := range data {
							m := d.(map[string]any)
							extractedData := collect.NewDataCell()
							extractedData.Set("分类", m["title"])
							ctx.StructuredData = append(ctx.StructuredData, extractedData)
							type Rating struct {
								Count     int     `json:"count"`
								Max       float64 `json:"max"`
								Value     float64 `json:"value"`
								StarCount float64 `json:"star_count"`
							}
							type Name struct {
								Name string `json:"name"`
							}
							type Item struct {
								Rating    Rating   `json:"rating"`
								Title     string   `json:"title"`
								PubDate   []string `json:"pubdate"`
								Genres    []string `json:"genres"`
								Actors    []Name   `json:"actors"`
								Directors []Name   `json:"directors"`
							}
							for _, item := range m["items"].([]any) {
								jsonText, err := json.Marshal(item)
								if err != nil {
									log.Errorf("%v 序列化失败 %v", item, err)
								}
								var itemData Item
								err = json.Unmarshal(jsonText, &itemData)
								if err != nil {
									log.Errorf("%s 反序列化失败 %v", jsonText, err)
								}
								extractedItemData := collect.NewDataCell()
								extractedItemData.Set("title", itemData.Title)
								extractedItemData.Set("pubdate", itemData.PubDate)
								extractedItemData.Set("rating", itemData.Rating.Value)
								extractedItemData.Set("genres", itemData.Genres)
								extractedItemData.Set("star_count", itemData.Rating.StarCount)
								extractedItemData.Set("directors", itemData.Directors)
								ctx.StructuredData = append(ctx.StructuredData, extractedItemData)
							}
						}
					}
				},
			},
		},
	}
	scheduler.Register(spider1)
	scheduler.Run()

	scheduler.Submit(process.DefaultCrawlTask("https://m.douban.com", "豆瓣网", "Home"))
	scheduler.Wait()
}
