package main

import (
	"demo/controller"
	"github.com/linhuman/sgf/config"
	"os"
	"time"
)

var SgfConfig config.Cfg
var routers = [][]interface{}{
	{"/", controller.Test{}, "Hello"},
}
func init() {
	SgfConfig.Db = config.DbCfgMap{}
	SgfConfig.Db["default"] = config.DbCfg{"mysql", "root", "123456", "127.0.0.1", "3306", "test", 10, 5, time.Minute * 10}
	SgfConfig.Db["test"] = config.DbCfg{"mysql", "root", "123456", "127.0.0.1", "3306", "test", 10, 5, time.Minute * 10}
	pwd, _ := os.Getwd()
	SgfConfig.LogPath = pwd + "/log"
	SgfConfig.Routers = routers
}
