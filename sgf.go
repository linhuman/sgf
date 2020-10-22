package sgf

import (
	"sgf/config"
	"sync"
)

var once sync.Once

func Initialize(custom_cfg config.Cfg) {
	once.Do(func() {
		config.Entity = custom_cfg
	})
}
