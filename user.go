package main

import (
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string // 客户端的channel
	conn net.Conn

	//需要调用当前的服务实例
	server *Server
}

// 创建一个用户的api
func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name:   conn.RemoteAddr().String(), // 默认以客户端地址为用户名
		Addr:   conn.RemoteAddr().String(),
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 	// 启动监听
	go user.ListenMessage()

	return user
}

// 以下3个方法把服务端发送消息的功能放到类上

// 用户上线业务
func (this *User) Online() {
	// 用户上线，将用户加入onlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	// 广播当前用户上线消息
	this.server.Broadcast(this, "已上线")
}

// 用户下线业务
func (this *User) Offline() {
	this.server.Broadcast(this, "下线")
}

// 发送消息给当前用户
func (this *User) SendMsgToSelf(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	// 随便写一个变量。查询当前在线用户
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			message := "user [" + user.Name + "]在线" + "\n"
			this.SendMsgToSelf(message)
		}
		this.server.mapLock.Unlock()
	} else {
		this.server.Broadcast(this, msg)
	}
}

// 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
