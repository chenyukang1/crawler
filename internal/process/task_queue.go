// Package process process main loop
package process

import (
	"container/heap"
	"context"
	"time"
)

type TaskQueue interface {
	Init()
	Push(*CrawlTask)
	Pop() (*CrawlTask, bool)
}

// TaskQueueHeapWrapper 基于最小堆的优先级队列，线程安全
type TaskQueueHeapWrapper struct {
	in   chan *CrawlTask // 生产者输入
	out  chan *CrawlTask // 分发给 worker
	heap *TaskQueueHeap
	ctx  context.Context
}

func NewTaskQueue(c context.Context) TaskQueue {
	h := make(TaskQueueHeap, 0)
	return &TaskQueueHeapWrapper{
		in:   make(chan *CrawlTask),
		out:  make(chan *CrawlTask),
		heap: &h,
		ctx:  c,
	}
}

func (t *TaskQueueHeapWrapper) Init() {
	heap.Init(t.heap)
	go t.watchQueue()
}

func (t *TaskQueueHeapWrapper) Push(task *CrawlTask) {
	t.in <- task
}

func (t *TaskQueueHeapWrapper) Pop() (*CrawlTask, bool) {
	select {
	case <-t.ctx.Done():
		return nil, false

	case task, ok := <-t.out:
		return task, ok

	case <-time.After(time.Second):
		return nil, true
	}
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
		}

		select {
		// 尝试清空通道
		case <-t.ctx.Done():
			for {
				select {
				case _, ok := <-t.in:
					if !ok {
						return
					}
					// do nothing
				default:
					return
				}
			}

		case task, ok := <-t.in:
			if !ok {
				t.in = nil
				continue
			}
			heap.Push(t.heap, task)

		case out <- nextTask:
			heap.Pop(t.heap)
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
