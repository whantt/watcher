package message

import (
	"bytes"
	"text/template"

	"github.com/zssky/log"

	"github.com/dearcode/tracker/alertor"
	"github.com/dearcode/tracker/config"
	"github.com/dearcode/tracker/meta"
)

var (
	ma = messageAlertor{}
)

type messageAlertor struct {
	body string
}

func init() {
	alertor.Register("message", &ma)
}

func (ma *messageAlertor) Handler(msg *meta.Message, ac config.ActionConfig) error {
	buf := bytes.NewBufferString("")
	t, err := template.New("message").Parse(ac.MessageBody)
	if err != nil {
		log.Errorf("parse message body error:%v, src:%v", err, ac.MessageBody)
		return err
	}
	if err = t.Execute(buf, msg.DataMap); err != nil {
		log.Errorf("Execute message body error:%v, src:%v", err, ac.MessageBody)
		return err
	}
	ma.body = buf.String()

	return ma.send()
}

//send TODO
func (ma *messageAlertor) send() error {
	return nil
}
