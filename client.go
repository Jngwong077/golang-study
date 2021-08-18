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
	//创建对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       99,
	}
	//链接Server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}

	client.conn = conn

	return client
}

func (client *Client) menu() bool {
	var f int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	//Scanln从标准输入扫描文本，将成功读取的空白分隔的值保存进成功传递给f。在换行时停止扫描
	fmt.Scanln(&f)

	if f >= 0 && f <= 3 {
		client.flag = f
		return true
	} else {
		fmt.Println("请输入合法范围")
		return false
	}

}

var serverIp string
var serverPort int

//StringVar用指定的名称、默认值、使用信息注册一个string类型flag，并将flag的值保存到指向的变量。
//什么类型对于着使用什么Var
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置IP（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置端口号（默认是8888）")

}

func (client *Client) UpdateName() bool {

	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"

	//将消息写到服务器中
	_, err := client.conn.Write([]byte(sendMsg))

	if err != nil {
		fmt.Println("client err :", err)
		return false
	}
	return true
}

func (client *Client) PublicChat() {
	var chatMsg string
	//Scanln从标准输入扫描文本，将成功读取的空白分隔的值保存进成功传递给charMsg。在换行时停止扫描
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("client err :", err)
				break
			}

			chatMsg = ""
			fmt.Scanln(&chatMsg)
		}
	}

}

func (client *Client) selectUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client err:", err)
		return
	}

}
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.selectUser()
	fmt.Println("输入聊天对象，exit表示退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("输出消息，exit表示退出")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("client err :", err)
					break
				}

				chatMsg = ""
				fmt.Println("请输入聊天信息，exit表示退出")
				fmt.Scanln(&chatMsg)
			}
		}

		client.selectUser()
		fmt.Println("输入聊天对象，exit表示退出")
		fmt.Scanln(&remoteName)

	}

}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}

		switch client.flag {
		case 1:
			client.PublicChat()
		case 2:
			client.PrivateChat()
		case 3:
			client.UpdateName()


		}

	}
}

//处理server回应的消息，直接显示到标准输出中
func (client *Client) DealResponse() {
	//将client.conn的数据拷贝到os.Stdout，直到在client.conn上到达EOF或发生错误。返回拷贝的字节数和遇到的第一个错误。
	//
	//一旦client.conn有数据，就直接copy到sStdout标准输出上，永久阻塞监听
	//os.Stdout指向系统的标准输出，它实际上是一个文件指针
	io.Copy(os.Stdout, client.conn)
}

func main() {
	//命令行解析注册的flag
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("连接服务器失败")
		return
	}
	fmt.Println("链接成功")

	go client.DealResponse()

	//启动客户端业务
	client.Run()
}
