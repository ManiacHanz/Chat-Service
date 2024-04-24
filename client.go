// cli工具
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		// 默认值
		flag: 999,
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

// 处理server的回应消息
// 直接显示标准输出
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 改名模式")
	fmt.Println("0. 退出")
	// 这里会阻塞
	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>> 请输入合法范围内的数字<<<<<<<<")
		return false
	}

}

// 更新用户名
func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>> 请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write.err: ", err)
		return false
	}

	return true
}

// 公聊
func (client *Client) PublicChat() {
	// 提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>> 请输入聊天内容， exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送给服务器

		// 消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println("请输入聊天内容。。。")
		fmt.Scanln(&chatMsg)
	}
}

// 查询在线用户
func (client *Client) ShowOnlineUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("select user err:", err)
		return
	}
}

// 私聊
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.ShowOnlineUsers()
	fmt.Println(" >>>> 请输入聊天对象[用户名]，exit退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>> 请输入消息内容，exit退出：")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			// 消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println("请输入聊天内容。。。")
			fmt.Scanln(&chatMsg)
		}

		client.ShowOnlineUsers()
		fmt.Println(" >>>> 请输入聊天对象[用户名]，exit退出")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}
		/// 根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			// 公聊模式
			client.PublicChat()
			break
		case 2:
			// 私聊模式
			client.PrivateChat()
			break
		case 3:
			// 更新用户名
			client.UpdateName()
			break
		}
	}
}

var ip string
var port int

// 在main之前执行
func init() {
	flag.StringVar(&ip, "ip", "127.0.0.1", "设置ip为127.0。0.1")
	flag.IntVar(&port, "port", 8888, "设置ip为127.0。0.1")
}

func main() {
	// 命令行解析参数
	flag.Parse()

	client := NewClient(ip, port)
	if client == nil {
		fmt.Println(">>>>>> 链接服务器失败...")
		return
	}

	fmt.Println(">>>>> 链接服务器成功....")

	//单独开一个goroutine来处理服务器输出
	go client.DealResponse()
	// 启动客户端业务
	client.Run()
}
