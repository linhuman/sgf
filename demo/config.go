package main

import (
	"os"
	"time"

	"github.com/linhuman/sgf/config"
)

var SgfConfig config.Cfg

func init() {
	SgfConfig.Db = config.DbCfgMap{}
	SgfConfig.Db["default"] = config.DbCfg{"mysql", "root", "123456", "127.0.0.1", "3306", "test", 10, 5, time.Minute * 10}
	SgfConfig.Db["test"] = config.DbCfg{"mysql", "root", "123456", "127.0.0.1", "3306", "test", 10, 5, time.Minute * 10}
	pwd, _ := os.Getwd()
	SgfConfig.Log_path = pwd + "/log"
}
