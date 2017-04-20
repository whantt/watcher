package json

import (
	"encoding/json"
	"fmt"

	"github.com/zssky/log"

	"github.com/dearcode/tracker/editor"
	"github.com/dearcode/tracker/meta"
)

var (
	e = jsonEditor{}
)

type jsonEditor struct {
}

func init() {
	editor.Register("json", &e)
}

type jsonConfig struct {
	Fields []string
}

func parseArgs(argv map[string]interface{}) jsonConfig {
	jc := jsonConfig{}

	if fo, ok := argv["fields"]; ok {
		for _, fo := range fo.([]interface{}) {
			jc.Fields = append(jc.Fields, fo.(string))
		}
	}

	return jc
}

func (e *jsonEditor) jsonDecode(msg *meta.Message, data []byte) error {
	log.Debugf("Unmarshal data:%s", string(data))
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		log.Errorf("Unmarshal error:%v", err)
		return err
	}

	for k, v := range m {
		msg.DataMap[k] = v
	}

	return nil
}

func (e *jsonEditor) Handler(msg *meta.Message, argv map[string]interface{}) error {
	c := parseArgs(argv)
	log.Debugf("fields:%v", c.Fields)
	if len(c.Fields) == 0 {
		return e.jsonDecode(msg, []byte(msg.Source))
	}

	for _, f := range c.Fields {
		do, ok := msg.DataMap[f]
		if !ok {
			log.Errorf("key:%v not found", f)
			return fmt.Errorf("key:%v not found", f)
		}
		ds := do.(string)
		if err := e.jsonDecode(msg, []byte(ds)); err != nil {
			return err
		}
	}
	return nil
}
