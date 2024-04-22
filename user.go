package main

import (
	"net"
	"strings"
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
			message := "user [" + user.Addr + "]" + user.Name + ":在线" + "\n"
			this.SendMsgToSelf(message)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式: rename|zhangsan
		newName := strings.Split(msg, "|")[1]

		// 判断name是否㛮
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsgToSelf("当前用户名已被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsgToSelf("您已成功修改用户名为" + newName + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 私聊功能 使用 "to|userName|content"的格式
		// 1 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsgToSelf("消息格式不正确，请使用 \"to|张三|你好啊\"的格式。 \n")
			return
		}
		// 2根据用户名得到User对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsgToSelf("该用户不存在或不在线\n")
			return
		}
		// 3 获取消息内容，发送消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsgToSelf("请输入内容 \n")
			return
		}
		remoteUser.SendMsgToSelf(this.Name + "对您说：" + content)
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
