package ziface

/*
	路由接口， 这里路由是 使用 框架者给该链接自定的 处理业务方法
	路由里的 IRequest 则包含用该链接的连接信息和该链接的请求数据信息
*/
type IRouter interface {
	PreHandle(request IRequest)  // 在处理conn 业务之前的钩子方法
	Handle(request IRequest)     // 处理conn业务方法
	PostHandle(request IRequest) // 处理conn业务之后的钩子方法
}
