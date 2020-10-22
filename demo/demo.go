package main

import (
	"fmt"

	"github.com/linhuman/sgf"
	"github.com/linhuman/sgf/common"
	"github.com/linhuman/sgf/db/mysql"
)

func stop_func() {
	fmt.Println("shutdown...")
}
func main() {
	sgf.Initialize(SgfConfig)
	db := mysql.GetInstance()
	db.Report(true)
	data, err, report := db.GetRow("SELECT * FROM t_test")
	fmt.Println(data, err, report) //map[id:3 user_id:11 username:test] <nil> [SQL]SELECT * FROM t_test [time]1ms

	data2, err, report := db.GetAll("SELECT * FROM t_test")
	fmt.Println(data2, err, report) //[map[id:3 user_id:11 username:test]] <nil> [SQL]SELECT * FROM t_test [time]

	last_insert_id, rows_affected, err, report := db.AutoInsert("t_test", map[string]interface{}{"user_id": 22, "username": "test2"}, mysql.INSERT_INTO)
	fmt.Println(last_insert_id, rows_affected, err, report) //4 1 <nil> [SQL]INSERT INTO t_test (`user_id`, `username`) VALUES (22, "test2") [time]4ms

	rs, err, report := db.AutoUpdate("t_test", map[string]interface{}{"user_id": 21}, map[string]interface{}{"id": 3})
	fmt.Println(rs, err, report) //1 <nil> [SQL]UPDATE t_test SET `user_id`=21 WHERE `id`=3 [time]1ms

	db2 := mysql.GetInstance("test")
	data, err, report = db2.GetRow("SELECT * FROM t_test")
	fmt.Println(data, err, report)

	common.HandleSignalToStop(stop_func)

}
