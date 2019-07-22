package znet

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
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
		// 创建拆包 解包 对象
		dp := NewDataPack()
		// 读取客户端的 Msg head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			c.ExitBuffChan <- true
			continue
		}
		// 拆包， 得到 msgId dataLen 放到 msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.ExitBuffChan <- true
			continue
		}
		// 根据 dataLen 读取 data, 放在 msg.Data 中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.ExitBuffChan <- true
				continue
			}
		}
		msg.SetData(data)
		// 得到当前客户端请求的 Request 数据
		req := Request{
			conn: c,
			msg:  msg,
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

// 直接将 Message 数据 发送数据给 远程的 TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	// 将 data 封包 ，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg")
	}
	// 写回客户端
	if _, err := c.Conn.Write(msg); err != nil {
		fmt.Println("Write msg id ", msgId, " error ")
		c.ExitBuffChan <- true
		return errors.New("conn Write error")
	}
	return nil
}
