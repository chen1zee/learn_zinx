package znet

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
	"zee.com/work/learn_zinx/utils"
	"zee.com/work/learn_zinx/ziface"
)

// 封包拆包类实例， 暂时不需要成员
type DataPack struct{}

// 封包拆包实例初始化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

// 获取包头部 长度方法
func (dp *DataPack) GetHeadLen() uint32 {
	// Id uint32(4字节) + DataLen uint32(4字节)
	return 8
}

// 封包方法(压缩数据)
func (*DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	// 创建一个存放 bytes 字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})
	// 写 dataLen
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}
	// 写msgID
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}
	// 写data数据
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}
	return dataBuff.Bytes(), nil
}

// 拆包方法(解压数据)
func (dp *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	// 创建一个从输入二进制数据的 ioReader
	dataBuff := bytes.NewReader(binaryData)
	// 只解压head的信息， 得到 dataLen 和 msgID
	msg := &Message{}
	// 读dataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}
	// 读取msgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}
	// 判断dataLen 的长度是否超出我们允许的最大包长度
	if utils.GlobalObject.MaxPacketSize > 0 && msg.DataLen > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("Too large msg data received")
	}
	// 这里只需要把 head数据拆包出来就可以， 然后再通过head 的长度 ， 再从 conn 读取一次数据
	return msg, nil
}
