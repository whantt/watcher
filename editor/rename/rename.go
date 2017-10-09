package rename

import (
	"encoding/json"
	"github.com/zssky/log"

	"github.com/dearcode/watcher/editor"
	"github.com/dearcode/watcher/meta"
)

var (
	re = renameEditor{}
)

type renameEditor struct {
}

func init() {
	editor.Register("rename", &re)
}

type renameField struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type renameConfig struct {
	Fields []renameField `json:"fields"`
}

func parseArgs(argv map[string]interface{}) renameConfig {
	rc := renameConfig{}
	log.Debugf("argv:%v", argv)

	buf, _ := json.Marshal(argv)

	if err := json.Unmarshal(buf, &rc); err != nil {
		log.Errorf("Unmarshal rename data error:%v, buf:%s", err.Error(), buf)
		return rc
	}
	log.Debugf("renameConfig:%+v", rc)

	return rc
}

func (e *renameEditor) Handler(msg *meta.Message, argv map[string]interface{}) error {
	c := parseArgs(argv)

	for _, f := range c.Fields {
		if v, ok := msg.DataMap[f.From]; ok {
			delete(msg.DataMap, f.From)
			msg.DataMap[f.To] = v
			log.Debugf("rename key from:%v, to:%v", f.From, f.To)
		}
	}
	return nil
}
