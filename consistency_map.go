package main

import (
	"fmt"
	"hash"
	"hash/fnv"
	"math"
)

type Range struct {
	Start uint64
	End   uint64
	Size  uint64
}

type Stack struct {
	Data  []*Range
	Count int
}

func (s *Stack) Push(item *Range) {
	if s.Count < len(s.Data) {
		s.Data[s.Count] = item
	} else {
		s.Data = append(s.Data, item)
	}
	s.Count++
	s.Float(s.Count - 1)
}

func (s *Stack) Pop() *Range {
	res := s.Data[0]
	s.Count--
	s.Data[0] = s.Data[s.Count]
	s.Sink(0)
	return res
}

// 0
// 1 2
func (s *Stack) Float(i int) {
	if i == 0 {
		return
	}
	if s.Data[i].Size > s.Data[(i-1)/2].Size {
		s.Data[i], s.Data[(i-1)/2] = s.Data[(i-1)/2], s.Data[i]
		s.Float((i - 1) / 2)
	}
}

func (s *Stack) Sink(i int) {
	target := i
	if i*2+1 < s.Count && s.Data[i*2+1].Size > s.Data[target].Size {
		target = i*2 + 1
	}
	if i*2+2 < s.Count && s.Data[i*2+2].Size > s.Data[target].Size {
		target = i*2 + 2
	}
	if target != i {
		s.Data[i], s.Data[target] = s.Data[target], s.Data[i]
		s.Sink(target)
	}
}

func NewStack() *Stack {
	return &Stack{Data: make([]*Range, 0), Count: 0}
}

type SubMap struct {
	Start uint64
	End   uint64
	Data  map[string]string
}

func NewSubMap(start uint64, end uint64) *SubMap {
	return &SubMap{Start: start, End: end, Data: make(map[string]string)}
}

type ConsistencyMap struct {
	Stack        *Stack          // 最大范围堆  每次添加节点都是拆分最大范围
	InvalidRange map[string]bool // 违规的范围 删除不太方便标记违规
	Data         []*SubMap       // 至少包含一个 0~2^64
	Hash         hash.Hash64     // 一致性 hash
}

func NewConsistencyMap() *ConsistencyMap {
	stack := NewStack()
	stack.Push(&Range{0, math.MaxUint64, math.MaxUint64})
	data := make([]*SubMap, 0)
	data = append(data, NewSubMap(0, math.MaxUint64))
	return &ConsistencyMap{
		Stack:        stack,
		InvalidRange: make(map[string]bool),
		Data:         data,
		Hash:         fnv.New64(),
	}
}

func (c *ConsistencyMap) RemoveNode(start uint64) {
	if start == 0 {
		return // 暂时不允许移除从 0 开始的节点
	}
	var m1, m2 *SubMap
	data := make([]*SubMap, 0)
	for _, item := range c.Data { // 先简单遍历
		if item.End == start {
			m1 = item
		} else if item.Start == start {
			m2 = item
		} else {
			data = append(data, item)
		}
	}
	if m1 == nil || m2 == nil {
		panic(fmt.Sprintf("not find node of %d", start))
	}
	// 合并区间移除无效范围  不太方便移除 直接标记失效
	c.InvalidRange[c.genRangeKey(&Range{Start: m1.Start, End: m1.End})] = true
	c.InvalidRange[c.genRangeKey(&Range{Start: m2.Start, End: m2.End})] = true
	for key, val := range m2.Data { // 合并数据
		m1.Data[key] = val
	}
	m1.End = m2.End
	c.Data = append(data, m1)
	c.Stack.Push(&Range{Start: m1.Start, End: m1.End, Size: m1.End - m1.Start})
}

func (c *ConsistencyMap) AddNode() { // 添加新节点并迁移数据
	item := c.Stack.Pop() // 默认向占领区域最大的地方添加新节点
	for {
		key := c.genRangeKey(item)
		if c.InvalidRange[key] { // 注意清除无效值
			delete(c.InvalidRange, key)
			item = c.Stack.Pop()
		} else {
			break
		}
	}
	var oldM *SubMap // 一样可以使用 二分 暂时不使用
	for _, sub := range c.Data {
		if sub.Start == item.Start {
			oldM = sub
			break
		}
	}
	if oldM == nil {
		panic(fmt.Sprintf("oldM of %d %d not find", item.Start, item.End))
	}
	mid := item.Start + item.Size/2
	oldM.End = mid
	m := NewSubMap(mid, item.End)
	for key, val := range oldM.Data {
		c.Hash.Reset()
		_, err := c.Hash.Write([]byte(key))
		HandleErr(err)
		temp := c.Hash.Sum64() // 有需要就进行数据转移
		if temp >= m.Start && temp < m.End {
			m.Data[key] = val
			delete(oldM.Data, key)
		}
	}
	c.Stack.Push(&Range{Start: m.Start, End: m.End, Size: m.End - m.Start})
	c.Stack.Push(&Range{Start: oldM.Start, End: oldM.End, Size: oldM.End - oldM.Start})
	c.Data = append(c.Data, m)
}

func (c *ConsistencyMap) Get(key string) string {
	m := c.getSubMap(key)
	return m.Data[key]
}

func (c *ConsistencyMap) Set(key, val string) {
	m := c.getSubMap(key)
	m.Data[key] = val
}

func (c *ConsistencyMap) Del(key string) {
	m := c.getSubMap(key)
	delete(m.Data, key)
}

func (c *ConsistencyMap) getSubMap(key string) *SubMap {
	c.Hash.Reset()
	_, err := c.Hash.Write([]byte(key))
	HandleErr(err)
	val := c.Hash.Sum64()
	for _, item := range c.Data {
		if val >= item.Start && val < item.End {
			return item // 可以使用 二分 这里简单起见 不再使用二分
		}
	}
	return nil
}

func (c *ConsistencyMap) genRangeKey(item *Range) string {
	return fmt.Sprintf("%d-%d", item.Start, item.End)
}
