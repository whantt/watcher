package main

import (
	"fmt"

	"github.com/dearcode/tracker/config"
	"github.com/dearcode/tracker/editor"
	_ "github.com/dearcode/tracker/editor/json"
	_ "github.com/dearcode/tracker/editor/remove"
	"github.com/dearcode/tracker/harvester"
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

    for msg := range  harvester.Reader() {
        fmt.Printf("msg:%v\n", msg.Source)
        editor.Run(msg)

        fmt.Printf("result:%v", msg.TraceStack())
        for k, v := range msg.DataMap {
            fmt.Printf("key:%v, value:%v\n", k, v)
        }
    }


}
