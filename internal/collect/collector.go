package collect

type Collector interface {
	Pipeline()
	Push(cell DataCell)
	Stop()
}

type DataCell map[string]any

type BaseCollector struct {
	DataCells    chan DataCell
	ProcessBatch func([]DataCell)

	dataBatch []DataCell //分批输出结果缓存
	batchSize int        //分批大小
	count     int        //分批数
	finish    chan bool  //停止channel
}

func (c *BaseCollector) Pipeline() {
	go func() {
		for cell := range c.DataCells {
			c.dataBatch = append(c.dataBatch, cell)
			if len(c.dataBatch) == c.batchSize {
				c.ProcessBatch(c.dataBatch)
				c.count++
			}
		}
		c.ProcessBatch(c.dataBatch)
		c.count++
		c.finish <- true
	}()
	<-c.finish
}

func (c *BaseCollector) Push(cell DataCell) {
	c.DataCells <- cell
}

func (c *BaseCollector) Stop() {
	close(c.finish)
}
