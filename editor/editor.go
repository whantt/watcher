package editor

import (
	"fmt"
	"github.com/zssky/log"

	"github.com/dearcode/tracker/config"
	"github.com/dearcode/tracker/meta"
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

func Run(msg *meta.Message) {
	msg.Trace(meta.StageEditor, "begin", msg.Source)

	ec, err := config.GetConfig()
	if err != nil {
		msg.SetState(meta.StateError)
		msg.Trace(meta.StageEditor, "end", err.Error())
		return
	}

	for _, e := range ec.Editor {
		for i := range e.Topics {
			if msg.Topic == e.Topics[i] {
				m, ok := models[e.Model]
				if !ok {
					msg.Trace(meta.StageEditor, e.Model, "not found")
					continue
				}
                msg.Trace(meta.StageEditor, e.Model, fmt.Sprintf("begin data:%v", msg.DataMap))
				if err = m.Handler(msg, e.Data); err != nil {
					msg.SetState(meta.StateError)
					msg.Trace(meta.StageEditor, "end", err.Error())
					return
				}
                msg.Trace(meta.StageEditor, e.Model, fmt.Sprintf("end data:%v", msg.DataMap))
				break
			}
		}
	}

	msg.Trace(meta.StageEditor, "end", fmt.Sprintf("data:%v", msg.DataMap))
}