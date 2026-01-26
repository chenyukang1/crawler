package collect

import (
	"github.com/chenyukang1/crawler/pkg/log"
)

type LogCollector struct {
	base *BaseCollector
}

func NewLogCollector(s int) Collector {
	return &BaseCollector{
		ProcessBatch: func(dc []DataCell) {
			for _, cell := range dc {
				for k, v := range cell {
					log.Infof("collect %s: %v\n", k, v)
				}
			}
		},
		batchSize: s,
	}
}

func (l *LogCollector) Pipeline(seq int) {
	l.base.Pipeline(seq)
}

func (l *LogCollector) Push(cell DataCell) {
	l.base.Push(cell)
}

func (l *LogCollector) Finish() {
	l.base.Finish()
}
