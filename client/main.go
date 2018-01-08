package main

import (
	"fmt"
	"runtime"
	"time"
)

var (
	wsAddr   = "127.0.0.1:8082"
	httpAddr = "http://127.0.0.1:8081"

	begin    = 0  // 客户端连接起始
	end      = 1  // 客户端连接结束
	interval = 10 // 每10毫秒创建一个客户端连接
	freq     = 30 // 心跳间隔 30 秒

	prds   = []Product{}
	prdMap = map[string]Product{} // 存储所有产品id对应的产品信息, prdId： prdInfo
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := initHttp(); err != nil {
		fmt.Printf("initHttp Err===>: %v\n", err)
		return
	}

	// 统计qps
	go result()

	// 模拟客户端连接
	for i := begin; i < end; i++ {
		go client(fmt.Sprintf("%d", i))
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

	var exit chan bool
	<-exit
}
