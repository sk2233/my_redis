package main

import "time"

type TaskNode struct {
	Task   func(key string)
	Key    string // 对应的 key
	Circle int
	Next   *TaskNode
}

type TimeWheel struct {
	Slots    []*TaskNode    // 简单时间轮  数组+链表
	TaskMap  map[string]int // key->SlotIdx
	Interval time.Duration
	Index    int
}

func NewTimeWheel(size int, interval time.Duration) *TimeWheel {
	return &TimeWheel{Slots: make([]*TaskNode, size), TaskMap: make(map[string]int), Interval: interval}
}

func (t *TimeWheel) AddTask(key string, task func(key string), delay time.Duration) {
	count := int((delay + t.Interval - 1) / t.Interval)
	node := &TaskNode{
		Task:   task,
		Key:    key,
		Circle: count / len(t.Slots),
	}
	idx := count % len(t.Slots)
	t.Slots[idx] = t.addNode(t.Slots[idx], node)
	// 存在历史的清除掉
	if oldIdx, ok := t.TaskMap[key]; ok {
		t.Slots[oldIdx] = t.delNode(t.Slots[oldIdx], key)
	}
	t.TaskMap[key] = idx
}

func (t *TimeWheel) CancelTask(key string) {
	idx, ok := t.TaskMap[key]
	if !ok {
		return
	}
	t.Slots[idx] = t.delNode(t.Slots[idx], key)
}

func (t *TimeWheel) Start() {
	go t.loop()
}

func (t *TimeWheel) addNode(root *TaskNode, node *TaskNode) *TaskNode {
	if root == nil {
		return node
	}
	temp := root
	for temp.Next != nil {
		temp = temp.Next
	}
	temp.Next = node
	return root
}

func (t *TimeWheel) delNode(root *TaskNode, key string) *TaskNode {
	if root.Key == key {
		return root.Next
	}
	temp := root
	for temp.Next != nil && temp.Next.Key != key {
		temp = temp.Next
	}
	if temp.Next != nil {
		temp.Next = temp.Next.Next
	}
	return root
}

func (t *TimeWheel) loop() {
	// 简单时间轮
	timeChan := time.Tick(t.Interval)
	for {
		select { // 推荐添加，删除，等操作都通过通道在这里完成 并发也不用加锁了
		case <-timeChan:
			t.Slots[t.Index] = t.scanSlot(t.Slots[t.Index])
			t.Index = (t.Index + 1) % len(t.Slots)
		}
	}
}

func (t *TimeWheel) scanSlot(node *TaskNode) *TaskNode {
	res := &TaskNode{}
	temp := res
	for node != nil {
		if node.Circle > 0 {
			node.Circle--
			temp.Next = node
			temp = temp.Next
			node = node.Next
			temp.Next = nil
		} else {
			node.Task(node.Key)
			node = node.Next
		}
	}
	return res.Next
}
