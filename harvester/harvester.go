package harvester

import (
	"errors"

	"github.com/zssky/log"

	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/meta"
)

var (
	models           = map[string]Harvester{}
	ErrModelNotFound = errors.New("kafka model not found")
)

type Harvester interface {
	Init(c config.HarvesterConfig) error
	Start() <-chan *meta.Message
	Stop()
}

//Init init harvester.
func Init() error {
	for k := range models {
		log.Debugf("harvester model:%v", k)
	}

	ec, err := config.GetConfig()
	if err != nil {
		return err
	}

	kh, ok := models["kafka"]
	if !ok {
		return ErrModelNotFound
	}

	return kh.Init(ec.Harvester)
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

func Reader() <-chan *meta.Message {
	kh, _ := models["kafka"]

	return kh.Start()
}

func Stop() {
	kh, _ := models["kafka"]
	kh.Stop()
}
