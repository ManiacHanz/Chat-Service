package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	// 有关异步流程（这里是channel)里加入同步机制就是sync锁
	mapLock sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听Message Channel， 然后发给其他用户自己的Channel
// 这里的流程是，用户上线之后在Server端发送给Server端的Message Channel
// 然后监听Message Channel，有的话就发送给所有用户自己的channel
// 然后用户自己在进行监听，并从里面取消息
func (this *Server) ListenMessage() {
	// 外层相当于while循环
	for {
		msg := <-this.Message
		fmt.Println("ListenMessage: ", msg)
		this.mapLock.Lock()
		// 遍历所有在线用户，然后发送消息到管道
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()

	}

}

// 广播消息的方法
func (this *Server) Broadcast(user *User, msg string) {
	message := "[" + user.Addr + "]" + user.Name + ":" + msg
	fmt.Println("sended: " + message)
	this.Message <- message
}

func (this *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	fmt.Println("链接建立成功1", conn.RemoteAddr().String())
	// 传入server实例，方便user实例调用
	user := NewUser(conn, this)

	// 用户上线
	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			// 断开连接的时候这个消息是0
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read Err: ", err)
				return
			}

			// 提取消息 这个是给nc链接处理
			// 去掉最后一位，相当于str(0, length-1)
			msg := string(buf[:n-1])

			//广播
			user.DoMessage(msg)

			// 用户发了任意消息，代表当前用户是一个活跃的，刷新一下isLive的消息
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
			// 这里收到isLive刷新了消息，代表用户活跃（发送了消息）
			// 做刷新定时器的操作
			// 这里不用写刷新定时的逻辑，是因为在golang中，如果这里不写逻辑，就会走到写一个case的条件语句中（注意：是条件语句）
			// 那么走到条件以后，调用time包，则表示刷新了定时器
		case <-time.After(time.Second * 1000):
			// 已经超时
			// 关闭当前User
			user.SendMsgToSelf("你被踢了\n")

			// 关闭连接
			conn.Close()

			// 从在线列表里剔除
			this.mapLock.Lock()
			delete(this.OnlineMap, user.Name)
			this.mapLock.Unlock()

			// 销毁用的资源  关闭channel
			close(user.C)
			close(isLive)

			// 退出当前的handler循环
			return
		}
	}
}

// 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	fmt.Println("Server starting...")

	// close listen socket
	defer listener.Close()

	// 监听Message Channel, 并执行后续操作
	go this.ListenMessage()

	for {

		// accept   表示有链接接入
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err: ", err)
			return
		}

		// do handler
		go this.Handler(conn)
	}

}
