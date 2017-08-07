package main

import (
	"fmt"
	"log"
	_ "runtime"
	_ "strconv"
	"time"

	"github.com/lfq618/deaf-socket/conv"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/stretchr/objx"
)

const (
	redisAdress = "172.28.50.229:6379"
	redisKey    = "messageList"
)

type client struct {
	socket  *websocket.Conn
	send    chan *message
	ucenter *ucenter
	userId  string
}

func (c *client) readFromSocket() {

	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now().Unix()
			//用户信息
			log.Println("收到消息：", msg)
			c.ucenter.messages <- msg
		} else {
			log.Println("[读取socket信息失败]\t", err.Error())
			break
		}
	}
}

func (c *client) readFromRedis() {
	//链接redis
	myredis, err := redis.Dial("tcp", redisAdress)
	if err != nil {
		log.Fatal("[redis] redis链接失败,", err.Error())
		return
	}

	defer myredis.Close()
	for {

		msgJson, err := redis.String(myredis.Do("LPOP", redisKey))
		if err != nil {
			//log.Println("[redis] 读取消息错误,", err.Error())
			time.Sleep(time.Millisecond * 10000)
			//runtime.Gosched()
			continue
		}
		msgMap, _ := objx.FromJSON(msgJson)
		log.Println("[redis] 消息体：", msgMap)

		var msg message

		if err = conv.Map2Obj(msgMap, &msg, nil); err != nil {
			log.Println("[map2strcut]失败,", err.Error())
			continue
		}

		log.Println("[map2struct]成功,", msg.Message)

		c.ucenter.messages <- &msg
	}
}

func (c *client) write() {
	for msg := range c.send {
		fmt.Println("发送消息：", msg)
		if err := c.socket.WriteJSON(msg); err != nil {
			fmt.Println("发送消息失败:", err.Error())
			break
		}
	}

	c.socket.Close()
}
