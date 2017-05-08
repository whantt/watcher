package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"text/template"

	"github.com/zssky/log"
	"gopkg.in/gomail.v2"

	"github.com/dearcode/watcher/alertor"
	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/meta"
)

var (
	ma = mailAlertor{}
)

type mailAlertor struct {
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

	title := buf.String()

	buf.Truncate(0)

	if t, err = template.New("mail").Parse(ac.MailBody); err != nil {
		log.Errorf("parse mail body error:%v, src:%v", err, ac.MailTitle)
		return err
	}
	if err = t.Execute(buf, msg.DataMap); err != nil {
		log.Errorf("Execute mail body error:%v, src:%v", err, ac.MailTitle)
		return err
	}

	body := buf.String()

	return ma.send(msg, ac.MailTo, title, body)
}

func (ma *mailAlertor) send(msg *meta.Message, to []string, title, body string) error {
	ec, err := config.GetConfig()
	if err != nil {
		return err
	}

	msg.Trace(meta.StageAlertor, "mail", fmt.Sprintf("begin Dial:%v:%v", ec.Alertor.Mail.Host, ec.Alertor.Mail.Port))
	d := gomail.NewDialer(ec.Alertor.Mail.Host, ec.Alertor.Mail.Port, ec.Alertor.Mail.User, ec.Alertor.Mail.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	sc, err := d.Dial()
	if err != nil {
		msg.Trace(meta.StageAlertor, "mail", fmt.Sprintf("end Dial:%v:%v error:%v", ec.Alertor.Mail.Host, ec.Alertor.Mail.Port, err))
		return err
	}
	defer sc.Close()
	msg.Trace(meta.StageAlertor, "mail", fmt.Sprintf("end Dial:%v:%v success", ec.Alertor.Mail.Host, ec.Alertor.Mail.Port))

	m := gomail.NewMessage()
	m.SetHeader("From", ec.Alertor.Mail.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", body)

	msg.Trace(meta.StageAlertor, "mail", fmt.Sprintf("begin send from:%v to:%v", ec.Alertor.Mail.From, to))
	if err = sc.Send(ec.Alertor.Mail.From, to, m); err != nil {
		msg.Trace(meta.StageAlertor, "mail", fmt.Sprintf("end send from:%v to:%v, error:%v", ec.Alertor.Mail.From, to, err))
		return err
	}
	msg.Trace(meta.StageAlertor, "mail", fmt.Sprintf("end send from:%v to:%v, success", ec.Alertor.Mail.From, to))
	return nil
}
