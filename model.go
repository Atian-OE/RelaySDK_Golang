package relaysdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type MsgID int32

const (
	ConnectID    MsgID = 0   //连接
	DisconnectID MsgID = 1   //关闭连接
	Open         MsgID = 2   //打开继电器
	Close        MsgID = 3   //关闭继电器
	Reset        MsgID = 4   //重置继电器
	HeartBeatID  MsgID = 250 //心跳
	Illegal      MsgID = 255 //非法请求
)

var (
	SendErr = func(err error) error {
		return errors.New(fmt.Sprintf("客户端发送消息失败:%s", err.Error()))
	}
)

//打开继电器请求参数
type OpenMessageRequest struct {
	Relay []bool
}

//打开继电器响应参数
type OpenMessage struct {
	Success bool
	Err     string
}

//关闭继电器请求参数
type CloseMessageRequest struct {
	Relay []bool
}

//关闭继电器响应参数
type CloseMessage struct {
	Success bool
	Err     string
}

//重置继电器请求参数
type ResetMessageRequest struct {
}

//重置继电器响应参数
type ResetMessage struct {
	Success bool
	Err     string
}

//心跳
type HeartBeat struct {
}

type DisconnectedMessage struct {
	Success bool
	Err     string
}

type IllegalMessage struct {
	Success bool
	Err     string
}

func Encode(msgObj interface{}) ([]byte, error) {
	data, err := json.Marshal(msgObj)
	if err != nil {
		log.Println("结构体换成成字节类型失败", err)
		return nil, err
	}
	cache := make([]byte, len(data)+5)
	length, err := IntToBytes(int64(len(data)), 4)
	if err != nil {
		log.Println("整型转换成字节类型失败", err)
		return nil, err
	}
	copy(cache, length)
	switch msgObj.(type) {
	case *OpenMessage, *OpenMessageRequest:
		cache[4] = byte(Open)
	case *CloseMessage, *CloseMessageRequest:
		cache[4] = byte(Close)
	case *ResetMessage, *ResetMessageRequest:
		cache[4] = byte(Reset)
	case *HeartBeat:
		cache[4] = byte(HeartBeatID)
	default:
		cache[4] = byte(Illegal)
	}
	copy(cache[5:], data)
	return cache, err
}
