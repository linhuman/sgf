package config

import "time"

type DbCfgMap map[string]DbCfg
type RedisCfgMap map[string]RedisCfg

type Cfg struct {
	Db       DbCfgMap
	Redis    RedisCfgMap
	Log_path string
}
type DbCfg struct {
	Driver           string
	User             string
	Password         string
	Host             string
	Port             string
	Database         string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLiftetime time.Duration
}
type RedisCfg struct {
	Host        string
	Port        string
	Password    string
	Database    int
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
	Timeout     int
}

var Entity Cfg
