package collect

import "github.com/chenyukang1/crawler/pkg/log"

type Collector interface {
	Pipeline(seq int)
	Push(cell DataCell)
	Finish()
}

type DataCell map[string]any

type BaseCollector struct {
	DataCells    chan DataCell
	ProcessBatch func([]DataCell)

	dataBatch []DataCell // 分批输出结果缓存
	batchSize int        // 分批大小
	count     int        // 收集数量
}

func NewDataCell() DataCell {
	return make(DataCell)
}

func (c DataCell) Set(key string, value any) {
	c[key] = value
}

func (c *BaseCollector) Pipeline(seq int) {
	c.DataCells = make(chan DataCell)
	c.dataBatch = make([]DataCell, 0)
	for cell := range c.DataCells {
		c.dataBatch = append(c.dataBatch, cell)
		if len(c.dataBatch) == c.batchSize {
			c.ProcessBatch(c.dataBatch)
			c.dataBatch = c.dataBatch[:0]
			c.count += c.batchSize
		}
	}
	c.ProcessBatch(c.dataBatch)
	c.count += len(c.dataBatch)
	log.Infof("[crawler-%d] collect finish", seq)
}

func (c *BaseCollector) Push(cell DataCell) {
	c.DataCells <- cell
}

func (c *BaseCollector) Finish() {
	close(c.DataCells)
}
