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

// 启动一个Worker 工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started")
	// 不断的等待队列中的消息
	for {
		select {
		// 有消息则取出队列的 Request , 并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// 启动worker 工作池
func (mh *MsgHandle) StartWorkerPool() {
	// 遍历需要启动worker的数量，依次启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个worker 被启动
		// 给当前worker 对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		// 启动当前Worker,阻塞等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 将消息交给TaskQueue, 由 worker 进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	// 根据ConnID 来分配当前的连接应该由哪个worker负责处理
	// 轮询的平均分配法则

	// 得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnID=", request.GetConnection().GetConnID(), " to workerID=", workerID)
	// 将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
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
