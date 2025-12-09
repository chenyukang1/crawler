package spider

import (
	"errors"
	"sync/atomic"
)

type Rule struct {
	Name string
	Run  RuleFunc
}

// Spider 解析规则引擎
type Spider struct {
	Id          int64
	Name        string
	Description string
	Rules       map[string]*Rule // 核心：规则表 (Flat Tree) Key 是规则名，Value 是规则对象
	EntryRule   string           // 入口规则名称
}

// Registry 注册表
type Registry struct {
	m       map[int64]*Spider
	counter int64
}

var GlobalRegistry = &Registry{
	m: make(map[int64]*Spider),
}

func (s *Spider) Register(ruleName string, ruleFunc RuleFunc) (err error) {
	if ruleName == "" {
		err = errors.New("ruleName cannot be empty")
	}
	if ruleFunc == nil {
		err = errors.New("ruleFunc cannot be nil")
	}
	if s.Rules == nil {
		s.Rules = make(map[string]*Rule)
	}
	s.Rules[ruleName] = &Rule{
		Name: ruleName,
		Run:  ruleFunc,
	}
	return
}

func (r *Registry) Register(spider *Spider) {
	id := atomic.AddInt64(&r.counter, 1)
	spider.Id = id
	r.m[id] = spider
}
