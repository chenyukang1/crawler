package process

import (
	"context"
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
			t := NewTaskQueue(context.Background())
			t.Init()
			t.Push(&CrawlTask{
				URL:      "high priority",
				Priority: 1,
			})
			t.Push(&CrawlTask{
				URL:      "mid priority",
				Priority: 2,
			})
			t.Push(&CrawlTask{
				URL:      "low priority",
				Priority: 3,
			})
			got, _ := t.Pop()
			if !reflect.DeepEqual(got.URL, tt.want) {
				t1.Errorf("Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}
