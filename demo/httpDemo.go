package main

import (
	"demo/config"
	"github.com/linhuman/sgf"
)

func main(){
	sgf.Initialize(config.SgfConfig)
	sgf.Run("127.0.0.1:9009")
}
