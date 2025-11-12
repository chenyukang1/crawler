package scheduler

import "container/heap"

var queue = new(TaskQueueHeapWrapper)

type TaskQueue interface {
	Init()
	Push(CrawlTask)
	Pop() CrawlTask
	Len() int
}

type TaskQueueHeapWrapper struct {
	heap *TaskQueueHeap
}

func (t TaskQueueHeapWrapper) Init() {
	heap.Init(t.heap)
}

func (t TaskQueueHeapWrapper) Push(task CrawlTask) {
	heap.Push(t.heap, task)
}

func (t TaskQueueHeapWrapper) Pop() CrawlTask {
	return heap.Pop(t.heap).(CrawlTask)
}

func (t TaskQueueHeapWrapper) Len() int {
	return t.heap.Len()
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
