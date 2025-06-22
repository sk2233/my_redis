package main

import (
	"strings"
	"time"
)

var (
	cmdMap = make(map[string]Cmd)
)

func RegisterCmd(name string, cmd Cmd) {
	cmdMap[name] = cmd
}

type DB struct {
	DataMap *Map // key -> data
	TTLMap  *Map // key -> 过期时间
}

func (d *DB) Exec(req *Req, session *Session, aof *AOF, writeAOF bool) {
	cmd := strings.ToUpper(req.Cmd)
	switch cmd {
	case CmdMulti:
		d.ExecMulti(req, session)
	case CmdDiscard:
		d.ExecDiscard(req, session)
	case CmdExec:
		d.ExecExec(req, session, aof)
	case CmdWatch:
		d.ExecWatch(req, session)
	case CmdUnwatch:
		d.ExecUnwatch(req, session)
	default:
		d.ExecNormal(req, session, aof, writeAOF)
	}
}

func (d *DB) ExecNormal(req *Req, session *Session, aof *AOF, writeAOF bool) {
	if session.InTransaction {
		// 最好检查一下，不要无脑入队列，前置检查出错误，整个事务都不要执行了
		session.ReqQueue = append(session.ReqQueue, req)
		session.WriteOk(req.SeqID, "EnQueue")
		return
	}
	if writeAOF {
		aof.WriteAOF(req, session.DBIndex)
	}
	// 正常指令  get  set  等
	cmd := cmdMap[strings.ToUpper(req.Cmd)]
	if cmd == nil {
		session.WriteError(req.SeqID, "Invalid Cmd")
		Error("Invalid Cmd %s", req.Cmd)
		return
	}
	// 执行指令
	cmd.Exec(d, req, session)
}

func (d *DB) GetOrPutEntry(key string, entry *Entry) *Entry {
	res := d.GetEntry(key)
	if res != nil { // 存在直接使用
		return res
	}
	d.PutEntry(key, entry) // 不存在进行添加
	return entry
}

func (d *DB) PutEntry(key string, entry *Entry) {
	d.DataMap.Put(key, entry)
}

func (d *DB) GetEntry(key string) *Entry {
	if d.IsExpire(key) { // 惰性删除
		d.DelEntry(key)
		return nil
	}
	return d.DataMap.Get(key)
}

func (d *DB) IsExpire(key string) bool {
	entry := d.TTLMap.Get(key)
	if entry == nil {
		return false // 没有过期时间
	}
	return entry.Time.Before(time.Now())
}

func (d *DB) DelEntry(key string) {
	d.DataMap.Del(key)
	d.TTLMap.Del(key)
}

func (d *DB) ForEach(callback func(string, *Entry)) {
	d.DataMap.ForEach(callback)
}

func (d *DB) ExecMulti(req *Req, session *Session) {
	if session.InTransaction {
		session.WriteError(req.SeqID, "Already In Transaction")
		return
	}
	session.InTransaction = true
	session.ReqQueue = make([]*Req, 0)
	session.WriteOk(req.SeqID)
}

func (d *DB) ExecDiscard(req *Req, session *Session) {
	if !session.InTransaction {
		session.WriteError(req.SeqID, "Not In Transaction")
		return
	}
	session.InTransaction = false
	session.WriteNum(req.SeqID, len(session.ReqQueue))
}

func (d *DB) ExecExec(req *Req, session *Session, aof *AOF) {
	if !session.InTransaction {
		session.WriteError(req.SeqID, "Not In Transaction")
		return
	}
	if d.WatchKeyChange(session.WatchKey) {
		session.InTransaction = false // 需要退出事务
		session.WriteError(req.SeqID, "Exec Fail WatchKey Change")
		return
	}
	for _, item := range session.ReqQueue { // 队列任务全部执行了
		d.ExecNormal(item, session, aof, false)
	}
	session.WriteNum(req.SeqID, len(session.ReqQueue))
}

func (d *DB) ExecWatch(req *Req, session *Session) {
	if len(req.Args) == 0 {
		session.WriteError(req.SeqID, "Invalid Watch Param")
		return
	}
	if session.InTransaction {
		session.WriteError(req.SeqID, "Watch Not Allow In Transaction")
		return
	}
	count := 0
	for _, key := range req.Args {
		entry := d.GetEntry(key)
		if entry != nil {
			session.WatchKey[key] = entry.Version
			count++
		}
	}
	session.WriteNum(req.SeqID, count)
}

func (d *DB) ExecUnwatch(req *Req, session *Session) {
	if len(req.Args) == 0 {
		session.WriteError(req.SeqID, "Invalid Unwatch Param")
		return
	}
	count := 0
	for _, key := range req.Args {
		if _, ok := session.WatchKey[key]; ok {
			delete(session.WatchKey, key)
			count++
		}
	}
	session.WriteNum(req.SeqID, count)
}

func (d *DB) WatchKeyChange(watchKey map[string]int) bool {
	for key, version := range watchKey {
		entry := d.GetEntry(key) // 删除或版本不对都是修改了
		if entry == nil || entry.Version != version {
			return true
		}
	}
	return false
}

func (d *DB) RemoveTTL(key string) {
	d.TTLMap.Del(key)
}

func (d *DB) SetTTL(key string, ttl int) {
	if ttl > 0 {
		d.TTLMap.Put(key, &Entry{
			Type: TypeTime,
			Time: time.Now().Add(time.Duration(ttl) * time.Second),
		})
	} else { // 太小直接删除
		d.DelEntry(key)
	}
}

func (d *DB) SetAbsTTL(key string, ttl int) {
	time0 := time.Unix(int64(ttl), 0)
	if time0.After(time.Now()) {
		d.TTLMap.Put(key, &Entry{
			Type: TypeTime,
			Time: time0,
		})
	} else { // 在之前或者相等直接删除
		d.DelEntry(key)
	}
}

func (d *DB) GetTTL(key string) int {
	entry := d.TTLMap.Get(key)
	if entry == nil {
		return -1
	}
	return int(entry.Time.Sub(time.Now()) / time.Second)
}

func (d *DB) GetSize() int {
	return d.DataMap.GetSize()
}

func NewDB(conf *Conf) *DB {
	return &DB{
		DataMap: NewMap(conf.ShardCount),
		TTLMap:  NewMap(conf.ShardCount),
	}
}
