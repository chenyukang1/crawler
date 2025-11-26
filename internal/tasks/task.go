package tasks

import (
	"net/http"
	"time"
)

type CrawlTask struct {
	Url           string        //目标URL，必须设置
	Method        string        //GET POST POST-M HEAD
	Header        http.Header   //请求头信息
	EnableCookie  bool          //是否使用Cookie
	PostData      string        //POST values
	DialTimeout   time.Duration //创建连接超时 dial tcp: i/o timeout
	ConnTimeout   time.Duration //连接状态超时 WSARecv tcp: i/o timeout
	Retry         Retry         // 重试策略
	RedirectTimes int           //重定向的最大次数，-1为不限制次数
	Priority      int           //指定调度优先级，默认为0（最小优先级为0）
	Reloadable    bool          //是否允许重复该链接下载
	RuleName      string        //解析规则名称

	proxy string //当用户界面设置可使用代理IP时，自动设置代理
}

func DefaultCrawlTask(url string) CrawlTask {
	return CrawlTask{
		Url:         url,
		Method:      DEFAULT_METHOD,
		DialTimeout: time.Second,
		ConnTimeout: time.Second,
		Retry: &BackoffRetry{
			ReTryTimes: 3,
			Interval:   time.Second,
		},
		RedirectTimes: -1,
		Priority:      0,
		Reloadable:    false,
	}
}

const (
	DEFAULT_METHOD = "GET"
)
