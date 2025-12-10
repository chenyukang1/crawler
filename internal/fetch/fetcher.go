package fetch

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"errors"
	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/chenyukang1/crawler/pkg/retry"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Fetcher struct {
	CookieJar *cookiejar.Jar
}

type Request struct {
	Url           *url.URL      //目标URL，必须设置
	Method        string        //GET POST POST-M HEAD
	Header        http.Header   //请求头信息
	Body          io.Reader     //POST values
	Proxy         *url.URL      //代理IP时，自动设置代理
	DialTimeout   time.Duration //创建连接超时 dial tcp: i/o timeout
	ConnTimeout   time.Duration //连接状态超时 WSARecv tcp: i/o timeout
	RedirectTimes int           //重定向的最大次数，-1为不限制次数
	EnableCookie  bool          //是否使用Cookie
	Reloadable    bool          //是否允许重复该链接下载
	Retry         retry.Retry   // 重试策略
}

var (
	jar, _  = cookiejar.New(nil)
	Default = &Fetcher{
		CookieJar: jar,
	}
)

func (f *Fetcher) Fetch(ctx context.Context, req *Request) (request *http.Request, response *http.Response, err error) {
	var client *http.Client
	client, err = f.buildHttpClient(req)
	if err != nil {
		log.Errorf("create http client for %s fail: %v", req.Url, err)
		return
	}

	request, err = http.NewRequestWithContext(ctx, req.Method, req.Url.String(), req.Body)
	if err != nil {
		log.Errorf("create request for %s fail: %v", req.Url, err)
	}
	request.Header = req.Header

	res, err := req.Retry.DoRetry(ctx, func() (any, error) {
		return client.Do(request)
	})
	if err != nil {
		log.Errorf("get from Url %s fail: %v", req.Url, err)
		return
	}
	response = res.(*http.Response)

	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		var gzipReader *gzip.Reader
		gzipReader, err = gzip.NewReader(response.Body)
		if err != nil {
			log.Errorf("gzip read body from Url %s fail: %v", req.Url, err)
			return
		}
		response.Body = gzipReader

	case "deflate":
		response.Body = flate.NewReader(response.Body)

	case "zlib":
		var readCloser io.ReadCloser
		readCloser, err = zlib.NewReader(response.Body)
		if err != nil {
			log.Errorf("zlib read body from Url %s fail: %v", req.Url, err)
			return
		}
		response.Body = readCloser
	}
	return
}

func (f *Fetcher) buildHttpClient(req *Request) (*http.Client, error) {
	client := &http.Client{}

	if req.EnableCookie {
		client.Jar = f.CookieJar
	}

	client.CheckRedirect = func(q *http.Request, via []*http.Request) error {
		if req.RedirectTimes < 0 {
			return nil
		}
		if req.RedirectTimes == 0 {
			return errors.New("no allow redirects")
		}
		if len(via) >= req.RedirectTimes {
			return errors.New("stop after redirects " + strconv.Itoa(req.RedirectTimes))
		}
		return nil
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var (
				conn net.Conn
				err  error
			)
			ipPort, ok := dnsCache.Load(addr)
			if ok {
				defer func() {
					if err != nil {
						dnsCache.Delete(addr)
					}
				}()
			} else {
				ipPort = addr
				defer func() {
					if err == nil {
						dnsCache.Store(addr, conn.RemoteAddr().String())
					}
				}()
			}
			conn, err = net.DialTimeout(network, ipPort, req.DialTimeout)
			if err != nil {
				return nil, err
			}
			if req.ConnTimeout > 0 {
				err = conn.SetDeadline(time.Now().Add(req.ConnTimeout))
				if err != nil {
					return nil, err
				}
			}
			return conn, nil
		},
	}
	if req.Proxy != nil {
		transport.Proxy = http.ProxyURL(req.Proxy)
	}
	if strings.ToLower(req.Url.Scheme) == "https" {
		transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
		transport.DisableCompression = true
	}
	client.Transport = transport

	return client, nil
}
