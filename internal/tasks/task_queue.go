package tasks

import (
	"container/heap"
	"sync"
	"time"
)

// GlobalQueue 全局唯一实例
var GlobalQueue = new(TaskQueueHeapWrapper)

type TaskQueue interface {
	Init()
	Push(CrawlTask)
	Pop() *CrawlTask
	Len() int
}

// TaskQueueHeapWrapper 基于最小堆的优先级队列，线程安全
type TaskQueueHeapWrapper struct {
	heap *TaskQueueHeap
	in   chan CrawlTask // 生产者输入
	out  chan CrawlTask // 分发给 worker
	mu   sync.RWMutex
}

func (t *TaskQueueHeapWrapper) Init() {
	heap.Init(t.heap)
	go t.watchQueue()
}

func (t *TaskQueueHeapWrapper) Push(task CrawlTask) {
	t.in <- task
}

func (t *TaskQueueHeapWrapper) Pop() *CrawlTask {
	select {
	case task := <-t.out:
		return &task
	case <-time.After(time.Second):
		return nil
	}
}

func (t *TaskQueueHeapWrapper) Len() int {
	t.mu.RLock()
	defer t.mu.Unlock()
	return t.heap.Len()
}

func (t *TaskQueueHeapWrapper) watchQueue() {
	for {
		select {
		case task := <-t.in:
			t.mu.Lock()
			heap.Push(t.heap, task)
			t.mu.Unlock()
		default:
			t.mu.Lock()
			if t.Len() > 0 {
				task := heap.Pop(t.heap).(CrawlTask)
				t.out <- task
				t.mu.Unlock()
			} else {
				t.mu.Unlock()
				time.Sleep(10 * time.Millisecond) // 防止空转
			}
		}
	}
}

type TaskQueueHeap []*CrawlTask

func (t *TaskQueueHeap) Len() int {
	return len(*t)
}

func (t *TaskQueueHeap) Less(i, j int) bool {
	return (*t)[i].Priority < (*t)[j].Priority
}

func (t *TaskQueueHeap) Swap(i, j int) {
	(*t)[i], (*t)[j] = (*t)[j], (*t)[i]
}

func (t *TaskQueueHeap) Push(x any) {
	task := x.(*CrawlTask)
	*t = append(*t, task)
}

func (t *TaskQueueHeap) Pop() any {
	old := (*t)[len(*t)-1]
	*t = (*t)[:len(*t)-1]
	return old
}
