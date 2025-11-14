package collect

type Collector interface {
	Pipeline()
}

type DataCell map[string]any

type BaseCollector struct {
	DataCells  chan DataCell
	dataBatch  []DataCell //分批输出结果缓存
	batchCount uint64     //分批数
}

func (c *BaseCollector) Pipeline() {
	finish := make(chan bool, 1)
	go func() {
		for cell := range c.DataCells {
			c.dataBatch = append(c.dataBatch, cell)
			if len(c.dataBatch) == 10 {
				break
			}
			c.batchCount++
		}
	}()
	<-finish
}
