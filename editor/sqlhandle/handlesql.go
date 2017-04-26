package sqlhandle

import (
	"github.com/dearcode/crab/cache"
	"github.com/zssky/log"

	"github.com/dearcode/tracker/editor"
	"github.com/dearcode/tracker/meta"
	"github.com/youtube/vitess/go/vt/sqlparser"
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
	return sql(msg)

}
func sql(msg *meta.Message) error {
	sql := msg.DataMap["sql"].(string)
	tree, err := sqlparser.Parse(sql)
	sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node.(type) {
		case *sqlparser.Select:
			buf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Select).From.Format(buf, sqlparser.NonReWriteSQL)
			buf1 := sqlparser.NewTrackedBuffer(nil)
			msg.DataMap["action"] = "Select"
			msg.DataMap["table"] = buf.String()
			msg.DataMap["condition"] = buf1.String()

			return false, nil

		case *sqlparser.Update:
			buf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Update).Table.Format(buf, sqlparser.NonReWriteSQL)
			buf1 := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Update).Where.Format(buf1, sqlparser.NonReWriteSQL)
			msg.DataMap["action"] = "Update"
			msg.DataMap["table"] = buf.String()
			msg.DataMap["condition"] = buf1.String()
			return false, nil

		case *sqlparser.Insert:
			buf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Insert).Table.Format(buf, sqlparser.NonReWriteSQL)
			buf1 := sqlparser.NewTrackedBuffer(nil)

			msg.DataMap["action"] = "Insert"
			msg.DataMap["table"] = buf.String()
			msg.DataMap["condition"] = buf1.String()
			return false, nil

		case *sqlparser.DDL:
			buf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.DDL).Table.Format(buf, sqlparser.NonReWriteSQL)

			msg.DataMap["action"] = node.(*sqlparser.DDL).Action
			msg.DataMap["table"] = buf.String()
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
	if err != nil {
		log.Errorf("input: %s, err: %v", sql, err)
		return err
	}
	return nil
}
