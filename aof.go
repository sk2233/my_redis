package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type AOF struct {
	File      *os.File
	Fsync     string
	LastIndex int
}

var (
	aofCmdSet = map[string]bool{ // 只记录对数据有修改的
		CmdSet:     true,
		CmdIncrBy:  true,
		CmdSetNX:   true,
		CmdSetEX:   true,
		CmdZAdd:    true,
		CmdZRem:    true,
		CmdDel:     true,
		CmdExpire:  true,
		CmdPersist: true,
	}
)

func (a *AOF) WriteAOF(req *Req, index int) {
	cmd := strings.ToUpper(req.Cmd)
	if !aofCmdSet[cmd] {
		return
	}
	if a.LastIndex != index { // 切换数据库
		a.writeReq(a.File, &Req{
			Cmd:  CmdSelect,
			Args: []string{strconv.FormatInt(int64(index), 10)},
		})
		a.LastIndex = index
	}
	switch strings.ToUpper(req.Cmd) {
	case CmdSetEX: // 相对超时时间需要修改为绝对的
		a.writeReq(a.File, &Req{
			Cmd:  CmdSet,
			Args: req.Args[:2],
		})
		ttl, _ := strconv.ParseInt(req.Args[2], 10, 64)
		a.writeReq(a.File, &Req{
			Cmd:  CmdAbsExpire,
			Args: []string{req.Args[0], strconv.FormatInt(ttl+time.Now().Unix(), 10)},
		})
	case CmdExpire: // 相对超时时间需要修改为绝对的
		ttl, _ := strconv.ParseInt(req.Args[1], 10, 64)
		a.writeReq(a.File, &Req{
			Cmd:  CmdAbsExpire,
			Args: []string{req.Args[0], strconv.FormatInt(ttl+time.Now().Unix(), 10)},
		})
	default:
		a.writeReq(a.File, req)
	}
	if a.Fsync == FsyncAlways { // 判断是不是每次都要刷盘
		err := a.File.Sync()
		HandleErr(err)
	}
}

func (a *AOF) fsyncEverySec() {
	timeChan := time.Tick(time.Second)
	for {
		select {
		case <-timeChan:
			err := a.File.Sync()
			HandleErr(err)
		}
	}
}

func (a *AOF) writeReq(writer io.Writer, req *Req) {
	bs, err := json.Marshal(req)
	HandleErr(err)
	bs = append(bs, '\r', '\n')
	_, err = writer.Write(bs)
	HandleErr(err)
}

type FakeConn struct {
	*net.TCPConn
}

func NewFakeConn() *FakeConn {
	return &FakeConn{TCPConn: &net.TCPConn{}}
}

func (f *FakeConn) Write(bs []byte) (n int, err error) {
	Info("FakeConn Write %s", string(bs))
	return 0, nil
}

func (a *AOF) LoadAOF(handler *Handler, fileName string, size int64) {
	if !FileExist(fileName) {
		return
	}
	// 若是指定了大小只重放指定大小的
	bs, err := os.ReadFile(fileName)
	HandleErr(err)
	if size > 0 {
		bs = bs[:size]
	}
	lines := bytes.Split(bs, []byte("\r\n"))
	// 使用虚假的 session 进行重放
	session := NewSession(NewFakeConn())
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		req := &Req{}
		err = json.Unmarshal(line, req)
		HandleErr(err)
		handler.HandleDBCmd(req, session, false)
	}
}

func (a *AOF) ReWrite(fileName string, conf *Conf) { // AOF 日志重写
	file, err := os.Open(fileName)
	if os.IsNotExist(err) {
		return
	}
	HandleErr(err)
	info, err := file.Stat()
	HandleErr(err)
	// 记录快照点该点之前的参与本次压缩，之后的正常存储
	size := info.Size()
	// 重放日志到临时内存
	handler := NewHandler(conf)
	a.LoadAOF(handler, fileName, size)
	// 生成指令
	lastIdx := -1
	buff := &bytes.Buffer{} // idx 是顺序来的一般不会变化太大
	handler.ForEach(func(idx int, key string, entry *Entry) {
		if lastIdx != idx {
			lastIdx = idx
			a.writeReq(buff, &Req{
				Cmd:  CmdSelect,
				Args: []string{strconv.FormatInt(int64(idx), 10)},
			})
		}
		a.writeEntry(buff, key, entry)
	})
	// 写入 aof 文件 并重新打开文件
	_, err = file.Seek(size, 0)
	HandleErr(err)
	_, err = io.Copy(buff, file) // 先把剩余的与压缩后的放到一块，写入文件
	HandleErr(err)
	a.File.Close() // 写入文件
	err = os.WriteFile(fileName, buff.Bytes(), 0777)
	HandleErr(err) // 重新打开
	a.File, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	HandleErr(err)
}

func (a *AOF) writeEntry(buff *bytes.Buffer, key string, entry *Entry) {
	switch entry.Type {
	case TypeStr:
		a.writeReq(buff, &Req{
			Cmd:  CmdSet,
			Args: []string{key, entry.Str},
		})
	case TypeZSet:
		args := make([]string, 0)
		args = append(args, key)
		m := entry.SkipList.GetMap()
		for name, score := range m {
			args = append(args, name, strconv.FormatFloat(score, 'f', -1, 64))
		}
		a.writeReq(buff, &Req{
			Cmd:  CmdZAdd,
			Args: args,
		})
	default:
		panic(fmt.Sprintf("unkown type %v", entry.Type))
	}
}

func NewAOF(fileName string, fsync string) *AOF {
	// 追加写文件
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	HandleErr(err)
	res := &AOF{File: file, Fsync: fsync, LastIndex: -1}
	if fsync == FsyncEverySec {
		go res.fsyncEverySec()
	}
	return res
}
