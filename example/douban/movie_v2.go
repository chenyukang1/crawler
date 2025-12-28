package main

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
)

func main() {
	scheduler := process.GlobalScheduler
	spider := spider.Spider{
		Name:        "豆瓣电影网",
		Description: "豆瓣电影爬虫",
		Rules: map[string]*spider.Rule{
			"Home": {
				Name: "解析首页",
				Run: func(ctx *spider.Context) {
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("解析 dom 出错: %v", err)
					}
					var path string
					dom.Find(".screening-hd").Each(func(i int, s *goquery.Selection) {
						h2 := s.Find("h2")
						targetText := strings.TrimSpace(h2.Contents().Not("span").Text())
						if targetText == "正在热映" {
							href, exists := h2.Find("a").First().Attr("href")
							if exists {
								path = href
							} else {
								log.Warnf("%s href not found", ctx.Url)
							}
							return
						}
					})
					u, err := url.Parse(ctx.Url)
					if err != nil {
						log.Errorf("解析url失败 %v", err)
					}
					movieUrl, err := url.JoinPath(ctx.Url, path)
					if err != nil {
						log.Errorf("join path失败 %v", err)
					}
                    log.Infof("movie url: %s", movieUrl)
					header := make(http.Header)
					header.Add("Referer", u.String())
					nextTask := process.DefaultCrawlTask(movieUrl, "豆瓣电影网", "NowPlaying")
					nextTask.Header = header
					nextTask.DialTimeout = 5 * time.Second
					nextTask.ConnTimeout = 5 * time.Second
					scheduler.Submit(nextTask)
				},
			},
			"NowPlaying": {
				Name: "正在上映",
				Run: func(ctx *spider.Context) {
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("dom 解析失败", err)
					}
                    parseMovie := func (i int, s *goquery.Selection) {
						type Movie struct {
							title    string
							score    string
							star     string
							release  string
							duration string
							region   string
							director string
							actors   string
							category string
						}
                        movie := Movie{
                            title: s.AttrOr("data-title", ""),
                            score: s.AttrOr("data-score", ""),
                            star: s.AttrOr("data-star", ""),
                            release: s.AttrOr("data-release", ""),
                            duration: s.AttrOr("data-duration", ""),
                            region: s.AttrOr("data-region", ""),
                            director: s.AttrOr("data-director", ""),
                            actors: s.AttrOr("data-actors", ""),
                            category: s.AttrOr("data-category", ""),
                        }
                        structuredData := collect.NewDataCell()
                        structuredData.Set("标题", movie.title)
                        structuredData.Set("评分", movie.score)
                        structuredData.Set("收藏", movie.star)
                        structuredData.Set("发布时间", movie.release)
                        structuredData.Set("时长", movie.duration)
                        structuredData.Set("地区", movie.region)
                        structuredData.Set("分类", movie.category)
                        structuredData.Set("导演", movie.director)
                        structuredData.Set("演员", movie.actors)
                        ctx.StructuredData = append(ctx.StructuredData, structuredData)
                    }
					dom.Find("#nowplaying .list-item").Each(parseMovie)
					dom.Find("#upcoming .list-item").Each(parseMovie)
				},
			},
		},
	}

	scheduler.Register(&spider)
	scheduler.Run()
	task := process.DefaultCrawlTask("https://movie.douban.com", "豆瓣电影网", "Home")
	task.ConnTimeout = 5 * time.Second
	task.DialTimeout = 5 * time.Second
	scheduler.Submit(task)
	scheduler.Wait()
}
