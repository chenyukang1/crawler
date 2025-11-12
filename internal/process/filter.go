package process

import (
	"fmt"
	"github.com/chenyukang1/crawler/internal/logger"
	"github.com/temoto/robotstxt"
	"net/http"
	"net/url"
)

type Filter interface {
	DoFilter(url string) bool
	CanCrawl(url string) bool
}

type DefaultFilter struct {
	bloomFilter *BloomFilter
}

func (f *DefaultFilter) DoFilter(url string) bool {
	if f.bloomFilter.Contains(url) {
		return false
	}
	f.bloomFilter.Add(url)
	return true
}

func (f *DefaultFilter) CanCrawl(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		logger.Errorf("非法的URL!!! %s", rawURL)
		return false
	}
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", u.Scheme, u.Host)
	resp, err := http.Get(robotsURL)
	if err != nil {
		logger.Errorf("robots URL访问失败!!! %v", err)
		// 无法访问robots.txt，默认允许
		return true
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return true
	}
	data, err := robotstxt.FromResponse(resp)
	if err != nil {
		logger.Errorf("robots.txt 规则解析失败!!! %v", err)
		return true
	}
	group := data.FindGroup("*")
	return group.Test(u.Path + "?" + u.RawQuery)
}
