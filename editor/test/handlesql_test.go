// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"errors"
	"fmt"
	"github.com/dearcode/watcher/editor/sqlhandle"
	"github.com/dearcode/watcher/meta"
	"strings"
	"testing"
)

func TestHandleSQL(t *testing.T) {

	dm := make(map[string]interface{})
	//dm["sql"] = `{"sql":"insert into user(id, v, name) values (:_Id0, 2, :_name0) /* vtgate:: keyspace_id:166b40b44aba4bd6 */","bindVal":{"_Id0":1,"__seq0":1,"_name0":"bXluYW1l"},"sendQueryDate":"01-1-1 00:00:00","recvResultDate":"17-4-30 20:00:23.724694493","sqlExecDuration":"2562047h47m16.854775807s","datanodes":[{"name":"-20","tabletType":1,"idx":1,"sendDate":"01-1-1 00:00:00","recvDate":"17-4-30 20:00:23.724691915","shardExecuteTime":"2562047h47m16.854775807s"}]}`
	dm["sql"] = "insert into user(id, v, name) values (:_Id0, 2, :_name0) /* vtgate:: keyspace_id:166b40b44aba4bd6 */"
	dm["json_data"] = `{"name":"mysql_rw","addr":"127.0.0.1:35636","sql":"insert into test1(ID, NAME, GENDER, AGE, CITY) values (:_ID0, :vtg2, :vtg3, :vtg4, :vtg5) /* vtgate:: keyspace_id:63c674a6c0467543 */","bindVal":{"_ID0":{"type":10262,"value":"Mg=="},"vtg1":{"type":10262,"value":"Mg=="},"vtg2":{"type":10262,"value":"d2FuZ2dhb3FpYW5n"},"vtg3":{"type":10262,"value":"55S3"},"vtg4":{"type":10262,"value":"MTk="},"vtg5":{"type":10262,"value":"5rKz5Y2X"}},"sendQueryDate":"17-4-30 20:27:49.968858588","recvResultDate":"17-4-30 20:27:49.986763601","sqlExecDuration":"17.905013ms","datanodes":[{"name":"0","tabletType":1,"idx":1,"sendDate":"17-4-30 20:27:49.968858588","recvDate":"17-4-30 20:27:49.986755798","shardExecuteTime":"17.89721ms"}]}`
	msg := &meta.Message{
		DataMap: dm,
	}
	sqlhandle.HandleSql(msg)
	fmt.Println("action ", msg.DataMap["action"])
	fmt.Println("table", msg.DataMap["table"])

	fmt.Println("condition", msg.DataMap["condition"])

}
func TestEmptyFile(t *testing.T) {
	parentPath, err := substr("dd/aa", 0, strings.LastIndex("dd/aa", "/"))
	if err != nil {
		fmt.Println("配置文件路径: %s路径错误或者文件不存在")

	}
	fmt.Println(parentPath)
}
func substr(s string, pos, length int) (string, error) {

	if s == "" || length < 0 {
		//log.Errorf("配置文件路径: %s路径错误或者文件不存在", s)
		return s, errors.New("配置文件路径: " + s + "路径错误或者文件不存在")
	}
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l]), nil
}
