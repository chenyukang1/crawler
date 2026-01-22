package global

import (
	"sync"

	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/spf13/viper"
)

type Config struct {
	Crawler struct {
		Parallelism int
	}
}

var (
	globalConf *Config
	once       sync.Once
)

func Init() *Config {
	once.Do(func() {
		v := viper.New()
		v.SetConfigFile("config.yaml")
		if err := v.ReadInConfig(); err != nil {
			log.Errorf("read in config fail %v", err)
			panic(err)
		}
		if err := v.Unmarshal(&globalConf); err != nil {
			log.Errorf("parse config fail %v", err)
			panic(err)
		}
	})
	return globalConf
}
