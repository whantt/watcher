// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"fmt"
	"github.com/dearcode/tracker/editor/sqlhandle"
	"github.com/dearcode/tracker/meta"
	"testing"
)

func TestHandleSQL(t *testing.T) {

	dm := make(map[string]interface{})
	dm["sql"] = "select ff.freight_type as freightType, ff.id as freightId, ff.yn as freightYn from fms_freight as ff where ff.id = 2 and ff.route_id =1 "
	//dm["sql"] = "delete from fms_freight where id=100"


	msg := &meta.Message{
		DataMap: dm,
	}
	sqlhandle.HandleSql(msg)
	fmt.Println("action ", msg.DataMap["action"])
	fmt.Println("table", msg.DataMap["table"])

	fmt.Println("condition", msg.DataMap["condition"])

}
