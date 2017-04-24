package common

import (
	"fmt"
	"github.com/dearcode/tracker/config"
	"github.com/dearcode/tracker/meta"
	"github.com/dearcode/tracker/processor"
)

var (
	cp = commonProcessor{}
)

type commonProcessor struct {
}

func init() {
	processor.Register("common", &cp)
}

func (cp *commonProcessor) Handler(msg *meta.Message, mc []config.MatchConfig) (bool, error) {
	if len(mc) == 0 {
		return false, nil
	}
	for _, m := range mc {
		vo, ok := msg.DataMap[m.Key]
		if !ok {
			msg.Trace(meta.StageProcessor, "common", fmt.Sprintf("%v not found key:%v", m.Method, m.Key))
			continue
		}
		val := vo.(string)
		msg.Trace(meta.StageProcessor, "common", fmt.Sprintf("%v [%v:%v]", m.Method, val, m.Value))
		if !cp.match(m.Method, val, m.Value) {
			msg.Trace(meta.StageProcessor, "common", fmt.Sprintf("%v no match:[%v:%v]", m.Method, val, m.Value))
			return false, nil
		}
		msg.Trace(meta.StageProcessor, "common", fmt.Sprintf("%v match %v:%v", m.Method, m.Key, val))
	}
	return true, nil
}

func (cp *commonProcessor) match(method, key, val string) bool {
	switch method {
	case "equal":
		return equal(key, val)
	case "lesser":
		return lesser(key, val)
	case "larger":
		return larger(key, val)
	case "contains":
		return contains(key, val)
	}

	return false
}
