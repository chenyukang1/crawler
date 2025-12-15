package collect

import "github.com/chenyukang1/crawler/pkg/log"

type Collector interface {
	Pipeline()
	Push(cell DataCell)
	Stop()
}

type DataCell map[string]any

type BaseCollector struct {
	DataCells    chan DataCell
	ProcessBatch func([]DataCell)

	dataBatch []DataCell    //分批输出结果缓存
	batchSize int           //分批大小
	count     int           //分批数
	finish    chan struct{} //停止channel
}

func NewDataCell() DataCell {
	return make(DataCell)
}

func (c DataCell) Set(key string, value any) {
	c[key] = value
}

func (c *BaseCollector) Pipeline() {
	go func() {
		for cell := range c.DataCells {
			c.dataBatch = append(c.dataBatch, cell)
			if len(c.dataBatch) == c.batchSize {
				c.ProcessBatch(c.dataBatch)
				c.dataBatch = c.dataBatch[c.batchSize:]
				c.count++
			}
		}
		c.ProcessBatch(c.dataBatch)
		c.count++
		c.finish <- struct{}{}
		log.Info("数据收集完成!")
	}()
	<-c.finish
}

func (c *BaseCollector) Push(cell DataCell) {
	c.DataCells <- cell
}

func (c *BaseCollector) Stop() {
	close(c.finish)
}
