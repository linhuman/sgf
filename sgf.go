package sgf

import (
	"sync"

	"github.com/linhuman/sgf/config"
)

var once sync.Once

func Initialize(custom_cfg config.Cfg) {
	once.Do(func() {
		config.Entity = custom_cfg
	})
}
