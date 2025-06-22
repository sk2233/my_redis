package main

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

//func TestClient(t *testing.T) {
//	conn, err := net.Dial("tcp", "127.0.0.1:3000")
//	HandleErr(err)
//	WriteObj(conn, &Req{
//		Cmd: "get",
//	})
//}

func TestSelect(t *testing.T) {
	stopChan := make(chan struct{})
	timeChan := time.Tick(time.Second)
	count := 10
	for {
		select { // 多路复用？  非常好用
		case <-stopChan: // 停止信号
			fmt.Println("END")
			return
		case temp := <-timeChan: // 定时器
			fmt.Println(temp.Second())
			count--
			if count < 0 {
				stopChan <- struct{}{}
			}
		}
	}
}

func TestTimeWheel(t *testing.T) {
	tw := NewTimeWheel(8, time.Second)
	for i := 0; i < 16; i++ {
		tw.AddTask(strconv.FormatInt(int64(i), 10), func(key string) {
			fmt.Println(key, time.Now())
		}, time.Second*time.Duration(i))
	}
	tw.Start()
	for {

	}
}

var (
	arr = make([]byte, 0)
)

func TestChan(t *testing.T) {
	chan1 := make(chan []byte, 1024)
	chan2 := make(chan []byte, 1024)
	go func() {
		for {
			//arr = append(arr, 'a', 'b', 'c', 'd')
			chan1 <- []byte{'a', 'b', 'c', 'd'}
		}
	}()
	go func() {
		for {
			//arr = append(arr, 'e', 'f', 'g', 'h')
			chan2 <- []byte{'e', 'f', 'g', 'h'}
		}
	}()
	//time.Sleep(time.Millisecond)
	for {
		select {
		case bs1 := <-chan1:
			for _, b := range bs1 {
				arr = append(arr, b)
			}
		case bs2 := <-chan2:
			for _, b := range bs2 {
				arr = append(arr, b)
			}
		case <-time.After(time.Second):
			fmt.Println(string(arr))
			return
		}
	}
}

func TestMap(t *testing.T) {
	m := NewConsistencyMap()
	for i := 0; i < 1000; i++ {
		str := strconv.FormatInt(int64(i), 10)
		m.Set(str, str)
	}
	for i := 0; i < 1000; i++ {
		if rand.Intn(3) > 1 {
			m.AddNode()
		}
		str := strconv.FormatInt(int64(i), 10)
		fmt.Println(str, m.Get(str))
	}
	for _, item := range m.Data {
		if len(item.Data) > 0 {
			fmt.Println(len(item.Data))
		}
	}
}

func TestClient(t *testing.T) {
	//client := NewClient("127.0.0.1:3000")
	var line string
	for {
		_, err := fmt.Scanln(&line)
		HandleErr(err)
		//client.Send()
	}
}

func TestHash(t *testing.T) {
	res := fnv.New64()
	res.Reset()
	res.Write([]byte("2233"))
	fmt.Println(res.Sum64())
	res.Reset()
	res.Write([]byte("2233"))
	fmt.Println(res.Sum64())
	res.Reset()
	res.Write([]byte("1122"))
	fmt.Println(res.Sum64())
	res.Reset()
	res.Write([]byte("1122"))
	fmt.Println(res.Sum64())
	fmt.Println(GenID())
}

func TestRead(t *testing.T) {
	//file, err := os.Open("/Users/sky/GolandProjects/my_redis/data/aof.log")
	//HandleErr(err)
	//file.Seek(100, 0)
	//buff := &bytes.Buffer{}
	//io.Copy(buff, file)
	//fmt.Println(buff.String())
	fmt.Println(strconv.FormatFloat(23.232323, 'f', -1, 64))
}
