package main

import (
	"math"
	"math/rand"
)

type NodeNext struct {
	Next   *SkipNode
	Offset int // 与前面一个相比相差了多少个节点
}

type SkipNode struct {
	Key   string
	Score float64
	Nexts []*NodeNext
}

func NewSkipNode(height int) *SkipNode {
	nexts := make([]*NodeNext, 0)
	for i := 0; i < height; i++ {
		nexts = append(nexts, &NodeNext{})
	}
	return &SkipNode{Nexts: nexts}
}

type SkipList struct {
	Map    map[string]float64 // key -> score
	Root   *SkipNode          // 头节点是无效占位节点
	Height int
}

func NewSkipList(height int) *SkipList {
	return &SkipList{Height: height, Map: make(map[string]float64), Root: NewSkipNode(height)}
}

func (s *SkipList) Add(key string, score float64) {
	node := NewSkipNode(s.Height)
	node.Key, node.Score = key, score
	if oldScore, ok := s.Map[key]; ok {
		if oldScore == score {
			return // 重复添加
		} else {
			s.Del(key) // 需要先删除
		}
	}
	pres := make([]*SkipNode, 0)
	temp := s.Root
	for i := 0; i < s.Height; i++ {
		for temp.Nexts[i].Next != nil && temp.Nexts[i].Next.Score < score {
			temp = temp.Nexts[i].Next
		}
		pres = append(pres, temp)
	}
	for i := s.Height - 1; i >= 0; i-- {
		// 最后一层肯定是有链接的
		next := pres[i].Nexts[i].Next
		pres[i].Nexts[i].Next = node
		node.Nexts[i].Next = next
		pres[i].Nexts[i].Offset = s.getOffset(pres[i], node)
		node.Nexts[i].Offset = s.getOffset(node, next)
		// 后面每一层有 1/2 的概率停止链接
		if rand.Intn(2) == 0 {
			break
		}
	}
	s.Map[key] = score
}

func (s *SkipList) Del(key string) bool {
	score, ok := s.Map[key]
	if !ok { // 不存在
		return false
	}
	node := s.Root
	for i := 0; i < s.Height; i++ { // 最底层 100% 链接，每上一层有一半的概率建立链接，一旦不再建立就直接结束
		for node.Nexts[i].Next != nil && node.Nexts[i].Next.Score < score {
			node = node.Nexts[i].Next
		}
		if node.Nexts[i].Next != nil && node.Nexts[i].Next.Key == key {
			node.Nexts[i].Next = node.Nexts[i].Next.Nexts[i].Next
			node.Nexts[i].Offset = s.getOffset(node, node.Nexts[i].Next)
		}
	}
	delete(s.Map, key)
	return true
}

func (s *SkipList) Range(start int, end int) []string {
	idx := s.Height - 1
	node := s.Root.Nexts[idx].Next
	if node == nil {
		return make([]string, 0)
	}
	// 跨过开始的节点 实际可以借助上层链接快速找到开始节点  可以借助 offset 快速跳跃
	for i := 0; i < start; i++ {
		node = node.Nexts[idx].Next
		if node == nil {
			return make([]string, 0)
		}
	}
	// 根据结束位置进行搜集
	if end == -1 {
		end = math.MaxInt
	}
	res := make([]string, 0)
	for i := start; i <= end; i++ {
		res = append(res, node.Key)
		node = node.Nexts[idx].Next
		if node == nil {
			return res
		}
	}
	return res
}

func (s *SkipList) getOffset(node *SkipNode, end *SkipNode) int {
	res := 0
	for end != nil && node != end { // 计算偏移
		res++
		node = node.Nexts[s.Height-1].Next
	}
	return res
}

func (s *SkipList) GetMap() map[string]float64 {
	return s.Map
}

func (s *SkipList) GetCount() int {
	return len(s.Map)
}

func (s *SkipList) GetScore(key string) (float64, bool) {
	val, ok := s.Map[key]
	return val, ok
}

func (s *SkipList) GetRank(key string) (int, bool) {
	if _, ok := s.Map[key]; !ok {
		return 0, false
	}
	res := 0
	node := s.Root
	idx := s.Height - 1 // 也可以借助 offset 进行二分查找排名
	for node.Nexts[idx].Next.Key != key {
		res++
		node = node.Nexts[idx].Next
	}
	return res, true
}
