package main

// 这里对 redis 协议简单处理 直接使用 json

type Req struct {
	SeqID string
	Cmd   string
	Args  []string
}

type Resp struct {
	SeqID string // 用于匹配Req
	Cmd   string
	Args  []string
}
