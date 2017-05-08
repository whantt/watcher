package json

import (
	"encoding/json"
	"fmt"

	"github.com/dearcode/crab/cache"
	"github.com/zssky/log"

	"github.com/dearcode/watcher/editor"
	"github.com/dearcode/watcher/meta"
)

var (
	je *jsonEditor
)

type jsonEditor struct {
	args *cache.Cache
}

func init() {
	je = &jsonEditor{
		args: cache.NewCache(5),
	}
	editor.Register("json", je)
}

type jsonArgv struct {
	Fields []string
}

func (j *jsonEditor) parseArgs(m map[string]interface{}) jsonArgv {
	var argv jsonArgv

	if m == nil {
		return argv
	}

	if o := j.args.Get(fmt.Sprintf("%p", m)); o != nil {
		return o.(jsonArgv)
	}

	if fo, ok := m["fields"]; ok {
		for _, fo := range fo.([]interface{}) {
			argv.Fields = append(argv.Fields, fo.(string))
		}
	}

	j.args.Add(fmt.Sprintf("%p", m), argv)

	return argv
}

func (j *jsonEditor) jsonDecode(msg *meta.Message, data []byte) error {
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

func (j *jsonEditor) Handler(msg *meta.Message, dm map[string]interface{}) error {
	argv := j.parseArgs(dm)

	if len(argv.Fields) == 0 {
		return j.jsonDecode(msg, []byte(msg.Source))
	}

	for _, f := range argv.Fields {
		do, ok := msg.DataMap[f]
		if !ok {
			log.Errorf("key:%v not found", f)
			return fmt.Errorf("key:%v not found", f)
		}
		ds := do.(string)
		if err := j.jsonDecode(msg, []byte(ds)); err != nil {
			return err
		}
	}
	return nil
}
