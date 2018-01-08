package main

import (
	"flag"
	"runtime"

	log "github.com/thinkboy/log4go"
)

var (
	Debug         bool
	DefaultServer *Server
	Buckets       *Bucket
)

func main() {

	flag.Parse()

	if err := InitConfig(); err != nil {
		panic(err)
	}

	Debug = Conf.Debug

	runtime.GOMAXPROCS(Conf.MaxProc)
	//src, _ := osext.Executable()
	//dest := filepath.Dir(src)
	//usr, _ := user.Current()

	if runtime.GOOS == "windows" {
	} else {
	}

	log.LoadConfiguration(Conf.Log)
	defer log.Close()

	log.Info("ity-srv start")

	operator := new(DefaultOperator)
	DefaultServer = NewServer(operator)
	Buckets = NewBucket(BucketOptions{ChannelSize: Conf.BucketChannel})

	if err := InitHTTP(); err != nil {
		panic(err)
	}

	if err := InitWebsocket(Conf.WebsocketBind); err != nil {
		panic(err)
	}

	if err := InitPush(); err != nil {
		panic(err)
	}

	if err := InitKafka(); err != nil {
		panic(err)
	}

	InitSignal()
}
