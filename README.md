# Go 语言实现的 Redis
> 参考教程：https://github.com/gofish2020/easyredis
## 指令支持
string：set get(支持多个无需mset,mget) incrby(支持负数无需decr) setnx(同样支持多个) setex<br>
zset：zadd zrem zrange zcard zscore zrank<br>
系统：ping auth select dbsize bgrewriteaof<br>
消息订阅：subscribe unsubscribe publish<br>
事务：multi discard exec watch unwatch<br>
key管理：exists type ttl del expire persist<br>
## 其他特性
支持 aof 日志与 redis 启动自动重放