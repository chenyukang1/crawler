package fetch

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"errors"
	"github.com/chenyukang1/crawler/internal/logger"
	"github.com/chenyukang1/crawler/internal/spider"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

type Fetcher struct {
	CookieJar *cookiejar.Jar
}

var (
	jar, _  = cookiejar.New(nil)
	Default = &Fetcher{
		CookieJar: jar,
	}
)

func (f *Fetcher) Fetch(ctx context.Context, req *Request, c *spider.Context) (err error) {
	client, err := f.buildHttpClient(req)
	if err != nil {
		logger.Errorf("create request for %s fail: %v", req.url, err)
		return
	}

	request, err := http.NewRequestWithContext(ctx, req.method, req.url.String(), req.body)
	request.Header = req.header

	res, err := req.retry.DoRetry(ctx, func() (any, error) {
		return client.Do(request)
	})
	if err != nil {
		logger.Errorf("get from url %s fail: %v", req.url, err)
		return
	}
	resp := res.(*http.Response)
	c.Request = request
	c.Response = resp

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var gzipReader *gzip.Reader
		gzipReader, err = gzip.NewReader(resp.Body)
		if err != nil {
			logger.Errorf("gzip read body from url %s fail: %v", req.url, err)
			return
		}
		resp.Body = gzipReader

	case "deflate":
		resp.Body = flate.NewReader(resp.Body)

	case "zlib":
		var readCloser io.ReadCloser
		readCloser, err = zlib.NewReader(resp.Body)
		if err != nil {
			logger.Errorf("zlib read body from url %s fail: %v", req.url, err)
			return
		}
		resp.Body = readCloser
	}
	return
}

func (f *Fetcher) buildHttpClient(req *Request) (*http.Client, error) {
	client := &http.Client{}

	if req.enableCookie {
		client.Jar = f.CookieJar
	}

	client.CheckRedirect = func(q *http.Request, via []*http.Request) error {
		if req.redirectTimes < 0 {
			return nil
		}
		if req.redirectTimes == 0 {
			return errors.New("no allow redirects")
		}
		if len(via) >= req.redirectTimes {
			return errors.New("stop after redirects " + strconv.Itoa(req.redirectTimes))
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
				if err == nil {
					dnsCache.Store(addr, conn.RemoteAddr().String())
				}
			}
			conn, err = net.DialTimeout(network, ipPort, req.dialTimeout)
			if err != nil {
				return nil, err
			}
			if req.connTimeout > 0 {
				err = conn.SetDeadline(time.Now().Add(req.connTimeout))
				if err != nil {
					return nil, err
				}
			}
			return conn, nil
		},
	}
	if req.proxy != nil {
		transport.Proxy = http.ProxyURL(req.proxy)
	}
	if strings.ToLower(req.url.Scheme) == "https" {
		transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
		transport.DisableCompression = true
	}
	client.Transport = transport

	return client, nil
}
