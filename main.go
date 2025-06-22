package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// https://github.com/gofish2020/easyredis
// string set get(支持多个无需mset,mget) incrby(支持负数无需decr) setnx(同样支持多个) setex
// zset zadd zrem zrange zcard zscore zrank
// hash 简单 map 暂不支持
// list 简单双向链表暂不支持
// set 类似值为 null 的 hash 暂时不支持
// ping auth select dbsize bgrewriteaof
// subscribe unsubscribe publish
// multi discard exec watch unwatch
// exists type ttl del expire persist
// scan 是每次扫描，以一个分片 map 下的一个 hash 槽为单位进行扫描 返回数量可能大于 count

func main() {
	// 非阻塞启动服务
	conf := GetConf()
	server := NewServer(conf)
	server.Start()
	// 阻塞客户端
	client := NewClient("127.0.0.1:3000")
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		HandleErr(err)
		items := strings.Split(text, " ")
		temp := make([]string, 0)
		for _, item := range items {
			item = strings.TrimSpace(item)
			if len(item) > 0 {
				temp = append(temp, item)
			}
		}
		if len(temp) == 0 {
			Error("Err Input : %s", text)
			continue
		}
		resp := client.Send(&Req{
			SeqID: GenID(),
			Cmd:   temp[0],
			Args:  temp[1:],
		})
		res := resp.Load()
		fmt.Println(res.Cmd, res.Args)
	}
}
