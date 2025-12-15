package collect

import (
	"github.com/chenyukang1/crawler/pkg/log"
)

type LogCollector struct {
	base *BaseCollector
}

var Log = &LogCollector{
	base: &BaseCollector{
		DataCells: make(chan DataCell),
		ProcessBatch: func(batch []DataCell) {
			for _, cell := range batch {
				for k, v := range cell {
					log.Infof("collect %s : %v\n", k, v)
				}
			}
		},
		dataBatch: make([]DataCell, 0),
		batchSize: 1,
		finish:    make(chan struct{}, 1),
	},
}

func (l *LogCollector) Pipeline() {
	l.base.Pipeline()
}

func (l *LogCollector) Push(cell DataCell) {
	l.base.Push(cell)
}

func (l *LogCollector) Stop() {
	l.base.Stop()
}
