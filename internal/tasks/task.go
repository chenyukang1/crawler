package tasks

import (
	"net/http"
	"net/url"
	"time"
)

type CrawlTask struct {
	Url           url.URL       //目标URL，必须设置
	Method        string        //GET POST POST-M HEAD
	Header        http.Header   //请求头信息
	EnableCookie  bool          //是否使用Cookie
	PostData      string        //POST values
	DialTimeout   time.Duration //创建连接超时 dial tcp: i/o timeout
	ConnTimeout   time.Duration //连接状态超时 WSARecv tcp: i/o timeout
	ReTryTimes    int           //尝试下载的最大次数
	RetryPause    time.Duration //下载失败后，下次尝试下载的等待时间
	RedirectTimes int           //重定向的最大次数，-1为不限制次数
	Priority      int           //指定调度优先级，默认为0（最小优先级为0）
	Reloadable    bool          //是否允许重复该链接下载

	proxy string //当用户界面设置可使用代理IP时，自动设置代理
}

const (
	DEFAULT_METHOD   = "GET"
	DEFAULT_PostData = ""
)
