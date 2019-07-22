package znet

import "zee.com/work/learn_zinx/ziface"

// 实现router时， 先嵌入这个基类， 然后根据需要对这个基类的方法进行重写
type BaseRouter struct{}

// 这里之所以BaseRouter方法为空
// 是因为 有的Route 不希望有PreHandle 或者 PostHandle
// 所以Router 全部继承 BaseRouter 的好处是， 不需要实现 PreHandle PostHandle 也可以实例化

func (br *BaseRouter) PreHandle(request ziface.IRequest) {}

func (br *BaseRouter) Handle(request ziface.IRequest) {}

func (br *BaseRouter) PostHandle(request ziface.IRequest) {}
