package znet

import (
	"fmt"
	"strconv"
	"zee.com/work/learn_zinx/utils"
	"zee.com/work/learn_zinx/ziface"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter // 存放 每个 MsgId 所对应的 处理方法的 map 属性
	WorkerPoolSize uint32                    // 业务工作Worker 池的数量
	TaskQueue      []chan ziface.IRequest    // Worker 负责去任务的消息队列
}

func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId =", request.GetMsgID(), " is not FOUND")
		return
	}
	// 执行对应处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// 为 消息 添加 具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	// 1 判断当前msg 绑定的API 处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	// 2 添加msg与api 的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize), // 一个Worker 对应一个 queue
	}
}
