package harvester

import (
	"time"

	"github.com/dearcode/tracker/meta"
)

var (
	h *harvester
)

type harvester struct {
	c chan *meta.Message
}

func Init() error {
	h = &harvester{
		c: make(chan *meta.Message),
	}

	go runtest()

	return nil
}

func Reader() <-chan *meta.Message {
	return h.c
}

func runtest() {
	msg := []struct {
		topic string
		msg   string
	}{
		{"api_dbs", "2107/04/20 17:18:19 error mysql_rw select * from a"},
		{"api_dbs", "2107/04/20 17:18:19 info write file success"},
		{"sql", `{"user":"tianguangyu", "mail": "jltgy@qq.com", "age": 123}`},
        {"sql", `{"json_data":"{\"user\":\"tgy\", \"password\":\"abc\"}", "id": 123}`},
	}

	t := time.NewTicker(time.Second)
	idx := 0
	for range t.C {
		idx = (idx + 1) % len(msg)
		h.c <- meta.NewMessage(msg[idx].topic, msg[idx].msg)
	}
}
