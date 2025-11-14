package fetch

import (
	"compress/gzip"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Fetcher struct {
	jar *cookiejar.Jar
}

type Request struct {
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
}

var httpClient *http.Client

func (f *Fetcher) Init() {
	f.jar, _ = cookiejar.New(nil)
	httpClient = &http.Client{
		Jar: f.jar,
		Transport: &http.Transport{
			MaxIdleConns:    100,
			IdleConnTimeout: 90 * time.Second,
		},
	}
}

func (f *Fetcher) Fetch(respList chan model.FetchResp) {
	for task := range f.TaskList {
		go func(Task model.FetchTask) {
			for attempt := 1; attempt < f.Config.MaxRetries; attempt++ {
				resp, err := doFetch(f.Config.RequestTimeout, task.Url)
				if err == nil {
					respList <- model.FetchResp{Content: resp}
					break
				}
				time.Sleep(time.Second * time.Duration(attempt))
			}
		}(task)
	}
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
