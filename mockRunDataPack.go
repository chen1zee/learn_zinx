package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zee.com/work/learn_zinx/znet"
)

// 只负责 测试 dataPack 拆包 封包 功能
func server() {
	// 创建socket TCP Server
	listener, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("server listen err: ", err)
		return
	}
	// 创建服务器 goroutine, 负责从客户端 goroutine 读取粘包的数据 然后进行解析
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("server accept err: ", err)
			continue
		}
		// 处理 客户端请求
		go func(conn net.Conn) {
			// 创建封包拆包对象 dp
			dp := znet.NewDataPack()
			for {
				// 1 先读出流中的head部分
				headData := make([]byte, dp.GetHeadLen())
				_, err := io.ReadFull(conn, headData) // ReadFull 把 msg 填满为止
				if err != nil {
					fmt.Println("read head error")
					break
				}
				// 将headData 字节流 拆包到 msg中
				msgHead, err := dp.Unpack(headData)
				if err != nil {
					fmt.Println("server unpack err: ", err)
					return
				}
				if msgHead.GetDataLen() > 0 {
					// msg 是有 data 数据的，需要再次读取data 数据
					msg := msgHead.(*znet.Message) // interface 类型 转型
					msg.Data = make([]byte, msg.GetDataLen())
					// 根据dataLen 从 io 中读取字节流
					_, err := io.ReadFull(conn, msg.Data)
					if err != nil {
						fmt.Println("server unpack data err: ", err)
						return
					}
					fmt.Println("===> Recv Msg: ID=", msg.Id, " len=", msg.DataLen, " data=", string(msg.Data))
				}
			}
		}(conn)
	}
}

func client() {
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client dial err: ", err)
		return
	}
	// 创建一个封包对象 dp
	dp := znet.NewDataPack()
	// 封装一个msg1 包
	msg1 := &znet.Message{
		Id:      0,
		DataLen: 5,
		Data:    []byte{'h', 'e', 'l', 'l', 'o'},
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack msg1 err: ", err)
		return
	}
	msg2 := &znet.Message{
		Id:      1,
		DataLen: 7,
		Data:    []byte{'w', 'o', 'r', 'l', 'd', '!', '!'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client temp msg2 err: ", err)
		return
	}
	// 将sendData1 和sendData2 拼接在一起 组成 粘包
	sendData1 = append(sendData1, sendData2...)
	// 向服务器写数据
	_, _ = conn.Write(sendData1)
	// 客户端阻塞
	//for {
	//	time.Sleep(10 * time.Second)
	//}
}

func main() {
	go server()
	go client()

	for {
		time.Sleep(10 * time.Second)
	}
}
