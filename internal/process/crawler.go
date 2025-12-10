package process

import (
	"context"
	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/internal/fetch"
	"github.com/chenyukang1/crawler/internal/filter"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/internal/status"
	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/chenyukang1/crawler/pkg/retry"
	"github.com/chenyukang1/crawler/pkg/utils"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ICrawler interface {
	Start()
	Pause()
	Stop()
	CanStop() bool
	Status() int
}

type Crawler struct {
	Collector collect.Collector

	spider   *spider.Spider //解析规则
	fetcher  *fetch.Fetcher
	filter   filter.Filter
	status   int           //执行状态
	finish   chan int      //结束channel
	idleTime time.Duration //空闲时间，超过被回收
	lock     sync.RWMutex
}

type CrawlTask struct {
	Url           string        //目标URL，必须设置
	Method        string        //GET POST POST-M HEAD
	Header        http.Header   //请求头信息
	EnableCookie  bool          //是否使用Cookie
	PostData      string        //POST values
	DialTimeout   time.Duration //创建连接超时 dial tcp: i/o timeout
	ConnTimeout   time.Duration //连接状态超时 WSARecv tcp: i/o timeout
	Retry         retry.Retry   // 重试策略
	RedirectTimes int           //重定向的最大次数，-1为不限制次数
	Priority      int           //指定调度优先级，默认为0（最小优先级为0）
	Reloadable    bool          //是否允许重复该链接下载
	SpiderName    string        //spider名称
	RuleName      string        //解析规则名称

	proxy string //当用户界面设置可使用代理IP时，自动设置代理
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
}

const (
	DEFAULT_METHOD = "GET"
)

func DefaultCrawlTask(url string, spider string, rule string) *CrawlTask {
	return &CrawlTask{
		Url:         url,
		Method:      DEFAULT_METHOD,
		DialTimeout: time.Second,
		ConnTimeout: time.Second,
		Retry: &retry.BackoffRetry{
			ReTryTimes: 3,
			Interval:   time.Second,
		},
		RedirectTimes: -1,
		Priority:      0,
		Reloadable:    false,
		SpiderName:    spider,
		RuleName:      rule,
	}
}

func (c *CrawlTask) BuildRequest() (req *fetch.Request, err error) {
	req = &fetch.Request{}
	req.Url, err = utils.UrlEncode(c.Url)

	req.Header = c.Header
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	switch method := strings.ToUpper(c.Method); method {
	case "GET":
		req.Method = method
	case "POST":
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Body = strings.NewReader(c.PostData)
	default:
		req.Method = "GET"
	}
	req.Header.Add("User-Agent", userAgents[rand.Intn(len(userAgents))])

	req.ConnTimeout = c.ConnTimeout
	req.DialTimeout = c.DialTimeout
	req.RedirectTimes = c.RedirectTimes
	req.EnableCookie = c.EnableCookie
	req.Reloadable = c.Reloadable
	req.Retry = c.Retry
	return
}

func NewCrawler(spider *spider.Spider) *Crawler {
	return &Crawler{
		Collector: collect.Log,
		spider:    spider,
		fetcher:   fetch.Default,
		filter:    filter.GlobalFilter,
		status:    status.INITIAL,
		finish:    make(chan int, 1),
		idleTime:  60 * time.Second,
	}
}

func (c *Crawler) Start() {
	c.setStatus(status.RUN)

	// 开始收集数据
	go c.Collector.Pipeline()
	go func() {
		c.run()
		c.finish <- 1
	}()
	<-c.finish
	c.setStatus(status.STOPPED)
}

func (c *Crawler) Pause() {
	c.setStatus(status.PAUSE)
}

func (c *Crawler) Stop() {
	if c.Status() == status.STOP {
		return
	}
	if c.CanStop() {
		c.Collector.Stop()
		close(c.finish)
		c.setStatus(status.STOP)
	}
}

func (c *Crawler) CanStop() bool {
	return true
}

func (c *Crawler) Status() int {
	c.lock.RLock()
	s := c.status
	c.lock.RUnlock()
	return s
}

func (c *Crawler) run() {
	idled := 0 * time.Second
loop:
	for idled < c.idleTime {
		if c.isPause() {
			time.Sleep(time.Second)
			goto loop
		}
		var (
			task     *CrawlTask
			ctx      *spider.Context
			request  *http.Request
			response *http.Response
		)
		task = GlobalScheduler.Fetch(c.spider.Name)
		if task == nil {
			time.Sleep(time.Second)
			idled++
			continue
		}
		if !c.filter.DoFilter(task.Url) {
			log.Errorf("【%s】重复Url！！", task.Url)
			continue
		}
		if !c.filter.CanCrawl(task.Url) {
			log.Warnf("【%s】该Url不可爬！！", task.Url)
		}

		req, err := task.BuildRequest()
		if err != nil {
			log.Errorf("【%s】Url解析失败, %v ", task.Url, err)
			continue
		}

		request, response, err = c.fetcher.Fetch(context.Background(), req)
		if err != nil {
			log.Errorf("【%s】Url访问失败, %v ", task.Url, err)
			continue
		}
		ctx = &spider.Context{
			Spider:   c.spider,
			Url:      task.Url,
			Request:  request,
			Response: response,
		}
		if err = ctx.Rule(task.RuleName); err != nil {
			log.Errorf("【%s】Url【%s】规则解析失败 %v ", task.Url, task.RuleName, err)
			continue
		}
		for _, cell := range ctx.StructuredData {
			c.Collector.Push(cell)
		}
		idled = 0
	}
}

func (c *Crawler) isPause() bool {
	return c.Status() == status.PAUSE
}

func (c *Crawler) setStatus(status int) {
	c.lock.Lock()
	c.status = status
	c.lock.Unlock()
}
