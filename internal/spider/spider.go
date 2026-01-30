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
	Name      string           // 规则名称
	Rules     map[string]*Rule // 核心：规则表 (Flat Tree) Key 是规则名，Value 是规则对象
	EntryRule string           // 入口规则名称
}

type Registry map[string]*Spider

var GlobalRegistry = make(Registry)

func (r Registry) Register(s *Spider) error {
	if s == nil {
		return errors.New("spider cannnot be nil")
	}
	if s.Name == "" {
		return errors.New("spider name cannnot be empty")
	}
	if s.EntryRule == "" {
		return errors.New("spider entry rule cannot be empty")
	}
	if _, ok := r[s.Name]; ok {
		return errors.New("duplicate spider name")
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
