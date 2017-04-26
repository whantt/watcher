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
	dm["sql"] = "select * from t where id>3;"

	msg := &meta.Message{
		DataMap: dm,
	}
	sqlhandle.HandleSql(msg)
	fmt.Println(msg)
}
