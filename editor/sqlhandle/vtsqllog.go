package sqlhandle
// 日志的信息的输出级别
const (
	INFO    = iota // 0
	WARNING        // 1
	ERROR          // 2
)

// SQLLogInfo sql审计的日志信息
type SQLLogInfo struct {
	Name            string                 `json:"name,omitempty"`
	Addr            string                 `json:"addr,omitempty"`
	ConnectionID    uint32                 `json:"connectionID,omitempty"`
	SQL             string                 `json:"sql,omitempty"`
	BindVal         map[string]interface{} `json:"bindVal,omitempty"`
	SendQueryDate   string                 `json:"sendQueryDate,omitempty"`
	RecvResultDate  string                 `json:"recvResultDate,omitempty"`
	SQLExecDuration string                 `json:"sqlExecDuration,omitempty"`

	ErrorDate string `json:"ErrorDate,omitempty"`

	Info    string `json:"info ,omitempty"`
	Error   string `json:"error,omitempty"`
	Warning string `json:"warning,omitempty"`

	Datanodes []*Datanode `json:"datanodes,omitempty"`
}

// Datanode sql审计的shard日志信息
type Datanode struct {
	Name              string `json:"name,omitempty"`
	TabletType        int32  `json:"tabletType,omitempty"`
	Idx               int32  `json:"idx,omitempty"`
	SendDate          string `json:"sendDate,omitempty"`
	RecvDate          string `json:"recvDate,omitempty"`
	ShardExecDuration string `json:"shardExecuteTime,omitempty"`
}
