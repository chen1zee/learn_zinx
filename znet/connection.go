package znet

import (
	"fmt"
	"net"
	"zee.com/work/learn_zinx/ziface"
)

type Connection struct {
	// 当前连接的 socket TCP 套接字
	Conn *net.TCPConn
	// 当前连接的ID 也可以 称作为 SessionID ID全局唯一
	ConnID uint32
	// 当前连接的 关闭 状态  初始化为 false
	isClosed bool
	// 该链接的处理方法 router
	Router ziface.IRouter
	// 告知 该连接 已经退出/停止的 channel
	ExitBuffChan chan bool
}

// 创建连接的方法
func NewConnection(conn *net.TCPConn, connID uint32, router ziface.IRouter) *Connection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		Router:       router,
		ExitBuffChan: make(chan bool, 1),
	}
	return c
}

/** 处理conn 读取数据的 goroutine */
func (c *Connection) StartReader() {
	fmt.Println("Reader goroutine is running")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit!")
	defer c.Stop()
	for {
		// 读取我们最大的数据到 buf 中
		buf := make([]byte, 512)
		_, err := c.Conn.Read(buf)
		if err != nil {
			fmt.Println("接收数据err  ", err)
			c.ExitBuffChan <- true
			continue
		}
		// 得到当前客户端请求的 Request 数据
		req := Request{
			conn: c,
			data: buf,
		}
		// 从路由 Routers 中找到注册绑定Conn 的 对应 Handle
		go func(request ziface.IRequest) {
			c.Router.PreHandle(request)
			c.Router.Handle(request)
			c.Router.PostHandle(request)
		}(&req)
	}
}

// 启动连接 让 当前连接开始工作
func (c *Connection) Start() {
	// 开启处理该连接读取到客户端数据之后的 请求业务
	go c.StartReader()
	for {
		select {
		case <-c.ExitBuffChan:
			// 得到退出消息， 不在阻塞
			return
		}
	}
}

// 停止 连接， 结束当前连接状态M
func (c *Connection) Stop() {
	// 1. 如果当前连接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true
	// TODO Connection Stop() 如果用户注册了该链接的关闭回调业务， 那么此刻应该显示调用

	// 关闭socket链接
	_ = c.Conn.Close()
	// 通知 从缓冲队列读取数据的业务 , 该链接已经关闭
	c.ExitBuffChan <- true
	// 关闭该链接全部通道
	close(c.ExitBuffChan)
}

// 从当前连接获取原始的 socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
