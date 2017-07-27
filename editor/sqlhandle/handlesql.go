package sqlhandle

import (
	"encoding/base64"
	"fmt"

	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/sqlparser"
	"github.com/zssky/log"

	"github.com/dearcode/watcher/editor"
	"github.com/dearcode/watcher/meta"
)

var (
	hs *handlesqlEditor
)

type handlesqlEditor struct {
}

func init() {
	editor.Register("sqlhandle", &handlesqlEditor{})
}

func (r *handlesqlEditor) Handler(msg *meta.Message, m map[string]interface{}) error {
	sql, ok := msg.DataMap["sql"]
	if !ok {
		log.Infof("sql not found")
		return nil
	}

	stmt, err := sqlparser.Parse(sql.(string))
	if err != nil {
		log.Error("解析语法树失败！！")
		return err
	}

	if bv, ok := msg.DataMap["bindVal"]; ok {
		UnNormalize(stmt, bv.(map[string]interface{}))
	}

	log.Debugf("stmt:%v", sqlparser.String(stmt))

	return analysis(msg)

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

// analysis 通过语法树分析sql的类型，表明，条件
func analysis(msg *meta.Message) error {
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
			node.(*sqlparser.Update).TableExprs.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Update).Where.Format(whereBuf)
			msg.DataMap["action"] = "Update"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil
		case *sqlparser.Delete:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Delete).TableExprs.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Delete).Where.Format(whereBuf)
			msg.DataMap["action"] = "Delete"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil

		case *sqlparser.Insert:
			whereBuf := sqlparser.NewTrackedBuffer(nil)

			msg.DataMap["action"] = "Insert"
			msg.DataMap["table"] = node.(*sqlparser.Insert).Table.Name.String()
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

		default:
			return true, nil

		}
	}, tree)
	log.Infof("%v", msg.DataMap)

	return nil
}
