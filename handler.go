package main

import (
	"net"
	"strconv"
	"strings"
)

type Handler struct {
	Conf   *Conf
	DBs    []*DB
	AOF    *AOF
	Pubhub *Pubhub
}

func (h *Handler) Handle(conn net.Conn) {
	session := NewSession(conn) // 记录一次连接的相关信息
	for {
		req := session.ReadReq()
		Info("req %s", ToStr(req))
		// ping 与 auth 是不需要登录的
		cmd := strings.ToUpper(req.Cmd)
		if cmd == CmdPing {
			h.HandlePing(req, session)
			continue
		}
		if cmd == CmdAuth {
			h.HandleAuth(req, session)
			continue
		}
		// 检查登录
		if !session.Auth {
			session.WriteError(req.SeqID, "Need Auth")
			continue
		}
		h.HandleDBCmd(req, session, true)
	}
}

func (h *Handler) HandleDBCmd(req *Req, session *Session, writeAOF bool) {
	cmd := strings.ToUpper(req.Cmd)
	switch cmd {
	case CmdSelect:
		h.HandleSelect(req, session)
	case CmdDBSize:
		h.HandleDBSize(req, session)
	case CmdBGRewriteAOF:
		h.HandleBGRewriteAOF(req, session)
	case CmdSubscribe:
		h.Pubhub.Subscribe(req, session)
	case CmdUnsubscribe:
		h.Pubhub.Unsubscribe(req, session)
	case CmdPublish:
		h.Pubhub.Publish(req, session)
	default: // 剩下的就是 DB 命令了
		h.ExecDB(req, session, writeAOF)
	}
}

func (h *Handler) Close() {

}

func (h *Handler) HandlePing(req *Req, session *Session) {
	if len(req.Args) > 1 {
		session.WriteError(req.SeqID, "Invalid Args")
		return
	}
	session.WriteResp(&Resp{
		SeqID: req.SeqID,
		Cmd:   "PONG",
		Args:  req.Args,
	})
}

func (h *Handler) HandleAuth(req *Req, session *Session) {
	if len(req.Args) != 1 {
		session.WriteError(req.SeqID, "Invalid Args")
		return
	}
	if req.Args[0] == h.Conf.Passwd {
		session.WriteOk(req.SeqID)
		session.Auth = true
	} else {
		session.WriteError(req.SeqID, "Invalid Passwd")
	}
}

func (h *Handler) HandleSelect(req *Req, session *Session) {
	if len(req.Args) != 1 {
		session.WriteError(req.SeqID, "Invalid Args")
		return
	}
	if session.InTransaction { // 事务中不允许切换数据库
		session.WriteError(req.SeqID, "Invalid Select In Transaction")
		return
	}
	index, err := strconv.ParseInt(req.Args[0], 10, 64)
	if err != nil {
		session.WriteError(req.SeqID, "Invalid Index")
		Error("Invalid Index %s , err %s", req.Args[0], ToStr(err))
		return
	}
	if index < 0 && index >= int64(h.Conf.MaxDB) {
		session.WriteError(req.SeqID, "Index Out Of Range")
		return
	}
	session.DBIndex = int(index)
	session.WriteOk(req.SeqID)
}

func (h *Handler) ExecDB(req *Req, session *Session, writeAOF bool) {
	db := h.DBs[session.DBIndex]
	db.Exec(req, session, h.AOF, writeAOF)
}

func (h *Handler) ForEach(callback func(int, string, *Entry)) {
	for i := 0; i < len(h.DBs); i++ {
		h.DBs[i].ForEach(func(key string, entry *Entry) {
			callback(i, key, entry)
		})
	}
}

func (h *Handler) HandleBGRewriteAOF(req *Req, session *Session) {
	go h.AOF.ReWrite(h.Conf.AOFFile, h.Conf)
	session.WriteOk(req.SeqID)
}

func (h *Handler) HandleDBSize(req *Req, session *Session) {
	db := h.DBs[session.DBIndex]
	session.WriteNum(req.SeqID, db.GetSize())
}

func NewHandler(conf *Conf) *Handler {
	dbs := make([]*DB, 0)
	for i := 0; i < conf.MaxDB; i++ {
		dbs = append(dbs, NewDB(conf))
	}
	res := &Handler{Conf: conf, DBs: dbs, Pubhub: NewPubhub(), AOF: NewAOF(conf.AOFFile, conf.AOFFsync)}
	res.AOF.LoadAOF(res, conf.AOFFile, 0)
	return res
}
