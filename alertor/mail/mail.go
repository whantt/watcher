package mail

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/dearcode/crab/http/client"
	"github.com/zssky/log"
	"gopkg.in/gomail.v2"

	"github.com/dearcode/watcher/alertor"
	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/meta"
)

var (
	ma      = mailAlertor{}
	webMail = flag.Bool("webmail", false, "use web mail api")
)

type mailAlertor struct {
}

func init() {
	alertor.Register("mail", &ma)
}

func (ma *mailAlertor) Handler(msg *meta.Message, ac config.ActionConfig) error {
	buf := bytes.NewBufferString("")
	t, _ := template.New("mail").Parse("日志平台报警")
	if err := t.Execute(buf, msg.DataMap); err != nil {
		log.Errorf("Execute mail title error:%v, src:%v", err, msg.DataMap)
		return err
	}

	title := buf.String()

	buf.Truncate(0)

	html := strings.Replace(ac.Message, "\n", "<br />", -1)

	t, err := template.New("mail").Parse(html)
	if err != nil {
		log.Errorf("parse mail body error:%v, src:%v", err, msg.DataMap)
		return err
	}
	if err = t.Execute(buf, msg.DataMap); err != nil {
		log.Errorf("Execute mail body error:%v, src:%v", err, msg.DataMap)
		return err
	}

	body := buf.String()
	if *webMail {
		return sendWeb(msg, ac.MailTo, title, body)
	}
	return sendMail(msg, ac.MailTo, title, body)

}

func sendMail(msg *meta.Message, to []string, title, body string) error {
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

func sendWeb(_ *meta.Message, to []string, title, body string) error {
	ec, err := config.GetConfig()
	if err != nil {
		return err
	}

	wa := struct {
		OrderID int    `json:"OrderId"`
		To      string `json:"toAddress"`
		CC      string `json:"ccAddress"`
		Subject string `json:"subject"`
		Content string `json:"content"`
	}{
		To:      strings.Join(to, ";"),
		Subject: title,
		Content: body,
		OrderID: 1,
	}

	buf, _ := json.Marshal(wa)

	buf, _, err = client.NewClient(time.Minute).POST(ec.Alertor.WebMail.URL, map[string]string{"Token": ec.Alertor.WebMail.Token}, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	log.Debugf("send mail:%v, response:%v", wa, string(buf))

	return nil
}
