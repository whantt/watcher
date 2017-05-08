package regexp

import (
	"fmt"
	re "regexp"

	"github.com/dearcode/crab/cache"
	"github.com/zssky/log"

	"github.com/dearcode/watcher/editor"
	"github.com/dearcode/watcher/meta"
)

var (
	rxe *regexpEditor
)

type regexpEditor struct {
	args *cache.Cache
}

type regexpConfig struct {
	key    string
	fields []string
	rx     *re.Regexp
}

func init() {
	rxe = &regexpEditor{
		args: cache.NewCache(3600),
	}
	editor.Register("regexp", rxe)
}

func (r *regexpEditor) parseArgv(m map[string]interface{}) regexpConfig {
	var rc regexpConfig
	if m == nil {
		return rc
	}

	if o := r.args.Get(fmt.Sprintf("%p", m)); o != nil {
		return o.(regexpConfig)
	}

	if fs, ok := m["key"]; ok {
		rc.key = fs.(string)
	}

	if fo, ok := m["fields"]; ok {
		for _, fo := range fo.([]interface{}) {
			rc.fields = append(rc.fields, fo.(string))
		}
	}

	if fs, ok := m["regexp"]; ok {
		rx, err := re.Compile(fs.(string))
		if err != nil {
			log.Errorf("invalid regexp:%v", fs)
			return rc
		}
		rc.rx = rx
	}

	r.args.Add(fmt.Sprintf("%p", m), rc)

	return rc
}

func (r *regexpEditor) Handler(msg *meta.Message, m map[string]interface{}) error {
	rc := r.parseArgv(m)

	data := msg.Source
	if rc.key != "" {
		if do, ok := msg.DataMap[rc.key]; ok {
			data = do.(string)
		}
	}

	log.Debugf("m:%v, data:%v", m, data)

	for i, v := range rc.rx.FindStringSubmatch(data) {
		//第一个是原字符串
		if i == 0 {
			continue
		}

		key := fmt.Sprintf("field_%d", i)
		if i <= len(rc.fields) {
			key = rc.fields[i-1]
		}

		msg.DataMap[key] = v
	}

	return nil
}
