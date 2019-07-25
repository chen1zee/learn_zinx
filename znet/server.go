package znet

import (
	"fmt"
	"net"
	"time"
	"zee.com/work/learn_zinx/utils"
	"zee.com/work/learn_zinx/ziface"
)

// iServer 接口实现， 定义一个 Server 服务类
type Server struct {
	// 服务器的方法
	Name string
	// tcp4 or other
	IPVersion string
	// 服务绑定的IP地址
	IP string
	// 服务绑定端口
	Port int
	// 当前Server的消息管理模块， 用来绑定MsgId和对应的处理方法
	msgHandler ziface.IMsgHandle
	// 当前Server 的连接管理器
	ConnMgr ziface.IConnManager
	// ==================
	// 新增两个hook函数原型
	// 该Server 的连接创建时的Hook函数
	OnConnStart func(conn ziface.IConnection)
	// 该Server 连接断开时 Hook函数
	OnConnStop func(conn ziface.IConnection)
}

// 设置该 Server 的连接创建Hook函数
func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 设置该 Server 的连接断开Hook函数
func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(connection ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallConStart ...")
		s.OnConnStart(connection)
	}
}

// 调用连接onConnStop Hook函数
func (s *Server) CallOnConnStop(connection ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> CallOnConnStop ...")
		s.OnConnStop(connection)
	}
}

// 得到链接管理
func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

// ============= 实现 ziface.IServer 里的全部 接口方法

// 添加 路由函数
func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
}

// 开启网络服务
func (s *Server) Start() {
	fmt.Printf("[START] Server name: %s, listener at IP: %s, Port %d, is starting \n", s.Name, s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d, MaxPacketSize: %d \n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)
	// 开启一个go 做服务端 Listener 业务
	go func() {
		// 0 启动worker工作池机制
		s.msgHandler.StartWorkerPool()
		// 1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err: ", err)
			return
		}
		// 2 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, s.IP, ":", s.Port, " err ", err)
			return
		}
		// 监听成功
		fmt.Println("start Zinx server  ", s.Name, " success, now listening...")
		// TODO server.go 应该有一个 自动生成 ID 的方法
		var cid uint32
		cid = 0
		// 3 启动 server 网络连接业务
		for {
			// 3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}
			// 3.2 设置服务器最大连接控制，如果超过最大连接， name则关闭此新链接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				_ = conn.Close()
				continue
			}
			// 3.3 Server.Start 处理该新连接请求的 业务方法， 此时应有 handler 和 conn 是绑定的
			dealConn := NewConnection(s, conn, cid, s.msgHandler)
			cid++
			// 3.4 启动当前连接的处理业务
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server, name ", s.Name)
	// TODO Server.Stop() 将其他需要清理的连接信息 或其他信息 也一并 停止或者清理
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()
	// TODO Server.Serve 是否在启动服务的时候， 还要处理其他事情， 可在此添加

	// 阻塞 否则主 Go 退出  listener 的 go 将退出
	for {
		time.Sleep(10 * time.Second)
	}
}

// 创建一个服务器句柄
func NewServer() ziface.IServer {
	// 先初始化 全局配置文件
	utils.GlobalObject.Reload()
	s := &Server{
		Name:       utils.GlobalObject.Name, // 从全局参数获取
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TcpPort,
		msgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(), // 创建ConnManager
	}
	return s
}
