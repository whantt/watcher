package message

import (
	"bytes"
	"text/template"

	"github.com/zssky/Mole/models/sms"
	"github.com/zssky/log"
	"github.com/zssky/tc"

	"github.com/dearcode/watcher/alertor"
	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/meta"
)

var (
	ma = messageAlertor{}
)

type messageAlertor struct {
}

func init() {
	alertor.Register("message", &ma)
}

func (ma *messageAlertor) Handler(msg *meta.Message, ac config.ActionConfig) error {
	buf := bytes.NewBufferString("")
	t, err := template.New("message").Parse(ac.Message)
	if err != nil {
		log.Errorf("parse message body error:%v, src:%v", err, ac.Message)
		return err
	}
	if err = t.Execute(buf, msg.DataMap); err != nil {
		log.Errorf("Execute message body error:%v, src:%v", err, ac.Message)
		return err
	}

	return ma.send(ac.MessageTo, buf.String())
}

func (ma *messageAlertor) send(to []string, body string) error {
	if len(to) == 0 {
		log.Infof("message to is null")
		return nil
	}
	ec, err := config.GetConfig()
	if err != nil {
		return err
	}

	is := &sms.SMS{
		SMSBaseInfo: sms.SMSBaseInfo{
			SenderNum: ec.Alertor.Message.Account,
			Extension: ec.Alertor.Message.Extension,
		},
		MobileNums: make([]sms.MobileInfo, len(to)),
		MsgContent: body,
	}

	for i, m := range to {
		is.MobileNums[i] = sms.MobileInfo{MobileNum: tc.TrimSpace(m)}
	}

	if err = sms.SendSMS(ec.Alertor.Message.URL, is); err != nil {
		log.Errorf("send sms error:%v", err)
		return err
	}

	return nil
}
