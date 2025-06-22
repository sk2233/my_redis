package main

import (
	"strconv"
)

type Cmd interface {
	Exec(db *DB, req *Req, session *Session)
}

func init() {
	RegisterCmd(CmdSet, &SetCmd{})
	RegisterCmd(CmdGet, &GetCmd{})
	RegisterCmd(CmdIncrBy, &IncrByCmd{})
	RegisterCmd(CmdSetNX, &SetNXCmd{})
	RegisterCmd(CmdSetEX, &SetEXCmd{})

	RegisterCmd(CmdZAdd, &ZAddCmd{})
	RegisterCmd(CmdZRem, &ZRemCmd{})
	RegisterCmd(CmdZRange, &ZRangeCmd{})
	RegisterCmd(CmdZCard, &ZCardCmd{})
	RegisterCmd(CmdZScore, &ZScoreCmd{})
	RegisterCmd(CmdZRank, &ZRankCmd{})

	RegisterCmd(CmdExists, &ExistsCmd{})
	RegisterCmd(CmdType, &TypeCmd{})
	RegisterCmd(CmdTTL, &TTLCmd{})
	RegisterCmd(CmdDel, &DelCmd{})
	RegisterCmd(CmdExpire, &ExpireCmd{})
	RegisterCmd(CmdPersist, &PersistCmd{})

	RegisterCmd(CmdAbsExpire, &AbsExpireCmd{})
}

//============================SetCmd=================================

type SetCmd struct {
}

// set name1 ss name2 mm  会取消 ttl
func (s *SetCmd) Exec(db *DB, req *Req, session *Session) {
	// 先只看最简单的
	if len(req.Args) == 0 || len(req.Args)%2 != 0 {
		session.WriteError(req.SeqID, "Invalid Set Param")
		return
	}
	count := 0
	for i := 0; i < len(req.Args); i += 2 {
		if entry := db.GetEntry(req.Args[i]); entry != nil {
			entry.Str = req.Args[i+1]
			entry.Version++
		} else { // 不存在
			db.PutEntry(req.Args[i], &Entry{
				Type: TypeStr,
				Str:  req.Args[i+1],
			})
			count++
		}
		db.RemoveTTL(req.Args[i])
	}
	session.WriteNum(req.SeqID, count)
}

//===========================GetCmd==============================

type GetCmd struct {
}

// get name1 name2
func (g *GetCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) == 0 {
		session.WriteError(req.SeqID, "Invalid Get Param")
		return
	}
	res := make([]string, 0)
	for _, key := range req.Args {
		entry := db.GetEntry(key)
		if entry != nil {
			res = append(res, entry.Str)
		} else {
			res = append(res, "NIL")
		}
	}
	session.WriteOk(req.SeqID, res...)
}

//=======================IncrByCmd==========================

type IncrByCmd struct {
}

// incrby key num
func (i *IncrByCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 2 {
		session.WriteError(req.SeqID, "Invalid IncrBy Param")
		return
	}
	key := req.Args[0]
	num, err := strconv.ParseFloat(req.Args[1], 64)
	if err != nil {
		Error("Err to ParseFloat %v %v", err, req.Args[1])
		session.WriteError(req.SeqID, "IncrBy Num Err")
		return
	}
	entry := db.GetEntry(key)
	if entry == nil {
		session.WriteError(req.SeqID, "IncrBy Key Not Exist")
		return
	}
	old, err := strconv.ParseFloat(entry.Str, 64)
	if err != nil {
		Error("Err to ParseFloat %v %v", err, entry.Str)
		session.WriteError(req.SeqID, "IncrBy Key Not Num")
		return
	}
	entry.Str = strconv.FormatFloat(old+num, 'f', -1, 64)
	entry.Version++
	session.WriteOk(req.SeqID)
}

//==================SetNXCmd=====================

type SetNXCmd struct {
}

// setnx key1 val key2 val
func (s *SetNXCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) == 0 || len(req.Args)%2 != 0 {
		session.WriteError(req.SeqID, "Invalid SetNX Param")
		return
	}
	count := 0
	for i := 0; i < len(req.Args); i += 2 {
		if entry := db.GetEntry(req.Args[i]); entry != nil {
			entry.Str = req.Args[i+1]
			entry.Version++
			count++
		}
		db.RemoveTTL(req.Args[i])
	}
	session.WriteNum(req.SeqID, count)
}

//======================SetEXCmd==========================

type SetEXCmd struct {
}

// setex key val ttl(s)
func (s *SetEXCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 3 {
		session.WriteError(req.SeqID, "Err SetEX Param")
		return
	}
	key := req.Args[0]
	val := req.Args[1]
	ttl, err := strconv.ParseInt(req.Args[2], 10, 64)
	if err != nil {
		Error("SetEX ParseInt Err %v %v", err, req.Args[2])
		session.WriteError(req.SeqID, "Err SetEX TTL")
		return
	}
	if entry := db.GetEntry(key); entry != nil {
		entry.Str = val
		entry.Version++
	} else {
		db.PutEntry(key, &Entry{
			Type: TypeStr,
			Str:  val,
		})
	}
	db.SetTTL(key, int(ttl))
	session.WriteOk(req.SeqID)
}

//=======================ZAddCmd========================

type ZAddCmd struct {
}

// zadd setname 22 name1 33 name2
func (z *ZAddCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) < 3 || len(req.Args)%2 != 1 {
		session.WriteError(req.SeqID, "Invalid ZAdd Param")
		return
	}
	entry := db.GetOrPutEntry(req.Args[0], &Entry{
		Type:     TypeZSet,
		SkipList: NewSkipList(4),
	})
	entry.Version++
	count := 0
	for i := 1; i < len(req.Args); i += 2 {
		score, err := strconv.ParseFloat(req.Args[i], 10)
		key := req.Args[i+1]
		if err != nil {
			Error("ParseFloat err %v %s", err, req.Args[i])
		} else {
			entry.SkipList.Add(key, score)
			count++
		}
	}
	session.WriteNum(req.SeqID, count)
}

//======================ZRangeCmd========================

type ZRangeCmd struct {
}

// zrange key 1 -1
func (z *ZRangeCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 3 {
		session.WriteError(req.SeqID, "Invalid ZRange Param")
		return
	}
	key := req.Args[0]
	start, err := strconv.ParseInt(req.Args[1], 10, 64)
	if err != nil {
		session.WriteError(req.SeqID, "Invalid Start")
		return
	} // 可以为 -1
	end, err := strconv.ParseInt(req.Args[2], 10, 64)
	if err != nil {
		session.WriteError(req.SeqID, "Invalid End")
		return
	}
	entry := db.GetEntry(key)
	if entry == nil {
		session.WriteOk(req.SeqID, "NIL")
		return
	}
	res := entry.SkipList.Range(int(start), int(end))
	session.WriteOk(req.SeqID, res...)
}

//========================ZRemCmd=======================

type ZRemCmd struct {
}

// zrem key name1 name2
func (z *ZRemCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) < 2 {
		session.WriteError(req.SeqID, "Invalid ZRem Param")
		return
	}
	entry := db.GetEntry(req.Args[0])
	if entry == nil {
		session.WriteNum(req.SeqID, 0)
		return
	}
	entry.Version++
	count := 0
	for _, key := range req.Args[1:] {
		if entry.SkipList.Del(key) {
			count++
		}
	}
	session.WriteNum(req.SeqID, count)
}

//========================ZCardCmd======================

type ZCardCmd struct {
}

// zcard key
func (z *ZCardCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 1 {
		session.WriteError(req.SeqID, "Invalid ZCard Param")
		return
	}
	entry := db.GetEntry(req.Args[0])
	if entry == nil {
		session.WriteNum(req.SeqID, 0)
		return
	}
	session.WriteNum(req.SeqID, entry.SkipList.GetCount())
}

//=====================ZScoreCmd======================

type ZScoreCmd struct {
}

// zscore key name
func (z *ZScoreCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 2 {
		session.WriteError(req.SeqID, "Invalid ZScore Param")
		return
	}
	entry := db.GetEntry(req.Args[0])
	if entry == nil {
		session.WriteError(req.SeqID, "ZSet Not Exist")
		return
	}
	val, ok := entry.SkipList.GetScore(req.Args[1])
	if !ok {
		session.WriteError(req.SeqID, "ZSet Key Not Exist")
		return
	}
	session.WriteOk(req.SeqID, strconv.FormatFloat(val, 'f', -1, 64))
}

//=======================ZRankCmd=====================

type ZRankCmd struct {
}

// zrank key name
func (z *ZRankCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 2 {
		session.WriteError(req.SeqID, "Invalid ZRank Param")
		return
	}
	entry := db.GetEntry(req.Args[0])
	if entry == nil {
		session.WriteError(req.SeqID, "ZSet Not Exist")
		return
	}
	rank, ok := entry.SkipList.GetRank(req.Args[1])
	if !ok {
		session.WriteError(req.SeqID, "ZSet Key Not Exist")
		return
	}
	session.WriteNum(req.SeqID, rank)
}

//======================ExistsCmd========================

type ExistsCmd struct {
}

// exists key
func (e *ExistsCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 1 {
		session.WriteError(req.SeqID, "Invalid Exists Param")
		return
	}
	if db.GetEntry(req.Args[0]) != nil {
		session.WriteNum(req.SeqID, 1)
	} else {
		session.WriteNum(req.SeqID, 0)
	}
}

//=======================TypeCmd========================

type TypeCmd struct {
}

// type key
func (t *TypeCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 1 {
		session.WriteError(req.SeqID, "Invalid Type Param")
		return
	}
	entry := db.GetEntry(req.Args[0])
	if entry == nil {
		session.WriteOk(req.SeqID, "none")
		return
	}
	switch entry.Type {
	case TypeStr:
		session.WriteOk(req.SeqID, "string")
	case TypeZSet:
		session.WriteOk(req.SeqID, "zset")
	default:
		session.WriteOk(req.SeqID, "none")
	}
}

//=====================TTLCmd=======================

type TTLCmd struct {
}

// ttl key
func (t *TTLCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 1 {
		session.WriteError(req.SeqID, "Invalid TTL Param")
		return
	}
	ttl := db.GetTTL(req.Args[0])
	session.WriteNum(req.SeqID, ttl)
}

//=====================DelCmd===================

type DelCmd struct {
}

// del key
func (d *DelCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 1 {
		session.WriteError(req.SeqID, "Invalid Del Param")
		return
	}
	if db.GetEntry(req.Args[0]) == nil {
		session.WriteNum(req.SeqID, 0)
	} else {
		db.DelEntry(req.Args[0])
		session.WriteNum(req.SeqID, 1)
	}
}

//====================ExpireCmd=======================

type ExpireCmd struct {
}

// expire key ttl
func (e *ExpireCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 2 {
		session.WriteError(req.SeqID, "Invalid Expire Param")
		return
	}
	key := req.Args[0]
	ttl, err := strconv.ParseInt(req.Args[1], 10, 64)
	if err != nil {
		Error("Expire TTL Err %v %s", err, req.Args[1])
		session.WriteError(req.SeqID, "Expire TTL Err")
		return
	}
	if db.GetEntry(key) != nil {
		db.SetTTL(key, int(ttl))
		session.WriteNum(req.SeqID, 1)
	} else {
		session.WriteNum(req.SeqID, 0)
	}
}

//====================PersistCmd====================

type PersistCmd struct {
}

// persist key
func (p *PersistCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 1 {
		session.WriteError(req.SeqID, "Invalid Persist Param")
		return
	}
	if db.GetEntry(req.Args[0]) != nil {
		db.RemoveTTL(req.Args[0])
		session.WriteNum(req.SeqID, 1)
	} else {
		session.WriteNum(req.SeqID, 0)
	}
}

//==================AbsExpireCmd====================

type AbsExpireCmd struct {
}

// absexpire key ttl(绝对时间戳)
func (s *AbsExpireCmd) Exec(db *DB, req *Req, session *Session) {
	if len(req.Args) != 2 {
		session.WriteError(req.SeqID, "Invalid AbsExpire Param")
		return
	}
	key := req.Args[0]
	ttl, err := strconv.ParseInt(req.Args[1], 10, 64)
	if err != nil {
		Error("AbsExpire TTL Err %v %s", err, req.Args[1])
		session.WriteError(req.SeqID, "AbsExpire TTL Err")
		return
	}
	if db.GetEntry(key) != nil {
		db.SetAbsTTL(key, int(ttl))
		session.WriteNum(req.SeqID, 1)
	} else {
		session.WriteNum(req.SeqID, 0)
	}
}
