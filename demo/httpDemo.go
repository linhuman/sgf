package main

import "github.com/linhuman/sgf"
func main(){
	sgf.Initialize(SgfConfig)
	sgf.Run("127.0.0.1:9009")
}
