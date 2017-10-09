package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/watcher/alertor"
	_ "github.com/dearcode/watcher/alertor/mail"
	_ "github.com/dearcode/watcher/alertor/message"
	"github.com/dearcode/watcher/config"
	"github.com/dearcode/watcher/editor"
	_ "github.com/dearcode/watcher/editor/json"
	_ "github.com/dearcode/watcher/editor/regexp"
	_ "github.com/dearcode/watcher/editor/remove"
	_ "github.com/dearcode/watcher/editor/rename"
	_ "github.com/dearcode/watcher/editor/sqlhandle"
	"github.com/dearcode/watcher/harvester"
	_ "github.com/dearcode/watcher/harvester/kafka"
	"github.com/dearcode/watcher/meta"
	"github.com/dearcode/watcher/processor"
)

var (
	debug = flag.Bool("debug", false, "output all logs.")
)

func main() {
	flag.Parse()
	log.SetLevel(log.LOG_LEVEL_INFO)

	if *debug {
		log.SetLevel(log.LOG_LEVEL_ALL)
	}

	msgChan := make(chan *meta.Message, 1)
	configChan := make(chan struct{}, 1)

	if err := config.Init(configChan); err != nil {
		panic(errors.ErrorStack(err))
	}

	if err := editor.Init(); err != nil {
		panic(errors.ErrorStack(err))
	}

	if err := harvester.Init(msgChan); err != nil {
		panic(errors.ErrorStack(err))
	}

	if err := processor.Init(); err != nil {
		panic(errors.ErrorStack(err))
	}

	if err := alertor.Init(); err != nil {
		panic(errors.ErrorStack(err))
	}

	for i := 0; i < 10; i++ {
		go worker(msgChan)
	}

	go func() {
		for range configChan {
			harvester.Stop()
			harvester.Init(msgChan)
		}
	}()

	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGUSR1, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)

	s := <-shutdown
	log.Warningf("recv signal %v, close.", s)
	harvester.Stop()
	log.Warningf("server exit")
}

func worker(reader <-chan *meta.Message) {
	for msg := range reader {
		run(msg)
	}
}

func run(msg *meta.Message) {
	msg.Trace(meta.StageEditor, "begin", msg.Source)
	if err := editor.Run(msg); err != nil {
		log.Errorf("editor run error:%v", err)
		log.Error(msg.TraceStack())
		return
	}

	msg.Trace(meta.StageProcessor, "begin", "")
	ac, err := processor.Run(msg)
	if err != nil {
		if err == processor.ErrNoMatch {
			return
		}
		log.Errorf("processor run error:%v", err)
		log.Error(msg.TraceStack())
		return
	}

	msg.Trace(meta.StageAlertor, "begin", "")
	if err = alertor.Run(msg, ac); err != nil {
		log.Errorf("alertor run error:%v", err)
		log.Error(msg.TraceStack())
		return
	}
	//	msg.Trace(meta.StageAlertor, "end", "OK")

}
