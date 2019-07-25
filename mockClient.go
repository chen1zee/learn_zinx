package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zee.com/work/learn_zinx/znet"
)

func main() {
	fmt.Println("Client ", 0, " Test ... start")
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
		msg, err := dp.Pack(znet.NewMsgPackage(0, []byte("Zinx V1 Client0 Test Message")))
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
