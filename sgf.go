package sgf

import (
	"context"
	"fmt"
	"github.com/linhuman/sgf/common"
	"github.com/linhuman/sgf/config"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
)

var once sync.Once

func Initialize(customCfg config.Cfg) {
	once.Do(func() {
		config.Entity = customCfg
		routersLen := len(config.Entity.Routers)
		for i := 0; i < routersLen; i++ {
			router := &config.Entity.Routers[i]
			http.HandleFunc((*router)[0].(string), func(ResponseWriter http.ResponseWriter, request *http.Request) {
				start := time.Now()
				ctx := new(Context)
				ctx.Response.HttpResp = ResponseWriter
				ctx.Request = request
				methodName := (*router)[2].(string)
				c := reflect.New(reflect.TypeOf((*router)[1]))
				ctx.Action = c.Type().Elem().String() + "." + methodName
				defer finished(ctx, &c)
				valCtx := reflect.ValueOf(ctx)
				c.Elem().FieldByName("Ctx").Set(valCtx)
				in := make([]reflect.Value, 0)
				//value不能访问value指针的成员函数，value指针能访问value成员函数和value指针成员函数
				c.MethodByName("Before").Call(in)
				c.MethodByName(methodName).Call(in)
				end := time.Now()
				//c的type如果是指针类型，不能调用FieldByName，需要先调用Elem()
				c.Elem().FieldByName("Ctx").Elem().FieldByName("Latency").Set(reflect.ValueOf(end.Sub(start).Microseconds()))
			})
		}
	})
}

func finished(ctx *Context, c *reflect.Value) {
	msg := recover()
	if nil != msg {
		strMsg := fmt.Sprintf("%s", msg)
		pos := strings.Index(strMsg, "200")
		if 0 <= pos {
			ctx.Response.HttpResp.Write([]byte(strMsg[pos+3:]))
		} else {
			common.WriteLog("程序异常", []interface{}{strMsg, string(debug.Stack())}, ctx.Action+"_error")
		}
	}
	c.MethodByName("After").Call(make([]reflect.Value, 0))
}
func listenSignal(ctx context.Context, httpSrv *http.Server) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-sigs:
		fmt.Println("Http Server Shutdown...")
		httpSrv.Shutdown(ctx)
	}
}
func Run(addr string) {
    server := http.Server{Addr : addr, Handler : nil}
	go server.ListenAndServe()
	listenSignal(context.Background(), &server)
}
