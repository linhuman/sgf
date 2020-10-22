package redis

import (
	"fmt"
	"sgf/config"
	"sync"
	"time"

	red "github.com/gomodule/redigo/redis"
)

var instance map[string]*red.Pool

var once sync.Once

type redis struct {
	pool *red.Pool
}

func initialize() {
	instance = make(map[string]*red.Pool)
	for k, _ := range config.Entity.Redis {
		instance[k] = &red.Pool{
			MaxIdle:     config.Entity.Redis[k].MaxIdle,
			MaxActive:   config.Entity.Redis[k].MaxActive,
			IdleTimeout: time.Duration(config.Entity.Redis[k].IdleTimeout),
			Dial: func() (red.Conn, error) {
				return red.Dial(
					"tcp",
					config.Entity.Redis[k].Host+":"+config.Entity.Redis[k].Port,
					red.DialReadTimeout(time.Duration(config.Entity.Redis[k].Timeout)*time.Millisecond),
					red.DialWriteTimeout(time.Duration(config.Entity.Redis[k].Timeout)*time.Millisecond),
					red.DialConnectTimeout(time.Duration(config.Entity.Redis[k].Timeout)*time.Millisecond),
					red.DialDatabase(config.Entity.Redis[k].Database),
					red.DialPassword(config.Entity.Redis[k].Password),
				)
			},
		}
	}
}
func GetInstance(field ...interface{}) *redis {
	once.Do(initialize)
	obj := new(redis)
	pool_name := "default"
	if 0 < len(field) {
		pool_name = field[0].(string)
	}
	if nil == instance[pool_name] {
		panic("数据库连接[" + pool_name + "]不存在")
	}
	obj.pool = instance[pool_name]

	return obj
}

func (r *redis) Exec(cmd string, key interface{}, args ...interface{}) (interface{}, error) {
	con := r.pool.Get()
	if err := con.Err(); err != nil {
		return nil, err
	}
	defer con.Close()
	parmas := make([]interface{}, 0)
	parmas = append(parmas, key)

	if len(args) > 0 {
		for _, v := range args {
			parmas = append(parmas, v)
		}
	}
	reply, err := con.Do(cmd, parmas...)
	return reply, err
}
func (r *redis) Set(key string, value interface{}) (string, error) {
	reply, err := r.Exec("set", key, value)
	if nil == reply {
		return "NO", err
	}
	return reply.(string), err
}
func (r *redis) Get(key string) (string, error) {
	reply, err := r.Exec("get", key)
	if nil == reply {
		return "", err
	}
	fmt.Printf("type: %T\n", reply)
	fmt.Printf("str: %s\n", reply)

	return string(reply.([]byte)), err
}
func (r *redis) Hset(key, field string, value interface{}) (int64, error) {
	reply, err := r.Exec("hset", key, field, value)
	if nil == reply {
		return -1, err
	}
	fmt.Printf("type: %T\n", reply)
	//fmt.Printf("str: %s\n", reply)
	return reply.(int64), err
}
func (r *redis) Hget(key, field string) (string, error) {
	reply, err := r.Exec("hget", key, field)
	if nil == reply {
		return "", err
	}
	fmt.Printf("type: %T\n", reply)
	return string(reply.([]byte)), err
}
