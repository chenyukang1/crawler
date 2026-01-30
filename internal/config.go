package crawler

import (
	"github.com/chenyukang1/crawler/pkg/log"
	"github.com/spf13/viper"
)

type Rule struct {
	Selector string            `mapstructure:"selector"`
	Category string            `mapstructure:"category"`
	Fields   map[string]string `mapstructure:"fields"`
}

type Spdier struct {
	URL    string            `mapstructure:"url"`
	Method string            `mapstructure:"method"`
	Entry  string            `mapstructure:"entry"`
	Rules  map[string][]Rule `mapstructure:"rules"`
}

type Retry struct {
	Times    int `mapstructure:"times"`
	Interval int `mapstructure:"interval"`
}

type Task struct {
	URL           string            `mapstructure:"url"`
	Method        string            `mapstructure:"method"`
	Headers       map[string]string `mapstructure:"headers"`
	Spider        string            `mapstructure:"spider"`
	Rule          string            `mapstructure:"rule"`
	Priority      int               `mapstructure:"priority"`
	EnableCookie  bool              `mapstructure:"enableCookie"`
	RedirectTimes int               `mapstructure:"redirectTimes"`
	DialTimeout   int               `mapstructure:"dialTimeout"`
	ConnTimeout   int               `mapstructure:"connTimeout"`
	Retry         Retry             `mapstructure:"retry"`
}

type Config struct {
	Crawler struct {
		Parallelism int `mapstructure:"parallelism"`
		Worker      int `mapstructure:"worker"`
		IdleTime    int `mapstructure:"idleTime"`
	} `mapstructure:"crawler"`

	Spiders map[string]Spdier `mapstructure:"spiders"`
	Tasks   []Task            `mapstructure:"tasks"`
}

var (
	Conf    *Config
	Cfgfile string
)

func ReadConfig() {
	v := viper.New()
	if Cfgfile != "" {
		v.SetConfigFile(Cfgfile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath("../config")
		v.AddConfigPath("../../config")
	}
	if err := v.ReadInConfig(); err != nil {
		log.Errorf("read in config fail %v", err)
		panic(err)
	}
	if err := v.Unmarshal(&Conf); err != nil {
		log.Errorf("parse config fail %v", err)
		panic(err)
	}
}
