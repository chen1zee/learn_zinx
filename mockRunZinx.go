package main

import (
	"fmt"
	"net"
	"time"
	"zee.com/work/learn_zinx/ziface"
	"zee.com/work/learn_zinx/znet"
)

// ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

func (r *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("ping...ping...ping \n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

func main() {
	// 服务端
	go func() {
		// 1 创建一个 server 句柄
		s := znet.NewServer()
		s.AddRouter(&PingRouter{}) // 添加路由
		// 2 开启服务
		s.Serve()
	}()
	// 客户端
	fmt.Println("Client Test ... start")
	// 3秒之后发起测试请求， 给服务端 开启服务机会
	time.Sleep(3 * time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client start err exit!")
		return
	}
	for {
		_, err := conn.Write([]byte("abc"))
		if err != nil {
			fmt.Println("write error err ", err)
			return
		}
		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read buf error ")
			return
		}
		fmt.Printf(" server call back : %s, cnt = %d \n", buf, cnt)
		time.Sleep(1 * time.Second)
	}
}
