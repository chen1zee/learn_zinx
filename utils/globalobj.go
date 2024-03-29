package utils

import (
	"encoding/json"
	"io/ioutil"
	"zee.com/work/learn_zinx/ziface"
)

/*
	存储一切有关Zinx框架的全局参数，供其他模块使用
	一些参数可以通过 用户更具 zinx.json来配置
*/
type GlobalObj struct {
	/*
		Server
	*/
	TcpServer ziface.IServer // 当前Zinx的全局Server 对象
	Host      string         // 当前服务器 主机IP
	TcpPort   int            // 当前服务器主机监听端口号
	Name      string         // 当前服务器名称
	/*
		Zinx
	*/
	Version          string // 当前Zinx版本号
	MaxPacketSize    uint32 // 都需数据包的最大值
	MaxConn          int    // 当前服务器主机允许的最大连接个数
	WorkerPoolSize   uint32 // 业务工作Worker 池的 数量
	MaxWorkerTaskLen uint32 // 业务工作Worker 对应负责的任务队列最大任务存储数量
	MaxMsgChanLen    uint32 // 业务消息 chan 长度
	/*
		config file path
	*/
	ConfFilePath string
}

func (g *GlobalObj) Reload() {
	data, err := ioutil.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}
	// 将json数据解析到 struct 中
	// fmt.Printf("json: %s \n", data)
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

/*
	提供init方法 默认加载
*/
func init() {
	// 初始化 GlobalObject 变量 设置一些默认值
	GlobalObject = &GlobalObj{
		Name:             "ZinxServerApp",
		Version:          "V0.4",
		TcpPort:          7777,
		Host:             "0.0.0.0",
		MaxConn:          12000,
		MaxPacketSize:    4096,
		ConfFilePath:     "conf/zinx.json",
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    10,
	}
	// 从配置文件中加载一些用户配置的参数
	GlobalObject.Reload()
}

/*
	定义一个全局的对象
*/
var GlobalObject *GlobalObj
