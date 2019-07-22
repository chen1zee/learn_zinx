package ziface

/*
	iRequest 接口:
	实际上是把客户端请求的连接信息 和 请求的数据 包装到 Request 里
*/
type IRequest interface {
	GetConnection() IConnection // 获取请求链接信息
	GetData() []byte            // 获取请求消息的数据
	GetMsgID() uint32           // 获取 msgId
}
