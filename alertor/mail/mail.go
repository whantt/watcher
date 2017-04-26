package mail

import (
	"bytes"
	"crypto/tls"
	"text/template"

	"github.com/zssky/log"
	"gopkg.in/gomail.v2"

	"github.com/dearcode/tracker/alertor"
	"github.com/dearcode/tracker/config"
	"github.com/dearcode/tracker/meta"
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

	return ma.send(ac.MailTo, title, body)
}

func (ma *mailAlertor) send(to []string, title, body string) error {
	ec, err := config.GetConfig()
	if err != nil {
		return err
	}

	log.Debugf("config:%v", ec.Alertor.Mail)
	d := gomail.NewDialer(ec.Alertor.Mail.Host, ec.Alertor.Mail.Port, ec.Alertor.Mail.User, ec.Alertor.Mail.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	sc, err := d.Dial()
	if err != nil {
		return err
	}
	defer sc.Close()

	m := gomail.NewMessage()
	m.SetHeader("From", ec.Alertor.Mail.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", body)

	if err = sc.Send(ec.Alertor.Mail.From, to, m); err != nil {
		return err
	}
	return nil
}
