package fetch

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"errors"
	"github.com/chenyukang1/crawler/internal/logger"
	"io"
	"log"
	"math/rand"
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

func (f *Fetcher) Fetch(ctx context.Context, req *Request) (resp *http.Response, err error) {
	client, err := f.buildHttpClient(req)
	if err != nil {
		logger.Errorf("create request for %s fail: %v", req.url, err)
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, req.method, req.url.String(), req.body)
	request.Header = req.header

	res, err := req.retry.DoRetry(ctx, func() (any, error) {
		return client.Do(request)
	})
	if err != nil {
		logger.Errorf("get from url %s fail: %v", req.url, err)
		return nil, err
	}
	resp = res.(*http.Response)
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

func doFetch(timeout time.Duration, url string) (string, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("create request for %s fail: %v", url, err)
		return "", err
	}

	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("get from url %s fail: %v", url, err)
		return "", err
	}
	defer resp.Body.Close()

	body := resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Printf("gzip decode fail from %s: %v", url, err)
			return "", err
		}
		defer gz.Close()
		body = gz
	}

	bytes, err := io.ReadAll(body)
	if err != nil {
		log.Printf("read body from url %s fail: %v", url, err)
		return "", err
	}

	s := string(bytes)
	return s, nil
}
