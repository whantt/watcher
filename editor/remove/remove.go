package json

import (
	"github.com/zssky/log"

	"github.com/dearcode/tracker/editor"
	"github.com/dearcode/tracker/meta"
)

var (
	re = removeEditor{}
)

type removeEditor struct {
}

func init() {
	editor.Register("remove", &re)
}

type removeConfig struct {
	Fields []string
}

func parseArgs(argv map[string]interface{}) removeConfig {
	rc := removeConfig{}

	if fo, ok := argv["fields"]; ok {
		for _, fo := range fo.([]interface{}) {
			rc.Fields = append(rc.Fields, fo.(string))
		}
	}

	return rc
}

func (e *removeEditor) Handler(msg *meta.Message, argv map[string]interface{}) error {
	c := parseArgs(argv)

	for _, f := range c.Fields {
		if _, ok := msg.DataMap[f]; ok {
			delete(msg.DataMap, f)
			log.Debugf("remove key:%v", f)
		}
	}
	return nil
}
