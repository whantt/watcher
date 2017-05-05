package meta

import (
	"bytes"
	"fmt"
	"time"

	"github.com/dearcode/tracker/config"
)

type ProcessStage int

const (
	StageHarvester ProcessStage = iota
	StageEditor
	StageProcessor
	StageAlertor
	StageMAX
)

type ProcessState int

const (
	StateRun ProcessState = iota
	StateDrop
	StateAlter
	StateError
	StateOver
)

type traceStatus struct {
	Time   time.Time
	Model  string
	Result string
}

type traceInfo struct {
	Status []traceStatus
}

// Message 一行日志！
type Message struct {
	Topic   string
	Source  string
	Notice  string
	DataMap map[string]interface{}
	state   ProcessState
	trace   [StageMAX]traceInfo
	action  config.ActionConfig
}

func NewMessage(topic, log string) *Message {
	return &Message{
		Topic:   topic,
		Source:  log,
		DataMap: make(map[string]interface{}),
	}
}

func (m *Message) SetState(state ProcessState) {
	m.state = state
}

func (m *Message) Runable() bool {
	return m.state == StateRun
}

func (m *Message) Trace(stage ProcessStage, model, result string) {
	i := &m.trace[stage]
	i.Status = append(i.Status, traceStatus{Model: model, Result: result, Time: time.Now()})
}

func (s ProcessStage) String() string {
	switch s {
	case StageHarvester:
		return "Harvester"
	case StageEditor:
		return "Editor"
	case StageProcessor:
		return "Processor"
	case StageAlertor:
		return "Alertor"
	}
	return "Undefine"
}

func (s ProcessState) String() string {
	switch s {
	case StateRun:
		return "Run"
	case StateDrop:
		return "Drop"
	case StateAlter:
		return "Alter"
	case StateError:
		return "Error"
	case StateOver:
		return "Over"
	}
	return "Undefine"
}

func (m *Message) TraceStack() string {
	buf := bytes.NewBufferString("")

	for i := StageHarvester; i < StageMAX; i++ {
		for _, s := range m.trace[i].Status {
			fmt.Fprintf(buf, "%v %v %v %v\n", s.Time, i, s.Model, s.Result)
		}
	}
	return buf.String()
}
