package processor

import (
	"errors"
	"fmt"

	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/meta"
)

var (
	ErrNoMatch = errors.New("rules no match")
)

//Init init processor.
func Init() error {
	return nil
}

func Run(msg *meta.Message) (config.ActionConfig, error) {
	c, err := config.GetConfig()
	if err != nil {
		return config.ActionConfig{}, err
	}

	for _, p := range c.Processor {
		for i := range p.Topics {
			if msg.Topic == p.Topics[i] {
				for _, r := range p.Rules {
					if len(r.Match) == 0 {
						continue
					}
					ok, err := handler(msg, r.Match)
					if err != nil {
						return config.ActionConfig{}, err
					}
					if ok {
						return r.Action, nil
					}
				}
			}
		}
	}

	return config.ActionConfig{}, ErrNoMatch
}

func handler(msg *meta.Message, mc []config.MatchConfig) (bool, error) {
	var val string

	for _, m := range mc {
		vo, exist := msg.DataMap[m.Key]
		if exist {
			val = vo.(string)
		}
		msg.Trace(meta.StageProcessor, m.Method, fmt.Sprintf("begin key:%v, exist:%v value:%v, expect:%v", m.Key, exist, val, m.Value))
		if !match(m.Method, exist, val, m.Value) {
			msg.Trace(meta.StageProcessor, m.Method, fmt.Sprintf("end key:%v no match value:%v, expect:%v", m.Key, val, m.Value))
			return false, nil
		}
		msg.Trace(meta.StageProcessor, m.Method, fmt.Sprintf("end key:%v match value:%v expect:%v", m.Key, val, m.Value))
	}
	return true, nil
}
