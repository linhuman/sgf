package controller

type Test struct{
	Base
}

func (t *Test) Hello(){
	t.Ctx.Response.Write("hello world")
}