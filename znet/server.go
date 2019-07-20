package znet

import (
	"fmt"
	"github.com/pkg/errors"
	"net"
	"time"
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
}

// ============= 定义当前客户端链接的 handle api
func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
	// 回显业务
	fmt.Println("[Conn Handle] CallBackToClient ... ")
	if _, err := conn.Write(data[:cnt]); err != nil {
		fmt.Println("write back buf err ", err)
		return errors.New("放回客户端 处理函数 失败 CallBackToClient error \n")
	}
	return nil
}

// ============= 实现 ziface.IServer 里的全部 接口方法

// 开启网络服务
func (s *Server) Start() {
	fmt.Printf("[START] Server listener at IP: %s, Port %d, is starting \n", s.IP, s.Port)
	// 开启一个go 做服务端 Listener 业务
	go func() {
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
			// 3.2 TODO Server.Start 设置服务器最大连接控制， 如果超过最大连接， 则关闭 此新的连接
			// 3.3 TODO Server.Start 处理该新连接请求的 业务方法， 此时应有 handler 和 conn 是绑定的

			dealConn := NewConnection(conn, cid, CallBackToClient)
			cid++
			// 3.4 启动当前连接的处理业务
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server, name ", s.Name)
	// TODO Server.Stop() 将其他需要清理的连接信息 或其他信息 也一并 停止或者清理
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
func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:      name,
		IPVersion: "tcp4",
		IP:        "0.0.0.0",
		Port:      7777,
	}

	return s
}