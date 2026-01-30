/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	crawler "github.com/chenyukang1/crawler/internal"
	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/chenyukang1/crawler/pkg/retry"
	"github.com/spf13/cobra"
)

var (
	dialTimeout int
	connTimeout int
)

// crawlCmd represents the crawl command
var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "start crawl",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("start crawl...")
		options := make([]process.Option, 0)
		if cmd.Flags().Changed("dialTimeout") {
			options = append(options, process.WithDailTimeout(time.Duration(dialTimeout)))
		}
		if cmd.Flags().Changed("connTimeout") {
			options = append(options, process.WithConnTimeout(time.Duration(connTimeout)))
		}

		app := crawler.Get()
		spiders := parseSpiders(crawler.Conf)
		for _, s := range spiders {
			spider.GlobalRegistry.Register(s)
		}

		tasks := parseTasks(crawler.Conf, options...)
		for _, task := range tasks {
			app.Submit(task)
		}

		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(crawlCmd)
	crawlCmd.PersistentFlags().IntVar(&connTimeout, "connTimeout", 3, "The maximum amount of time to wait for a TCP connection to be established (including DNS lookup and the three-way handshake). Default 3s.")
	crawlCmd.PersistentFlags().IntVar(&dialTimeout, "dialTimeout", 3, "The total amount of time to wait for a HTTP connection. Default is 3s.")
}

func parseSpiders(conf *crawler.Config) []*spider.Spider {
	spiders := make([]*spider.Spider, 0)
	for k, v := range conf.Spiders {
		parsedRules := make(map[string]*spider.Rule)
		for n, rules := range v.Rules {
			r := &spider.Rule{
				Name: n,
				Run: func(ctx *spider.Context) {
					dom, err := ctx.GetDom()
					if err != nil {
						log.Errorf("parse dom fail %v", err)
					}
					for _, p := range rules {
						dom.Find(p.Selector).Each(func(i int, s *goquery.Selection) {
							for pk, pv := range p.Fields {
								data := collect.NewDataCell()
								data.Set(pk, s.AttrOr(pv, ""))
								ctx.StructuredData = append(ctx.StructuredData, data)
							}
						})
					}
				},
			}
			parsedRules[n] = r
		}
		spiders = append(spiders, &spider.Spider{
			Name:      k,
			EntryRule: v.Entry,
			Rules:     parsedRules,
		})
	}

	return spiders
}

func parseTasks(conf *crawler.Config, opts ...process.Option) []*process.CrawlTask {
	tasks := make([]*process.CrawlTask, 0)
	for _, t := range conf.Tasks {
		headers := make(http.Header)
		for k, v := range t.Headers {
			headers.Add(k, v)
		}
		task := &process.CrawlTask{
			URL:          t.URL,
			Method:       t.Method,
			Header:       headers,
			EnableCookie: t.EnableCookie,
			PostData:     "",
			DialTimeout:  time.Duration(t.DialTimeout),
			ConnTimeout:  time.Duration(t.ConnTimeout),
			Retry: &retry.FixedRetry{
				ReTryTimes: t.Retry.Times,
				Interval:   time.Duration(t.Retry.Interval),
			},
			RedirectTimes: t.RedirectTimes,
			Priority:      t.Priority,
			Reloadable:    false,
			SpiderName:    t.Spider,
			RuleName:      t.Rule,
			ShouldFilter:  false,
		}
		task.WithOptions(opts...)
		tasks = append(tasks)
	}
	return tasks
}
