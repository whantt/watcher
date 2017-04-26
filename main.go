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
	_ "github.com/dearcode/tracker/editor/sqlhandle"
	"github.com/dearcode/tracker/harvester"
	_ "github.com/dearcode/tracker/harvester/kafka"
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
	m := meta.NewMessage("sql", `I0426 11:21:40.488165      39 sql_log.go:54] json_data:{"name":"mysql_rw","addr":"192.168.81.31:48790","sql":"select ff.freight_type as freightType, ff.id as freightId, ff.yn as freightYn from fms_freight as ff where ff.id = :fsft_freight_id and ff.route_id = :vtg1","sendQueryDate":"17-4-26 11:21:40.482482612","recvResultDate":"17-4-26 11:21:40.488147166","sqlExecDuration":5664554,"datanodes":[{"name":"-50","tabletType":1,"idx":1,"sendDate":"17-4-26 11:21:40.482482612","recvDate":"17-4-26 11:21:40.488141224","shardExecuteTime":5658612}]}`)
	/*
		for msg := range harvester.Reader() {
			run(msg)
			log.Infof("msg trace:%v", msg.TraceStack())
		}
	*/
	run(m)
	log.Debugf("trace:%v", m.TraceStack())
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
