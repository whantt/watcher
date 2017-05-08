package editor

import (
	"fmt"
	"github.com/zssky/log"

	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/meta"
)

var (
	models = map[string]Editor{}
)

type Editor interface {
	Handler(msg *meta.Message, argv map[string]interface{}) error
}

//Init init editor.
func Init() error {
	for k := range models {
		log.Debugf("editor model:%v", k)
	}

	return nil
}

//Register 添加模块
func Register(name string, m Editor) {
	if _, ok := models[name]; ok {
		log.Errorf("editor model:%v exist", name)
		return
	}

	models[name] = m
	log.Debugf("new mode:%v", name)
}

func Run(msg *meta.Message) error {
	ec, err := config.GetConfig()
	if err != nil {
		return err
	}

	for ei, e := range ec.Editor {
		for i := range e.Topics {
			if msg.Topic == e.Topics[i] {
				m, ok := models[e.Model]
				if !ok {
					msg.Trace(meta.StageEditor, e.Model, "not found")
					continue
				}
				msg.Trace(meta.StageEditor, e.Model, fmt.Sprintf("begin data:%v", msg.DataMap))
				if err = m.Handler(msg, ec.Editor[ei].Data); err != nil {
					return err
				}
				msg.Trace(meta.StageEditor, e.Model, fmt.Sprintf("end data:%#v", msg.DataMap))
				break
			}
		}
	}

	return nil
}
