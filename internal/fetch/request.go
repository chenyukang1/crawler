package fetch

import (
	"github.com/chenyukang1/crawler/internal/tasks"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	url           *url.URL      //目标URL，必须设置
	method        string        //GET POST POST-M HEAD
	header        http.Header   //请求头信息
	body          io.Reader     //POST values
	proxy         *url.URL      //代理IP时，自动设置代理
	dialTimeout   time.Duration //创建连接超时 dial tcp: i/o timeout
	connTimeout   time.Duration //连接状态超时 WSARecv tcp: i/o timeout
	redirectTimes int           //重定向的最大次数，-1为不限制次数
	enableCookie  bool          //是否使用Cookie
	reloadable    bool          //是否允许重复该链接下载
	retry         tasks.Retry   // 重试策略
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
}

func BuildRequest(task *tasks.CrawlTask) *Request {
	return nil
}
