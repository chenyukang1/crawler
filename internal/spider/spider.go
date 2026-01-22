package spider

import (
	"errors"
)

type Rule struct {
	Name string
	Run  RuleFunc
}

// Spider 解析规则引擎
type Spider struct {
	Name        string
	Description string
	Rules       map[string]*Rule // 核心：规则表 (Flat Tree) Key 是规则名，Value 是规则对象
	EntryRule   string           // 入口规则名称
}

type Registry map[string]*Spider

var GlobalRegistry = make(Registry)

func (r Registry) Register(name string, spider *Spider) (err error) {
	if name == "" {
		err = errors.New("spider name cannnot be empty")
		return
	}
	if spider == nil {
		err = errors.New("spider cannnot be nil")
		return
	}
	if _, ok := r[name]; ok {
		err = errors.New("duplicate spider name")
		return
	}
	r[name] = spider
	return nil
}

func (r Registry) GetSpider(name string) (spider *Spider, err error) {
	if name == "" {
		return nil, errors.New("spider name cannnot be empty")
	}
	return r[name], nil
}
