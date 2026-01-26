package process

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chenyukang1/crawler/internal/collect"
	"github.com/chenyukang1/crawler/internal/fetch"
	"github.com/chenyukang1/crawler/internal/filter"
	"github.com/chenyukang1/crawler/internal/spider"
	"github.com/chenyukang1/crawler/internal/status"
	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/chenyukang1/crawler/pkg/retry"
	"github.com/chenyukang1/crawler/pkg/utils"
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

	seq      int
	queue    TaskQueue
	fetcher  *fetch.Fetcher
	filter   filter.Filter
	idle     int
	idleTime int
	status   int // 执行状态
	ctx      context.Context
	lock     sync.RWMutex
}

type CrawlTask struct {
	URL           string        // 目标URL，必须设置
	Method        string        // GET POST POST-M HEAD
	Header        http.Header   // 请求头信息
	EnableCookie  bool          // 是否使用Cookie
	PostData      string        // POST values
	DialTimeout   time.Duration // 创建连接超时 dial tcp: i/o timeout
	ConnTimeout   time.Duration // 连接状态超时 WSARecv tcp: i/o timeout
	Retry         retry.Retry   // 重试策略
	RedirectTimes int           // 重定向的最大次数，-1为不限制次数
	Priority      int           // 指定调度优先级，默认为0（最小优先级为0）
	Reloadable    bool          // 是否允许重复该链接下载
	SpiderName    string        // spider名称
	RuleName      string        // 解析规则名称
	ShouldFilter  bool          // 是否过滤

	proxy string // 当用户界面设置可使用代理IP时，自动设置代理
}

var userAgents = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
}

const (
	defaultMethod = "GET"
)

func DefaultCrawlTask(url string, spider string, rule string) *CrawlTask {
	return &CrawlTask{
		URL:         url,
		Method:      defaultMethod,
		DialTimeout: 2 * time.Second,
		ConnTimeout: 3 * time.Second,
		Retry: &retry.BackoffRetry{
			ReTryTimes: 3,
			Interval:   2 * time.Second,
		},
		RedirectTimes: -1,
		Priority:      0,
		Reloadable:    false,
		SpiderName:    spider,
		RuleName:      rule,
		ShouldFilter:  true,
	}
}

func (c *CrawlTask) BuildRequest() (req *fetch.Request, err error) {
	req = &fetch.Request{}
	req.Url, err = utils.UrlEncode(c.URL)

	req.Header = c.Header
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	switch method := strings.ToUpper(c.Method); method {
	case "GET":
		req.Method = method
	case "POST":
		req.Method = method
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Body = strings.NewReader(c.PostData)
	default:
		req.Method = "GET"
	}
	req.Header.Add("User-Agent", userAgents[rand.Intn(len(userAgents))])
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Add("Connection", "keep-alive")

	req.ConnTimeout = c.ConnTimeout
	req.DialTimeout = c.DialTimeout
	req.RedirectTimes = c.RedirectTimes
	req.EnableCookie = c.EnableCookie
	req.Reloadable = c.Reloadable
	req.Retry = c.Retry
	return
}

func NewCrawler(c context.Context, s int, q TaskQueue, t int) *Crawler {
	return &Crawler{
		Collector: collect.NewLogCollector(10),
		seq:       s,
		queue:     q,
		fetcher:   fetch.Default,
		filter:    filter.GlobalFilter,
		idleTime:  t,
		status:    status.INITIAL,
		ctx:       c,
	}
}

func (c *Crawler) Start() {
	c.setStatus(status.RUN)

	c.run()

	c.setStatus(status.STOPPED)
}

func (c *Crawler) Pause() {
	c.setStatus(status.PAUSE)
}

func (c *Crawler) Stop() {
	if c.Status() == status.STOP {
		return
	}
	c.Collector.Finish()
	c.setStatus(status.STOP)
}

func (c *Crawler) Status() int {
	c.lock.RLock()
	s := c.status
	c.lock.RUnlock()
	return s
}

func (c *Crawler) run() {
	log.Infof("[crawler-%d] start...", c.seq)
	c.idle = 0
	for c.idle < c.idleTime {
		select {
		case <-c.ctx.Done():
			log.Infof("[crawler-%d] cancel, reason: %v", c.seq, c.ctx.Err())
			return
		default:
		}

		if c.isPause() {
			time.Sleep(time.Second)
			continue
		}

		// 开始收集数据
		go c.Collector.Pipeline(c.seq)

		var (
			task     *CrawlTask
			ok       bool
			ctx      *spider.Context
			request  *http.Request
			response *http.Response
		)

		task, ok = c.queue.Pop()
		if !ok {
			log.Infof("[crawler-%d] cancel", c.seq)
		} else if task == nil {
			log.Infof("[crawler-%d] no more task, sleep 1 second", c.seq)
			c.idle++
			time.Sleep(time.Second)
			continue
		}

		log.Infof("[crawler-%d] fetched task", c.seq)

		if task.ShouldFilter && !c.filter.DoFilter(task.URL) {
			log.Errorf("【%s】重复Url！！", task.URL)
			continue
		}
		if !c.filter.CanCrawl(task.URL) {
			log.Warnf("【%s】该Url不可爬！！", task.URL)
		}

		req, err := task.BuildRequest()
		if err != nil {
			log.Errorf("【%s】Url解析失败, %v ", task.URL, err)
			continue
		}

		request, response, err = c.fetcher.Fetch(c.ctx, req)
		if err != nil {
			log.Errorf("【%s】Url访问失败, %v ", task.URL, err)
			continue
		}

		s, err := spider.GlobalRegistry.GetSpider(task.SpiderName)
		if err != nil {
			log.Errorf("【%s】规则获取失败, %v", err)
		}
		ctx = &spider.Context{
			Spider:         s,
			URL:            task.URL,
			Request:        request,
			Response:       response,
			StructuredData: make([]collect.DataCell, 0),
		}
		if err = ctx.Rule(task.RuleName); err != nil {
			log.Errorf("【%s】Url【%s】规则解析失败 %v ", task.URL, task.RuleName, err)
			continue
		}
		for _, cell := range ctx.StructuredData {
			c.Collector.Push(cell)
		}
		c.Collector.Finish()
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
