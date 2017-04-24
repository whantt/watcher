package processor

import (
	"errors"
	"fmt"
	"github.com/zssky/log"

	"github.com/dearcode/tracker/config"
	"github.com/dearcode/tracker/meta"
)

var (
	models = map[string]Processor{}

	ErrNoMatch = errors.New("rules no match")
)

type Processor interface {
	Handler(msg *meta.Message, mc []config.MatchConfig) (bool, error)
}

//Init init processor.
func Init() error {
	for k := range models {
		log.Debugf("processor model:%v", k)
	}

	return nil
}

//Register 添加模块
func Register(name string, m Processor) {
	if _, ok := models[name]; ok {
		log.Errorf("processor model:%v exist", name)
		return
	}

	models[name] = m
	log.Debugf("new processor mode:%v", name)
}

func Run(msg *meta.Message) (config.ActionConfig, error) {
	c, err := config.GetConfig()
	if err != nil {
		return config.ActionConfig{}, err
	}

	for pi, p := range c.Processor {
		for i := range p.Topics {
			if msg.Topic == p.Topics[i] {
				log.Debugf("topic:%v rules:%v", msg.Topic, p.Rules)
				for _, r := range p.Rules {
					m, ok := models[r.Model]
					if !ok {
						msg.Trace(meta.StageProcessor, r.Model, "not found")
						continue
					}

					msg.Trace(meta.StageProcessor, r.Model, fmt.Sprintf("begin data:%v, rules:%v", msg.DataMap, r.Match))
					ok, err := m.Handler(msg, r.Match)
					if err != nil {
						msg.Trace(meta.StageProcessor, r.Model, err.Error())
						return config.ActionConfig{}, err
					}
					if ok {
						log.Debugf("rule:%v, model:%v action:%#v", r.Match, r.Model, r.Action)
						msg.Trace(meta.StageProcessor, r.Model, fmt.Sprintf("match rules:%v", c.Processor[pi].Rules))
						return r.Action, nil
					}
					msg.Trace(meta.StageProcessor, r.Model, fmt.Sprintf("end no match data:%v", msg.DataMap))
				}
			}
		}
	}

	return config.ActionConfig{}, ErrNoMatch
}
