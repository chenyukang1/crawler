/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	crawler "github.com/chenyukang1/crawler/internal"
	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/chenyukang1/crawler/pkg/retry"
	"github.com/chenyukang1/crawler/recipes"
)

var (
	mode        string
	recipe      string
	goroutine   int
	worker      int
	maxIdleTime time.Duration
	dialTimeout time.Duration
	connTimeout time.Duration

	validModes = []string{"config", "recipe"}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crawler",
	Short: "A high performance crawler",
	Long:  "crawler is a high performance crawler framework that helps you crawl easily.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if slices.Contains(validModes, mode) {
			return nil
		}
		return fmt.Errorf("invalid mode %s", mode)
	},
	Run: func(cmd *cobra.Command, args []string) {
		switch mode {
		case "config":
			options := make([]process.Option, 0)
			if cmd.Flags().Changed("dialTimeout") {
				options = append(options, process.WithDailTimeout(dialTimeout))
			}
			if cmd.Flags().Changed("connTimeout") {
				options = append(options, process.WithConnTimeout(connTimeout))
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
		case "recipe":
			r := recipes.Get(recipe)
			if r == nil {
				fmt.Printf("Recipe %s not found, try again.\n", recipe)
				os.Exit(1)
			}
			r.Run(crawler.Get(), &spider.GlobalRegistry)
		}

		crawler.Get().Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func InitRoot() {
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "config", "Start crwal from which mode, support [config | recipe]")
	rootCmd.Flags().StringVarP(&recipe, "run", "r", "", "The specified recipe hard-coded.")
	rootCmd.Flags().StringVar(&crawler.Cfgfile, "config", "./config/config.yaml", "Start crwal from specied config file.")
	rootCmd.Flags().IntVar(&goroutine, "goroutine", 10, "The maximum groutines to use.")
	rootCmd.Flags().IntVar(&worker, "worker", 10, "The maximum groutines to use.")
	rootCmd.Flags().DurationVar(&maxIdleTime, "maxIdleTime", 3*time.Second, "The maximum idle time for a crawler to finish.")
	rootCmd.Flags().DurationVar(&dialTimeout, "dialTimeout", 3*time.Second, "The total amount of time to wait for a HTTP connection.")
	rootCmd.Flags().DurationVar(&connTimeout, "connTimeout", 3*time.Second, "The maximum amount of time to wait for a TCP connection to be established (including DNS lookup and the three-way handshake).")

	viper.BindPFlag("parallelism", rootCmd.Flags().Lookup("goroutine"))
	viper.BindPFlag("worker", rootCmd.Flags().Lookup("worker"))
	viper.BindPFlag("maxIdleTime", rootCmd.Flags().Lookup("maxIdleTime"))

	cobra.OnInitialize(crawler.ReadConfig)
}

// parseSpiders parse conf.Spiders to spdier.Spiders
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
			Name:  k,
			Rules: parsedRules,
		})
	}

	return spiders
}

// parseTasks parse conf tasks to process.CrawlTask
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
			DialTimeout:  time.Duration(t.DialTimeout) * time.Second,
			ConnTimeout:  time.Duration(t.ConnTimeout) * time.Second,
			Retry: &retry.FixedRetry{
				ReTryTimes: t.Retry.Times,
				Interval:   time.Duration(t.Retry.Interval) * time.Second,
			},
			RedirectTimes: t.RedirectTimes,
			Priority:      t.Priority,
			Reloadable:    false,
			SpiderName:    t.Spider,
			RuleName:      t.Rule,
			ShouldFilter:  false,
		}
		task.WithOptions(opts...)
		tasks = append(tasks, task)
	}
	return tasks
}
