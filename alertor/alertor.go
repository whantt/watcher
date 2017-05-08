package alertor

import (
	"github.com/zssky/log"

	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/meta"
)

var (
	models = map[string]Alertor{}
)

//Alertor 报警接口.
type Alertor interface {
	Handler(msg *meta.Message, ac config.ActionConfig) error
}

//Init init alertor.
func Init() error {
	for k := range models {
		log.Debugf("alertor model:%v", k)
	}

	return nil
}

//Register 添加模块
func Register(name string, a Alertor) {
	if _, ok := models[name]; ok {
		log.Errorf("alertor model:%v exist", name)
		return
	}

	models[name] = a
	log.Debugf("new alertor mode:%v", name)
}

//Run 报警模块入口.
func Run(msg *meta.Message, ac config.ActionConfig) error {
	log.Debugf("msg:%#v, action:%v", msg.DataMap, ac)
	if ac.Mail {
		if m, ok := models["mail"]; ok {
			if err := m.Handler(msg, ac); err != nil {
				msg.Trace(meta.StageAlertor, "mail", err.Error())
				return err
			}
		}
	}

	if ac.Message {
		if m, ok := models["message"]; ok {
			if err := m.Handler(msg, ac); err != nil {
				msg.Trace(meta.StageAlertor, "message", err.Error())
				return err
			}
		}
	}

	return nil
}
