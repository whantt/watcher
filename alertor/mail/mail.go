package mail

import (
	"bytes"
	"text/template"

	"github.com/zssky/log"

	"github.com/dearcode/tracker/alertor"
	"github.com/dearcode/tracker/config"
	"github.com/dearcode/tracker/meta"
)

var (
	ma = mailAlertor{}
)

type mailAlertor struct {
	title string
	body  string
}

func init() {
	alertor.Register("mail", &ma)
}

func (ma *mailAlertor) Handler(msg *meta.Message, ac config.ActionConfig) error {
	buf := bytes.NewBufferString("")
	t, err := template.New("mail").Parse(ac.MailTitle)
	if err != nil {
		log.Errorf("parse mail title error:%v, src:%v", err, ac.MailTitle)
		return err
	}
	if err = t.Execute(buf, msg.DataMap); err != nil {
		log.Errorf("Execute mail title error:%v, src:%v", err, ac.MailTitle)
		return err
	}

	ma.title = buf.String()

	buf.Truncate(0)

	if t, err = template.New("mail").Parse(ac.MailBody); err != nil {
		log.Errorf("parse mail body error:%v, src:%v", err, ac.MailTitle)
		return err
	}
	if err = t.Execute(buf, msg.DataMap); err != nil {
		log.Errorf("Execute mail body error:%v, src:%v", err, ac.MailTitle)
		return err
	}

	ma.body = buf.String()

	return ma.send()
}

func (ma *mailAlertor) send() error {
	return nil
}
