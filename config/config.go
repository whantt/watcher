package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/dearcode/configurator/models"
	"github.com/dearcode/configurator/util/keys"
	"github.com/dearcode/crab/http/client"
	"github.com/dearcode/crab/util"
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

type HarvesterConfig struct {
	Brokers  []string `json:"brokers"`
	Topics   []string `json:"topics"`
	Group    string   `json:"group"`
	ClientID string   `json:"client_id"`
}

type EditorConfig struct {
	Topics []string               `json:"topics"`
	Model  string                 `json:"model"`
	Data   map[string]interface{} `json:"data"`
}

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

type ProcessorConfig struct {
	Topics []string      `json:"topics"`
	Rules  []RulesConfig `json:"rules"`
}

type AlertorMailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	From     string `json:"from"`
	Password string `json:"password"`
}

type AlertorMessageConfig struct {
	URL       string `json:"url"`
	Account   string `json:"account"`
	Extension string `json:"extension"`
}

type AlertorWebMailConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type AlertorConfig struct {
	Mail    AlertorMailConfig    `json:"mail"`
	Message AlertorMessageConfig `json:"message"`
	WebMail AlertorWebMailConfig `json:"webmail"`
}

type ManagerConfig struct {
	Host string `json:"host"`
}

type Config struct {
	Manager   ManagerConfig     `json:"manager"`
	Harvester HarvesterConfig   `json:"harvester"`
	Editor    []EditorConfig    `json:"editor"`
	Processor []ProcessorConfig `json:"processor"`
	Alertor   AlertorConfig     `json:"alertor"`
}

//Init 加载配置文件
func Init() error {
	if err := loadConfig(); err != nil {
		return err
	}

	go reloadConfig()
	go reloadProcessor()

	return nil
}

func topics(name, modules string) []string {
	var ts []string
	for _, m := range strings.Split(modules, ",") {
		t := keys.ModuleKey(name, m)
		ts = append(ts, t)
	}

	return ts
}

func match(cond string) []MatchConfig {
	ms := []MatchConfig{}
	for _, c := range util.TrimSplit(cond, ",") {
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

func loadProcessor(url string) ([]ProcessorConfig, []string, error) {
	buf, _, err := client.NewClient(time.Second*5).Get(url, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	if bytes.Equal(oldBuf, buf) {
		return nil, nil, nil
	}

	acs := []models.AlertConfig{}
	if err = json.Unmarshal(buf, &acs); err != nil {
		return nil, nil, err
	}

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

		tps := topics(ac.Name, ac.Modules)
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

func reloadProcessor() {
	t := time.NewTicker(reloadInterval)
	url := fmt.Sprintf("http://%v/api/alerts/", cfg.Manager.Host)

	for range t.C {
		p, tps, err := loadProcessor(url)
		if err != nil {
			log.Errorf("loadProcessor error:%v", err)
			continue
		}
		if p == nil && tps == nil {
			continue
		}
		cfg.Processor = p
		cfg.Harvester.Topics = tps
		log.Debugf("new config:%+v", cfg)
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

func (ac ActionConfig) EnableMail() bool {
	return len(ac.MailTo) > 0 && ac.MailTo[0] != ""
}

func (ac ActionConfig) EnableMessage() bool {
	return len(ac.MessageTo) > 0 && ac.MessageTo[0] != ""
}
