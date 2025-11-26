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
	Rules       map[string]*Rule // 核心：规则注册表 (Flat Tree) Key 是规则名，Value 是规则对象
	EntryRule   string           // 入口规则名称
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
