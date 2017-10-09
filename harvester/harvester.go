package harvester

import (
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/meta"
)

var (
	models           = map[string]Harvester{}
	ErrModelNotFound = errors.New("kafka model not found")
)

type Harvester interface {
	Init(c config.HarvesterConfig, msgChan chan<- *meta.Message) error
	Stop()
}

//Init init harvester.
func Init(msgChan chan<- *meta.Message) error {
	for k := range models {
		log.Debugf("harvester model:%v", k)
	}

	ec, err := config.GetConfig()
	if err != nil {
		return errors.Trace(err)
	}

	kh, ok := models["kafka"]
	if !ok {
		return errors.Trace(ErrModelNotFound)
	}

	return kh.Init(ec.Harvester, msgChan)
}

//Register 添加模块
func Register(name string, m Harvester) {
	if _, ok := models[name]; ok {
		log.Errorf("harvester model:%v exist", name)
		return
	}

	models[name] = m
	log.Debugf("new mode:%v", name)
}

func Stop() {
	kh, _ := models["kafka"]
	kh.Stop()
}
