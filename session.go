package main

import (
	"net"
	"strconv"
)

type Session struct {
	Conn          net.Conn
	Auth          bool
	DBIndex       int
	Channels      map[string]bool
	InTransaction bool   // 是否在事务中
	ReqQueue      []*Req // 事务队列
	WatchKey      map[string]int
}

func (s *Session) ReadReq() *Req {
	req := &Req{}
	ReadObj(s.Conn, req)
	return req
}

func (s *Session) WriteResp(resp *Resp) {
	WriteObj(s.Conn, resp)
}

func (s *Session) WriteError(seqID string, args ...string) {
	s.WriteResp(&Resp{
		SeqID: seqID,
		Cmd:   "ERROR",
		Args:  args,
	})
}

func (s *Session) WriteOk(seqID string, args ...string) {
	s.WriteResp(&Resp{
		SeqID: seqID,
		Cmd:   "OK",
		Args:  args,
	})
}

func (s *Session) Subscribe(channel string) bool {
	has := s.Channels[channel]
	s.Channels[channel] = true
	return !has
}

func (s *Session) GetChannels() []string {
	res := make([]string, 0)
	for channel := range s.Channels {
		res = append(res, channel)
	}
	return res
}

func (s *Session) Unsubscribe(channel string) bool {
	has := s.Channels[channel]
	delete(s.Channels, channel)
	return has
}

func (s *Session) WriteNum(seqID string, count int) {
	s.WriteOk(seqID, strconv.FormatInt(int64(count), 10))
}

func NewSession(conn net.Conn) *Session {
	return &Session{Conn: conn, DBIndex: 0, Channels: make(map[string]bool), WatchKey: make(map[string]int)} // 默认选择 0 号
}
