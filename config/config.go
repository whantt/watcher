package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"time"
)

var (
	cf  = flag.String("c", "./config/tracker.json", "config path")
	cfg *Config
	cs  os.FileInfo

	// ErrConfigNeedInit config need init first.
	ErrConfigNeedInit = errors.New("config need init first")
)

type HarvesterConfig struct {
	Brokers []string `json:"brokers"`
	Topics  []string `json:"topics"`
	Group   string   `json:"group"`
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
	Mail        bool     `json:"mail"`
	MailTo      []string `json:"mail_to"`
	MailTitle   string   `json:"mail_title"`
	MailBody    string   `json:"mail_body"`
	Message     bool     `json:"message"`
	MessageTo   []string `json:"message_to"`
	MessageBody string   `json:"message_body"`
}

//RulesConfig 一个过滤事件.
type RulesConfig struct {
	Model  string        `json:"model"`
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

type AlertorConfig struct {
	Mail    AlertorMailConfig    `json:"mail"`
	Message AlertorMessageConfig `json::"message"`
}

type Config struct {
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

	return nil
}

func reloadConfig() {
	t := time.NewTicker(time.Second)

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
