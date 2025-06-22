package main

import (
	"hash"
	"hash/fnv"
	"time"
)

type Entry struct {
	Type     int
	Str      string    // string  str/int 都存储在这里暂时不单独存储 int 编码
	Time     time.Time // 过期时间
	SkipList *SkipList
	Version  int
}

type Shard struct {
	Data map[string]*Entry
	// 每个分片独立的读写锁
}

func (s *Shard) Put(key string, entry *Entry) bool {
	_, has := s.Data[key]
	s.Data[key] = entry
	return !has
}

func (s *Shard) Get(key string) *Entry {
	return s.Data[key]
}

func (s *Shard) Del(key string) bool {
	if _, ok := s.Data[key]; !ok {
		return false
	}
	delete(s.Data, key)
	return true
}

func (s *Shard) ForEach(callback func(string, *Entry)) {
	for key, entry := range s.Data {
		callback(key, entry)
	}
}

func NewShard() *Shard {
	return &Shard{Data: make(map[string]*Entry)}
}

type Map struct {
	Shards   []*Shard    // 分多个加锁，减小锁的粒度
	Count    int         // 一般会修改 Count 为比原来大的 2 的幂数  这样就可以把 % 换成位运算了
	AllCount int         // 一般使用原子 Int 记录总数量
	Hash     hash.Hash64 // hash 函数
}

func (m *Map) GetIndex(key string) uint64 {
	m.Hash.Reset()
	_, err := m.Hash.Write([]byte(key))
	HandleErr(err)
	return m.Hash.Sum64() % uint64(m.Count)
}

func (m *Map) Put(key string, entry *Entry) {
	idx := m.GetIndex(key)
	if m.Shards[idx].Put(key, entry) {
		m.AllCount++
	}
}

func (m *Map) Get(key string) *Entry {
	idx := m.GetIndex(key)
	return m.Shards[idx].Get(key)
}

func (m *Map) Del(key string) {
	idx := m.GetIndex(key)
	if m.Shards[idx].Del(key) {
		m.AllCount--
	}
}

func (m *Map) ForEach(callback func(string, *Entry)) {
	for _, shard := range m.Shards {
		shard.ForEach(callback)
	}
}

func (m *Map) GetSize() int {
	return m.AllCount
}

func NewMap(count int) *Map {
	shards := make([]*Shard, 0)
	for i := 0; i < count; i++ {
		shards = append(shards, NewShard())
	}
	return &Map{Count: count, Shards: shards, Hash: fnv.New64()}
}
