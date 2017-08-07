package main

type message struct {
	FromUserId string      //消息发送者uid
	ToUserId   string      //被发送人uid， 若ToUserId=""或ToUserId="0", 表示广播
	Type       int8        //消息类型
	Message    interface{} //消息体
	When       int64       //消息发送时间， 时间戳
}
