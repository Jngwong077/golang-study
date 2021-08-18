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

	OnlineMap map[string]*User
	mapLock   sync.Mutex

	Message chan string
}

//Server 接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,

		OnlineMap: make(map[string]*User),

		Message:   make(chan string),
	}
	return server
}

//广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {

	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

//监听Message广播消息channel的goroutine，一旦有消息就发送给全部在线的User
func (this *Server) Lm() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}

		this.mapLock.Unlock()
	}
}

//当前链接的业务
func (this *Server) Handler(conn net.Conn) {

	user := NewUser(conn,this)

	user.Online()

	//监听用户是否活跃的channnel
	isLive := make(chan bool)
	//接受客服端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Printf("Conn Read err: ", err)
				return
			}

			//提取用户的消息，去除“/n”
			msg := string(buf[:n-1])

			user.DoMeg(msg)
			//只要用户有消息，就是活跃
			isLive<-true
		}
	}()
	for{
		select {
		case <-isLive:
			//当前用户活跃，重置定时器
			//不需要做任何东西，为了触发下面的定时器
		case <-time.After(time.Second *10):
			//已经超10秒，将当前的User强制关闭

			close(user.C) //销毁通道

			conn.Close()//关闭连接
			//退出当前的Handler
			return // runtime.Goexit()
		}
	}


}

//启动服务器接口
func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listen err:", err)
		return
	}
	defer listener.Close()

	go this.Lm()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		go this.Handler(conn)

	}
}
