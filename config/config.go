package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/dearcode/crab/http/client"
	"github.com/dearcode/crab/util"
	"github.com/juju/errors"
	"github.com/zssky/log"
)

var (
	cf  = flag.String("cfg", "./config/watcher.json", "config path")
	cfg *Config
	cs  os.FileInfo

	// ErrConfigNeedInit config need init first.
	ErrConfigNeedInit = errors.New("config need init first")
)

const (
	reloadInterval = time.Second * 5
)

// HarvesterConfig 日志收集.
type HarvesterConfig struct {
	Brokers  []string `json:"brokers"`
	Topics   []string `json:"topics"`
	Group    string   `json:"group"`
	ClientID string   `json:"client_id"`
}

//EditorConfig 对日志内容进行修改.
type EditorConfig struct {
	Topics []string               `json:"topics"`
	Model  string                 `json:"model"`
	Data   map[string]interface{} `json:"data"`
}

//MatchConfig 检测日志内容与配置报警是否匹配.
type MatchConfig struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Method string `json:"method"`
}

//ActionConfig 配置成功后要执行的动作.
type ActionConfig struct {
	MailTo    []string `json:"mail_to"`
	MessageTo []string `json:"message_to"`
	Message   string   `json:"message"`
}

//RulesConfig 一个过滤事件.
type RulesConfig struct {
	Match  []MatchConfig `json:"match"`
	Action ActionConfig  `json:"action"`
}

//ProcessorConfig 根据规则遍历日志.
type ProcessorConfig struct {
	Topics []string      `json:"topics"`
	Rules  []RulesConfig `json:"rules"`
}

//AlertorMailConfig 报警邮件相关配置.
type AlertorMailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	From     string `json:"from"`
	Password string `json:"password"`
}

//AlertorMessageConfig 报警短信配置.
type AlertorMessageConfig struct {
	URL       string `json:"url"`
	Account   string `json:"account"`
	Extension string `json:"extension"`
}

//AlertorWebMailConfig 报警邮件(webmail)配置.
type AlertorWebMailConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

//AlertorConfig 日志报警相关配置.
type AlertorConfig struct {
	Mail    AlertorMailConfig    `json:"mail"`
	Message AlertorMessageConfig `json:"message"`
	WebMail AlertorWebMailConfig `json:"webmail"`
}

//ManagerConfig 管理节点配置.
type ManagerConfig struct {
	Host string `json:"host"`
}

//Config 全部配置.
type Config struct {
	Manager   ManagerConfig     `json:"manager"`
	Harvester HarvesterConfig   `json:"harvester"`
	Editor    []EditorConfig    `json:"editor"`
	Processor []ProcessorConfig `json:"processor"`
	Alertor   AlertorConfig     `json:"alertor"`
}

//Init 加载配置文件
func Init(configChan chan<- struct{}) error {
	if err := loadConfig(); err != nil {
		return errors.Trace(err)
	}

	if err := loadProcessor(configChan); err != nil {
		return errors.Trace(err)
	}

	go reloadConfig()
	go reloadProcessor(configChan)

	return nil
}

//ModuleKey 生成kafka中topic名.
func ModuleKey(app, module string) string {
	return fmt.Sprintf("a-%v-m-%v", app, module)
}

func topics(name, modules string) []string {
	var ts []string
	for _, m := range strings.Split(modules, ",") {
		t := ModuleKey(name, m)
		ts = append(ts, t)
	}

	return ts
}

func match(cond string) []MatchConfig {
	ms := []MatchConfig{}
	for _, c := range util.TrimSplit(cond, " && ") {
		kvs := util.TrimSplit(c, " ")
		if len(kvs) < 2 {
			log.Errorf("invalid line:%v", c)
			continue
		}

		m := MatchConfig{}

		//目前只支持常见少量操作
		switch kvs[1] {
		case "==":
			m.Method = "equal"
		case "=":
			m.Method = "contains"
		case ">":
			m.Method = "larger"
		case "<":
			m.Method = "lesser"
		}
		m.Key = kvs[0]
		m.Value = kvs[2]

		ms = append(ms, m)
	}
	return ms
}

var (
	oldBuf []byte
)

//AlertConfig 报警配置.
type AlertConfig struct {
	Name      string
	APP       string
	Modules   string
	Condition string
	Email     string
	Mobile    string
	Message   string
}

func loadProcessorConfig(url string) ([]ProcessorConfig, []string, error) {
	buf, _, err := client.New(time.Second*5).Get(url, nil, nil)
	if err != nil {
		return nil, nil, errors.Annotatef(err, "url:%v", url)
	}

	if bytes.Equal(oldBuf, buf) {
		return nil, nil, nil
	}

	acs := []AlertConfig{}
	if err = json.Unmarshal(buf, &acs); err != nil {
		return nil, nil, errors.Annotatef(err, "url:%v, buf:%s", url, buf)
	}

	log.Debugf("AlertConfig:%+v", acs)

	tpss := make(map[string]interface{})
	pcs := []ProcessorConfig{}
	for _, ac := range acs {
		rs := RulesConfig{
			Action: ActionConfig{
				MailTo:    util.TrimSplit(ac.Email, ","),
				MessageTo: util.TrimSplit(ac.Mobile, ","),
				Message:   ac.Message,
			},
			Match: match(ac.Condition),
		}

		tps := topics(ac.APP, ac.Modules)
		pc := ProcessorConfig{
			Topics: tps,
			Rules:  []RulesConfig{rs},
		}
		pcs = append(pcs, pc)
		for _, tp := range tps {
			tpss[tp] = nil
		}
	}

	ntpss := []string{}
	for k := range tpss {
		ntpss = append(ntpss, k)
	}

	oldBuf = buf

	return pcs, ntpss, nil
}

func reloadConfig() {
	t := time.NewTicker(reloadInterval)

	for range t.C {
		s, err := os.Stat(*cf)
		if err != nil {
			//TODO
			fmt.Printf("error:%v\n", err)
			continue
		}

		if !reflect.DeepEqual(s, cs) {
			loadConfig()
			cs = s
		}
	}
}

func loadProcessor(configChan chan<- struct{}) error {
	url := fmt.Sprintf("http://%v/api/alerts/", cfg.Manager.Host)
	p, tps, err := loadProcessorConfig(url)
	if err != nil {
		log.Errorf("loadProcessorConfig error:%v", errors.ErrorStack(err))
		return errors.Trace(err)
	}

	if p == nil && tps == nil {
		return nil
	}

	cfg.Processor = p
	cfg.Harvester.Topics = tps
	configChan <- struct{}{}
	log.Infof("new config:%+v", cfg)
	return nil
}

func reloadProcessor(configChan chan<- struct{}) {
	t := time.NewTicker(reloadInterval)

	for range t.C {
		if err := loadProcessor(configChan); err != nil {
			log.Errorf("loadProcessor error:%v", errors.ErrorStack(err))
		}
	}
}

func loadConfig() error {
	s, err := os.Stat(*cf)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadFile(*cf)
	if err != nil {
		return err
	}

	c := Config{}
	if err = json.Unmarshal(buf, &c); err != nil {
		return err
	}

	cfg = &c
	cs = s
	return nil
}

//GetConfig 获取配置文件，每次使用都要执行这个函数以便更新配置文件.
func GetConfig() (*Config, error) {
	if cfg == nil {
		return nil, ErrConfigNeedInit
	}
	return cfg, nil
}

//EnableMail 是否启用mail.
func (ac ActionConfig) EnableMail() bool {
	return len(ac.MailTo) > 0 && ac.MailTo[0] != ""
}

//EnableMessage 是否发短信.
func (ac ActionConfig) EnableMessage() bool {
	return len(ac.MessageTo) > 0 && ac.MessageTo[0] != ""
}
