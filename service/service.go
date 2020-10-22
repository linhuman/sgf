package service

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type service struct {
	task_num int
	task     Task
}
type Task interface {
	Work(int)
	Stop()
}

func New() *service {
	return &service{}
}
func (s *service) RegisterTask(task Task) {
	s.task = task

}

func (s *service) handle_signal() {
	ch := make(chan os.Signal, 1)
	// 监听信号
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM: // 终止进程执行
			signal.Stop(ch)
			s.task.Stop()
			log.Println("service shutdown...")
			return
		}
	}
}

func (s *service) SetTaskNum(num int) {
	s.task_num = num
}
func (s *service) Run() {
	var wg sync.WaitGroup

	if 0 >= s.task_num {
		panic("任务数不能为0")
	}
	for i := 0; i < s.task_num; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			s.task.Work(index)

		}(i)
	}
	s.handle_signal()
	wg.Wait()
}
