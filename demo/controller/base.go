package controller

import "github.com/linhuman/sgf/action"

type Base struct {
	action.Base
}
func (b *Base) Before(){
	b.Ctx.Response.Write("before")
}
func (b *Base) After(){
	b.Ctx.Response.Write("after")
}