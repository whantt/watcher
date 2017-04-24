package main

import (
	"github.com/zssky/log"

	"github.com/dearcode/tracker/alertor"
	_ "github.com/dearcode/tracker/alertor/mail"
	_ "github.com/dearcode/tracker/alertor/message"
	"github.com/dearcode/tracker/config"
	"github.com/dearcode/tracker/editor"
	_ "github.com/dearcode/tracker/editor/json"
	_ "github.com/dearcode/tracker/editor/regexp"
	_ "github.com/dearcode/tracker/editor/remove"
	"github.com/dearcode/tracker/harvester"
	"github.com/dearcode/tracker/meta"
	"github.com/dearcode/tracker/processor"
	_ "github.com/dearcode/tracker/processor/common"
)

func main() {
	if err := config.Init(); err != nil {
		panic(err.Error())
	}

	if err := editor.Init(); err != nil {
		panic(err.Error())
	}

	if err := harvester.Init(); err != nil {
		panic(err.Error())
	}

	if err := processor.Init(); err != nil {
		panic(err.Error())
	}

	if err := alertor.Init(); err != nil {
		panic(err.Error())
	}

	//正式应该多线程
	worker(harvester.Reader())
}

func worker(msg <-chan *meta.Message) {
	for msg := range harvester.Reader() {
		run(msg)
		log.Infof("msg trace:%v", msg.TraceStack())
	}
}

func run(msg *meta.Message) {
	msg.Trace(meta.StageEditor, "begin", msg.Source)
	if err := editor.Run(msg); err != nil {
		msg.Trace(meta.StageEditor, "end", err.Error())
		return
	}

	msg.Trace(meta.StageProcessor, "begin", "")
	ac, err := processor.Run(msg)
	if err != nil {
		msg.Trace(meta.StageProcessor, "end", err.Error())
		return
	}

	msg.Trace(meta.StageAlertor, "begin", "")
	if err = alertor.Run(msg, ac); err != nil {
		msg.Trace(meta.StageAlertor, "end", err.Error())
		return
	}
	msg.Trace(meta.StageAlertor, "end", "OK")

}
