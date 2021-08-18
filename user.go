package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

//新建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}

func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))

	}
}

//用户上线提醒
func (this *User) Online() {
	//map线程不安全，要加锁
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已上线")
}

//用户下线提醒
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "下线")
}

func (this *User) SendMeg(msg string) {
	this.conn.Write([]byte(msg))
}

//用户发消息广播
func (this *User) DoMeg(msg string) {
	if msg == "who" {
		//
		//查询当前在线用户有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlinemsg := "[" + user.Addr + "]" + user.Name + "在线.."
			this.SendMeg(onlinemsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//
		//当用户输入rename|张三时 更换名字
		newName := strings.Split(msg, "|")[1]

		_, ok := this.server.OnlineMap[newName]
		//判断当前是否存在一样的名字name
		if ok {
			this.SendMeg("名字已经存在")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMeg("你已经更新名字" + this.Name)

		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		//获取对方用户名
		toName := strings.Split(msg, "|")[1]
		if toName==""{
			this.SendMeg("格式错误，请重新输入格式为”to|张三|消息“")
			return
		}
		//根据用户名，得到对方user对象
		toUser, ok := this.server.OnlineMap[toName]
		if !ok {
			this.SendMeg("该用户不存在")
			return
		}
		//获取消息内容，通过user发送过去
		toMsg := strings.Split(msg,"|")[2]
		if toMsg==""{
			this.SendMeg("无消息，请重发")
			return
		}
		toUser.SendMeg(this.Name + "对你说："+toMsg)

	} else {
		this.server.BroadCast(this, msg)
	}

}
