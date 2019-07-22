package main

import (
	"fmt"
	"io"
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
	// 先读取 客户端的数据 ， 再会写 ping...ping...ping \n
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))
	// 回写数据
	err := request.GetConnection().SendMsg(0, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
		fmt.Println(err)
	}
}

type HelloZinxRouter struct {
	znet.BaseRouter
}

func (r *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call HelloZinxRouter Handle")
	// 先读取客户端的数据， 再会写 ping ping ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))
	err := request.GetConnection().SendMsg(1, []byte("Hello Zinx Router V0.6"))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	// 服务端
	go func() {
		// 1 创建一个 server 句柄
		s := znet.NewServer()
		s.AddRouter(0, &PingRouter{}) // 添加路由
		s.AddRouter(1, &HelloZinxRouter{})
		// 2 开启服务
		s.Serve()
	}()
	go client(0, []byte("Zinx V0.6 Client0 Test Message"))
	go client(1, []byte("Zinx V0.6 Client1 Test Message"))
	for {
		time.Sleep(10 * time.Second)
	}
}

// 客户端
func client(id uint32, data []byte) {
	fmt.Println("Client ", id, " Test ... start")
	// 3秒之后发起测试请求， 给服务端 开启服务机会
	time.Sleep(3 * time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client start err exit!")
		return
	}
	for {
		// 发封包message消息
		dp := znet.NewDataPack()
		msg, err := dp.Pack(znet.NewMsgPackage(id, data))
		if err != nil {
			fmt.Println("client 封包失败")
			continue
		}
		if _, err := conn.Write(msg); err != nil {
			fmt.Println("write error err ", err)
			return
		}
		// 先读出 流中的 head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData) // ReadFull 把 msg 填充满为止
		if err != nil {
			fmt.Println("read head error")
			break
		}
		// 将 headData 字节流 拆包 到 msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("client unpack err: ", err)
			return
		}
		if msgHead.GetDataLen() > 0 {
			// msg 是有 data数据的， 需要再次读取data 数据
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetDataLen())
			// 根据 dataLen 从 io 中 读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err: ", err)
				return
			}
			fmt.Println("===> Recv Msg: ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}
		time.Sleep(1 * time.Second)
	}
}
