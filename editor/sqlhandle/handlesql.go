package sqlhandle

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dearcode/crab/cache"
	"github.com/dearcode/tracker/editor"
	"github.com/dearcode/tracker/meta"
	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/sqlparser"
	"github.com/zssky/log"
)

var (
	rxe *handlesqlEditor
)

type handlesqlEditor struct {
	args *cache.Cache
}

func init() {
	rxe = &handlesqlEditor{
		args: cache.NewCache(3600),
	}
	editor.Register("sqlhandle", rxe)
}

func (r *handlesqlEditor) Handler(msg *meta.Message, m map[string]interface{}) error {

	err := HandleSql(msg)
	if err != nil {
		return err
	}
	return nil
}

func HandleSql(msg *meta.Message) error {

	//jsonLog := `{"name":"mysql_rw","addr":"127.0.0.1:35636","sql":"insert into test1(ID, NAME, GENDER, AGE, CITY) values (:_ID0, :vtg2, :vtg3, :vtg4, :vtg5) /* vtgate:: keyspace_id:63c674a6c0467543 */","bindVal":{"_ID0":{"type":10262,"value":"Mg=="},"vtg1":{"type":10262,"value":"Mg=="},"vtg2":{"type":10262,"value":"d2FuZ2dhb3FpYW5n"},"vtg3":{"type":10262,"value":"55S3"},"vtg4":{"type":10262,"value":"MTk="},"vtg5":{"type":10262,"value":"5rKz5Y2X"}},"sendQueryDate":"17-4-30 20:27:49.968858588","recvResultDate":"17-4-30 20:27:49.986763601","sqlExecDuration":"17.905013ms","datanodes":[{"name":"0","tabletType":1,"idx":1,"sendDate":"17-4-30 20:27:49.968858588","recvDate":"17-4-30 20:27:49.986755798","shardExecuteTime":"17.89721ms"}]}`
	slf := &SQLLogInfo{}
	//somethings := []map[string]interface{}{}
	//msg.DataMap["json_data"]=`{"sql":"insert into user(id, v, name) values (:_Id0, 2, :_name0) /* vtgate:: keyspace_id:166b40b44aba4bd6 */","bindVal":{"_Id0":1,"__seq0":1,"_name0":"bXluYW1l"},"sendQueryDate":"01-1-1 00:00:00","recvResultDate":"17-4-30 20:00:23.724694493","sqlExecDuration":"2562047h47m16.854775807s","datanodes":[{"name":"-20","tabletType":1,"idx":1,"sendDate":"01-1-1 00:00:00","recvDate":"17-4-30 20:00:23.724691915","shardExecuteTime":"2562047h47m16.854775807s"}]}`
	//aa:=`{"sql":"insert into user(id, v, name) values (:_Id0, 2, :_name0) /* vtgate:: keyspace_id:166b40b44aba4bd6 */","bindVal":{"_Id0":1,"__seq0":1,"_name0":"bXluYW1l"},"sendQueryDate":"01-1-1 00:00:00","recvResultDate":"17-4-30 20:00:23.724694493","sqlExecDuration":"2562047h47m16.854775807s","datanodes":[{"name":"-20","tabletType":1,"idx":1,"sendDate":"01-1-1 00:00:00","recvDate":"17-4-30 20:00:23.724691915","shardExecuteTime":"2562047h47m16.854775807s"}]}`
	json_data := fmt.Sprint(msg.DataMap["json_data"])
	err := json.Unmarshal([]byte(json_data), &slf)
	if err != nil {
		log.Errorf("%v", err)
		return err
	} else {
		stmt, err := sqlparser.Parse(slf.SQL)
		if err != nil {
			log.Error("解析语法树失败！！")
			return err
		}
		UnNormalize(stmt, slf.BindVal)
		out := sqlparser.String(stmt)
		fmt.Println(out)

	}
	msg.DataMap["sql"] = slf.SQL
	return sql(msg)

}

// UnNormalize changes the statement include bindbars  to oringnal sql by sql template and bindvars
func UnNormalize(stmt sqlparser.Statement, bindVals map[string]interface{}) {
	_ = sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node := node.(type) {
		case *sqlparser.SQLVal:
			bindValIndex := fmt.Sprint(string(node.Val))
			bindValIndex = bindValIndex[1:]
			bindValMap, ok := (bindVals[bindValIndex]).(map[string]interface{})
			if !ok {
				log.Error("bindVals类型失败")
				return true, nil
			}

			strVal := fmt.Sprint(bindValMap["value"])
			gotValType := fmt.Sprint(bindValMap["type"])
			wantValType := int(sqltypes.VarBinary)
			decodeStrVal, err := base64.StdEncoding.DecodeString(strVal)
			if err != nil {
				log.Error("base64 解码失败！")
			} else {
				strVal = string(decodeStrVal)
			}

			if gotValType == fmt.Sprint(wantValType) {

				node.Val = []byte("'" + strVal + "'")
			} else {
				node.Val = []byte(strVal)
			}

		}

		return true, nil
	}, stmt)
}

// sql 通过语法树分析sql的类型，表明，条件
func sql(msg *meta.Message) error {
	sql := msg.DataMap["sql"].(string)
	tree, err := sqlparser.Parse(sql)
	if err != nil {
		log.Errorf("input: %s, err: %v", sql, err)
		return err
	}
	sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node.(type) {
		case *sqlparser.Select:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Select).From.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Select).Where.Format(whereBuf)
			msg.DataMap["action"] = "Select"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil
		case *sqlparser.Update:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Update).Table.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Update).Where.Format(whereBuf)
			msg.DataMap["action"] = "Update"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil
		case *sqlparser.Delete:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Delete).Table.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Delete).Where.Format(whereBuf)
			msg.DataMap["action"] = "Delete"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil

		case *sqlparser.Insert:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Insert).Table.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)

			msg.DataMap["action"] = "Insert"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil

		case *sqlparser.DDL:
			action := node.(*sqlparser.DDL).Action
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			if action == sqlparser.DropStr {
				node.(*sqlparser.DDL).Table.Format(tableBuf)
				//tableBuf.Myprintf("%v", nil,node.(*sqlparser.DDL).Table.Name) //忽略database name

			} else {
				node.(*sqlparser.DDL).NewName.Format(tableBuf)
			}

			msg.DataMap["action"] = action
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = ""
			return false, nil

		case *sqlparser.Other:
			msg.DataMap["action"] = "Other"
			msg.DataMap["table"] = "Other"
			msg.DataMap["condition"] = "Other"
			return false, nil
		default:
			return true, nil

		}
	}, tree)
	log.Infof("%v", msg.DataMap)

	return nil
}
