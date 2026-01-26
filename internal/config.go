package crawler

import (
	"os"

	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Spider struct {
	Selector string            `mapstructure:"selector"`
	Category string            `mapstructure:"category"`
	Fields   map[string]string `mapstructure:"fields"`
}

type Rule struct {
	Entry   string              `mapstructure:"entry"`
	Spiders map[string][]Spider `mapstructure:"spiders"`
}

type Config struct {
	Crawler struct {
		Parallelism int `mapstructure:"parallelism"`
		Worker      int `mapstructure:"worker"`
		IdleTime    int `mapstructure:"idleTime"`
	} `mapstructure:"crawler"`

	Rules map[string]Rule `mapstructure:"rules"`
}

func readConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Errorf("load .env fail %v", err)
		return nil, err
	}

	var conf Config
	v := viper.New()
	confPath := os.Getenv("CRAWLER_CONF_PATH")
	if confPath == "" {
		confPath = "."
	}
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(confPath)
	if err := v.ReadInConfig(); err != nil {
		log.Errorf("read in config fail %v", err)
		return nil, err
	}
	if err := v.Unmarshal(&conf); err != nil {
		log.Errorf("parse config fail %v", err)
		return nil, err
	}

	return &conf, nil
}
