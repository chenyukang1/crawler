package parse

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chenyukang1/crawler/internal/logger"
	"net/http"
	"regexp"
	"strings"
)

type Parser struct {
	Filter *BloomFilter
}

var Default = &Parser{
	Filter: new(BloomFilter),
}

func (p *Parser) Parse(request *http.Request, response *http.Response) *context.Context {
	sem := make(chan int, p.Config.Concurrency)
	for i := 0; i < cap(sem); i++ {
		sem <- 1
	}
	for resp := range p.RespList {
		<-sem
		go func(resp model.FetchResp) {
			p.doParse(resp)
			sem <- 1
		}(resp)
	}
}

var titleReg = regexp.MustCompile(`<title>(.*?)</title>`)

func (p *Parser) doParse(resp model.FetchResp) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.Content))
	if err != nil {
		return
	}

	// 根据 CSS 选择器找到 DOM 节点
	doc.Find(".media__body .title a").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())
		logger.Infof("Found movie name %s", name)
		fmt.Println(i+1, name)
	})

	var urls []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if url, ok := s.Attr("url"); ok {
			if !p.Filter.Contains(url) {
				logger.Infof("Found url %s", url)
				p.Filter.Add(url)
				urls = append(urls, url)
			}
		}
	})
	for _, url := range urls {
		go func(url string) {
			p.TaskList <- model.FetchTask{Url: url}
		}(url)
	}
}
