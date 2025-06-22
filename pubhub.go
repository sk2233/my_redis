package main

import "strconv"

type SessionNode struct {
	Session *Session
	Next    *SessionNode
}

type Pubhub struct {
	Data map[string]*SessionNode
	// 注意操作都是需要加锁的  可以根据 通道进行 hash 使用对应的锁 保证同一通道的互斥
}

func NewPubhub() *Pubhub {
	return &Pubhub{Data: make(map[string]*SessionNode)}
}

// Subscribe ch1 ch2
func (p *Pubhub) Subscribe(req *Req, session *Session) {
	if len(req.Args) == 0 {
		session.WriteError(req.SeqID, "Invalid Arg Count")
		return
	}
	args := make([]string, 0)
	for _, channel := range req.Args {
		if session.Subscribe(channel) { // 订阅成功了 是本次新增的
			p.Data[channel] = p.addNode(p.Data[channel], session)
			args = append(args, channel)
		}
	}
	args = append(args, strconv.FormatInt(int64(len(args)), 10))
	session.WriteOk(req.SeqID, args...)
}

// Unsubscribe ch1 ch2
func (p *Pubhub) Unsubscribe(req *Req, session *Session) {
	channels := req.Args
	if len(channels) == 0 { // 没有选择就取消全部
		channels = session.GetChannels()
	}
	args := make([]string, 0)
	for _, channel := range channels {
		if session.Unsubscribe(channel) {
			p.Data[channel] = p.delNode(p.Data[channel], session)
			// 也可以在空的时间进行删除
			args = append(args, channel)
		}
	}
	args = append(args, strconv.FormatInt(int64(len(args)), 10))
	session.WriteOk(req.SeqID, args...)
}

func (p *Pubhub) addNode(root *SessionNode, session *Session) *SessionNode {
	if root == nil {
		return &SessionNode{Session: session}
	}
	temp := root
	for temp.Next != nil {
		temp = temp.Next
	}
	temp.Next = &SessionNode{Session: session}
	return root
}

func (p *Pubhub) delNode(node *SessionNode, session *Session) *SessionNode {
	res := &SessionNode{}
	temp := res
	if node != nil {
		if node.Session != session {
			temp.Next = node
			temp = temp.Next
		}
		node = node.Next
	}
	temp.Next = nil // 最后注意断开
	return res.Next
}

func (p *Pubhub) Publish(req *Req, session *Session) {
	if len(req.Args) != 2 {
		session.WriteError(req.SeqID, "Invalid Arg Count")
		return
	}
	count := 0
	node := p.Data[req.Args[0]]
	for node != nil {
		/// 不要给自己发
		if node.Session != session {
			node.Session.WriteResp(&Resp{
				SeqID: req.SeqID,
				Cmd:   "MESSAGE",
				Args:  req.Args,
			})
			count++
		}
		node = node.Next
	}
	session.WriteNum(req.SeqID, count)
}
