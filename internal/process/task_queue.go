package process

import (
	"container/heap"
	"time"
)

type TaskQueue interface {
	Init()
	Chan() chan *CrawlTask
	Push(*CrawlTask)
	Pop() *CrawlTask
}

// TaskQueueHeapWrapper 基于最小堆的优先级队列，线程安全
type TaskQueueHeapWrapper struct {
	in   chan *CrawlTask // 生产者输入
	out  chan *CrawlTask // 分发给 worker
	heap *TaskQueueHeap
	stop chan struct{} // 停止命令
}

func NewTaskQueue() TaskQueue {
	taskHeap := make(TaskQueueHeap, 0)
	return &TaskQueueHeapWrapper{
		in:   make(chan *CrawlTask),
		out:  make(chan *CrawlTask),
		heap: &taskHeap,
		stop: make(chan struct{}),
	}
}

func (t *TaskQueueHeapWrapper) Init() {
	heap.Init(t.heap)
	go t.watchQueue()
}

func (t *TaskQueueHeapWrapper) Chan() chan *CrawlTask {
	return t.out
}

func (t *TaskQueueHeapWrapper) Push(task *CrawlTask) {
	t.in <- task
}

func (t *TaskQueueHeapWrapper) Pop() *CrawlTask {
	select {
	case task := <-t.out:
		return task
	case <-time.After(time.Second):
		return nil
	}
}

func (t *TaskQueueHeapWrapper) Stop() {
	t.stop <- struct{}{}
}

func (t *TaskQueueHeapWrapper) watchQueue() {
	for {
		var (
			out      chan *CrawlTask
			nextTask *CrawlTask
		)
		if t.heap.Len() > 0 {
			out = t.out
			nextTask = t.heap.First()
		} else {
			out = nil // 禁用发送分支
		}

		select {
		case task := <-t.in:
			heap.Push(t.heap, task)
		case out <- nextTask:
			heap.Pop(t.heap)
		case <-t.stop:
			return
		}
	}
}

type TaskQueueHeap []*CrawlTask

func (t TaskQueueHeap) First() *CrawlTask {
	return t[0]
}

func (t TaskQueueHeap) Len() int {
	return len(t)
}

func (t TaskQueueHeap) Less(i, j int) bool {
	return t[i].Priority < t[j].Priority
}

func (t TaskQueueHeap) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t *TaskQueueHeap) Push(x any) {
	task := x.(*CrawlTask)
	*t = append(*t, task)
}

func (t *TaskQueueHeap) Pop() any {
	old := (*t)[len(*t)-1]
	(*t)[len(*t)-1] = nil
	*t = (*t)[:len(*t)-1]
	return old
}
