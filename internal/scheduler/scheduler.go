package scheduler

import (
	"github.com/chenyukang1/crawler/internal/process"
	"github.com/chenyukang1/crawler/pkg/model"
	"time"
)

func Start() {
	taskList := make(chan model.FetchTask, 20)
	respList := make(chan model.FetchResp, 100)

	c := &process.Config{Concurrency: 20, MaxRetries: 3, RequestTimeout: time.Second * 10}
	f := &process.Fetcher{Config: c, TaskList: taskList}
	p := &process.Parser{Config: c, Filter: process.NewBloomFilter(1024), TaskList: taskList, RespList: respList}
	go f.Fetch(respList)
	go p.Parse()

	tasks := []model.FetchTask{
		{"https://m.douban.com/movie/"},
		{"https://www.baidu.com"},
		{"https://golang.org"},
	}
	for _, task := range tasks {
		taskList <- task
	}

	time.Sleep(time.Second * 30)
}
