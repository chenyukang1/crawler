package crawler

import (
	"time"

	"github.com/spf13/viper"

	"github.com/chenyukang1/crawler/pkg/log"
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
	Parallelism int           `mapstructure:"parallelism"`
	Worker      int           `mapstructure:"worker"`
	MaxIdleTime time.Duration `mapstructure:"maxIdleTime"`

	Spiders map[string]Spdier `mapstructure:"spiders"`
	Tasks   []Task            `mapstructure:"tasks"`
}

var (
	Viper   *viper.Viper
	Conf    *Config
	Cfgfile string
)

func InitViper() {
	Viper = viper.New()
	Viper.SetDefault("parallelism", 10)
	Viper.SetDefault("worker", 10)
	Viper.SetDefault("maxIdleTime", 10*time.Second)
}

func ReadConfig() {
	if Cfgfile != "" {
		Viper.SetConfigFile(Cfgfile)
	} else {
		Viper.SetConfigName("config")
		Viper.SetConfigType("yaml")
		Viper.AddConfigPath("./config")
		Viper.AddConfigPath("../config")
		Viper.AddConfigPath("../../config")
	}
	if err := Viper.ReadInConfig(); err != nil {
		log.Errorf("read in config fail %v", err)
		panic(err)
	}
	if err := Viper.Unmarshal(&Conf); err != nil {
		log.Errorf("parse config fail %v", err)
		panic(err)
	}
}
