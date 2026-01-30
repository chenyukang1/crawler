package spider

import (
	"errors"
)

type Rule struct {
	Name string
	Run  RuleFunc
}

// Spider 采集规则引擎
type Spider struct {
	URL string
	Method string
	Name        string
	Rules       map[string]*Rule // 核心：规则表 (Flat Tree) Key 是规则名，Value 是规则对象
	EntryRule   string           // 入口规则名称
}

type Registry map[string]*Spider

var GlobalRegistry = make(Registry)

func (r Registry) Register(s *Spider) (err error) {
	if s == nil {
		err = errors.New("spider cannnot be nil")
		return
	}
	if s.Name == "" {
		err = errors.New("spider name cannnot be empty")
		return
	}
	if _, ok := r[s.Name]; ok {
		err = errors.New("duplicate spider name")
		return
	}
	r[s.Name] = s
	return nil
}

func (r Registry) GetSpider(name string) (spider *Spider, err error) {
	if name == "" {
		return nil, errors.New("spider name cannnot be empty")
	}
	return r[name], nil
}
