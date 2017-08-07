package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type ucenter struct {
	//将要发送到其他client的消息channel
	messages chan *message

	//将要加入到client链接池中的客户端channel
	join chan *client

	//将要离开client链接池的客户端channel
	leave chan *client

	//socket 链接池, userId => socket套接字
	clients map[string]*client
}

//newUcenter makes a new ucenter that is ready to go
func newUcenter() *ucenter {
	return &ucenter{
		messages: make(chan *message),
		join:     make(chan *client),
		leave:    make(chan *client),
		clients:  make(map[string]*client),
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (u *ucenter) run() {

	for {
		select {
		case client := <-u.join:
			//新的socket链接
			log.Println("建立socket链接：", client.userId)
			u.clients[client.userId] = client
		case client := <-u.leave:
			//断开socket链接
			log.Println("关闭socket链接：userId=", client.userId)
			delete(u.clients, client.userId)
			close(client.send)
		case msg := <-u.messages:
			//有新的消息
			if msg.ToUserId != "0" && msg.ToUserId != "" {
				//发送给ToUserId
				//判断ToUserId是否有逗号分隔符
				for _, userId := range strings.Split(msg.ToUserId, ",") {
					msg.ToUserId = strings.Trim(userId, " ")
					if client, ok := u.clients[msg.ToUserId]; ok {
						log.Println("toUserId=", msg.ToUserId)
						client.send <- msg
					}
				}

			} else {
				//发送给所有人
				for userId, client := range u.clients {
					select {
					case client.send <- msg:
					default:
						delete(u.clients, userId)
						close(client.send)
					}
				}
			}
		}
	}
}

func (u *ucenter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var userId string
	var msg *message

	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err.Error())
		return
	}

	if err := socket.ReadJSON(&msg); err != nil {
		fmt.Println("ServeHTTP: 无法解析userId,", err.Error())
		return
	} else {
		userId = msg.FromUserId
	}

	fmt.Println("userId=", userId)
	client := &client{
		socket:  socket,
		send:    make(chan *message, messageBufferSize),
		ucenter: u,
		userId:  userId,
	}

	u.join <- client
	defer func() {
		u.leave <- client
	}()

	go client.write()
	//go client.readFromRedis()
	client.readFromSocket()

}
