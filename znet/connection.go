package znet

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"zee.com/work/learn_zinx/utils"
	"zee.com/work/learn_zinx/ziface"
)

type Connection struct {
	// 当前Conn属于哪个Server
	TcpServer ziface.IServer // 当前Conn属于哪个Server, 在 conn 初始化的时候添加即可
	// 当前连接的 socket TCP 套接字
	Conn *net.TCPConn
	// 当前连接的ID 也可以 称作为 SessionID ID全局唯一
	ConnID uint32
	// 当前连接的 关闭 状态  初始化为 false
	isClosed bool
	// 消息管理MsgId 和对应处理方法的 消息管理模块
	MsgHandler ziface.IMsgHandle
	// 告知 该连接 已经退出/停止的 channel
	ExitBuffChan chan bool
	// 无缓冲管道 用于 读写 两个goroutine 之间的 消息通信
	msgChan chan []byte
	// 有缓冲管道 用于 读 写 两个goroutine 之间的消息通信
	msgBuffChan chan []byte // 定义 channel成员
}

// 创建连接的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server, // 将隶属的server 传递进来
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte), // msgChan初始化
		msgBuffChan:  make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
	}
	// 将新创建的Conn 添加到链接管理中
	c.TcpServer.GetConnMgr().Add(c) // 将当前新创建的连接添加到 ConnManager 中
	return c
}

// 发送带缓冲信息
func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}
	// 将data封包,并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}
	// 写回客户端
	c.msgBuffChan <- msg
	return nil
}

/*
	写消息Goroutine, 用户将数据发送给客户端
*/
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")
	for {
		select {
		case data := <-c.msgChan:
			// 有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Second Data error:, ", err, " Conn Writer exit")
				return
			}
		// 针对有缓冲channel 需要写的数据处理
		case data, ok := <-c.msgBuffChan:
			if !ok {
				fmt.Println("msgBuffChan is Closed")
				break
			}
			// 有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Buff Data error: ", err, " Conn Writer exit")
				return
			}
		case <-c.ExitBuffChan:
			// conn 已经关闭
			return
		}
	}
}

/** 读消息Goroutine 用于从客户端中读取数据 */
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
		if utils.GlobalObject.WorkerPoolSize > 0 {
			// 已经启动工作池机制， 将消息交给Worker 处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			// 从绑定好的消息和对应的处理方法中 执行 对应的 Handle 方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

// 启动连接 让 当前连接开始工作
func (c *Connection) Start() {
	// 开启处理该连接读取到客户端数据之后的 请求业务
	go c.StartReader()
	go c.StartWriter()
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
	fmt.Println("Conn Stop()... ConnID = ", c.ConnID)
	c.isClosed = true
	// 关闭socket链接
	_ = c.Conn.Close()
	// 通知 从缓冲队列读取数据的业务 , 该链接已经关闭
	c.ExitBuffChan <- true
	// 将链接从链接管理器 中删除
	c.TcpServer.GetConnMgr().Remove(c) // 删除 conn 从ConnManager 中
	// 关闭该链接全部通道
	close(c.ExitBuffChan)
	close(c.msgBuffChan)
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
	c.msgChan <- msg // 将之前直接回写给 conn.Writer 的方法 改为 发送给 Channel 供 Writer 读取
	return nil
}
