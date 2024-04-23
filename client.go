package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}

	// 链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}
	client.conn = conn

	// 返回对象
	return client
}

var ip string
var port int

// 在main之前执行
func init() {
	flag.StringVar(&ip, "ip", "127.0.0.1", "设置ip为127.0。0.1")
	flag.IntVar(&port, "port", 8888, "设置ip为127.0。0.1")
}

func main() {
	flag.Parse()

	client := NewClient(ip, port)
	if client == nil {
		fmt.Println(">>>>>> 链接服务器失败...")
		return
	}

	fmt.Println(">>>>> 链接服务器成功....")

	// 启动客户端业务
	select {}
}
