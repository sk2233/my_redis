package main

import (
	"fmt"
	"net"
)

type AsyncResp struct {
	Resp     *Resp
	WaitChan chan struct{}
}

func NewAsyncResp() *AsyncResp {
	return &AsyncResp{WaitChan: make(chan struct{}, 1)} // 设置一个
}

func (r *AsyncResp) Ready(resp *Resp) {
	r.Resp = resp
	r.WaitChan <- struct{}{} // 有一个空间，提前完成也不怕
}

func (r *AsyncResp) Load() *Resp {
	<-r.WaitChan // 需要等待 Ready 的调用
	return r.Resp
}

type Client struct {
	Conn     net.Conn
	Resps    map[string]*AsyncResp
	SendChan chan *Req
}

func (c *Client) readLoop() {
	for {
		resp := &Resp{}
		ReadObj(c.Conn, resp) // 异步读取数据
		if temp, ok := c.Resps[resp.SeqID]; ok {
			temp.Ready(resp) // 设置完毕并移除
			delete(c.Resps, resp.SeqID)
		} else {
			fmt.Println(resp.Cmd, resp.Args) // 没有匹配的
		}
	}
}

func (c *Client) Send(req *Req) *AsyncResp {
	resp := NewAsyncResp()
	c.Resps[req.SeqID] = resp
	c.SendChan <- req // 异步完成
	return resp
}

func (c *Client) writeLoop() {
	for req := range c.SendChan {
		WriteObj(c.Conn, req) // 有任务就发一下
	}
}

func NewClient(addr string) *Client {
	conn, err := net.Dial("tcp", addr)
	HandleErr(err)
	res := &Client{Conn: conn, SendChan: make(chan *Req, 1024), Resps: make(map[string]*AsyncResp)}
	go res.readLoop()
	go res.writeLoop()
	return res
}
