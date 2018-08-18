package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Statist struct {
	cLock   sync.RWMutex // protect the Devices
	Devices map[string]*Device
	OnCnt   int32
	AllCnt  int32
}

type Device struct {
	Online  string
	OnAt    int64
	Offline string
	OffAt   int64
	RxCnt   int64
	Mins    int64
	Average int64
}

var (
	wsAddr   = "127.0.0.1:8082"
	httpAddr = "http://127.0.0.1:8081"

	begin    = 0  // 客户端连接起始
	end      = 1  // 客户端连接结束
	interval = 10 // 每10毫秒创建一个客户端连接
	freq     = 30 // 心跳间隔 30 秒

	debug  = true
	prds   = []Product{}
	prdMap = map[string]Product{} // 存储所有产品id对应的产品信息, prdId： prdInfo
)

var statistics *Statist

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := initHttp(); err != nil {
		fmt.Printf("initHttp Err===>: %v\n", err)
		return
	}
	statistics = new(Statist)
	statistics.Devices = make(map[string]*Device, 100)

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
