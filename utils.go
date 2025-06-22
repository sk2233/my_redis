package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

// target 必须是指针
func ReadObj(reader io.Reader, target any) {
	// 先获取数量
	bs := make([]byte, 4)
	_, err := reader.Read(bs)
	HandleErr(err)
	count := binary.LittleEndian.Uint32(bs)
	// 再解析对象
	bs = make([]byte, count)
	_, err = reader.Read(bs)
	HandleErr(err)
	err = json.Unmarshal(bs, target)
	HandleErr(err)
}

func WriteObj(writer io.Writer, target any) {
	bs, err := json.Marshal(target)
	HandleErr(err)
	// 先写入数量
	temp := make([]byte, 4)
	binary.LittleEndian.PutUint32(temp, uint32(len(bs)))
	_, err = writer.Write(temp)
	HandleErr(err)
	// 再写入数据
	_, err = writer.Write(bs)
	HandleErr(err)
}

func ToStr(obj any) string {
	bs, err := json.Marshal(obj)
	HandleErr(err)
	return string(bs)
}

func FileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	panic(err)
}

func GenID() string {
	return fmt.Sprintf("%d-%03d", time.Now().Unix(), rand.Intn(1000))
}
