package process

import (
	"reflect"
	"testing"
)

func TestTaskQueueHeapWrapper_Pop(t1 *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			"test pop",
			"high priority",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := NewTaskQueue()
			t.Init()
			t.Push(&CrawlTask{
				Url:      "high priority",
				Priority: 1,
			})
			t.Push(&CrawlTask{
				Url:      "mid priority",
				Priority: 2,
			})
			t.Push(&CrawlTask{
				Url:      "low priority",
				Priority: 3,
			})
			if got := t.Pop().Url; !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}
