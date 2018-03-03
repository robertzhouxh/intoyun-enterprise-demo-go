package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/Terry-Mao/goconf"
)

var (
	gconf    *goconf.Config
	Conf     *Config
	confFile string
)

func init() {
	flag.StringVar(&confFile, "c", "./intoyun-enterprise-demo-go.conf", " set intoyun-enterprise-demo-go config file path")
}

type Config struct {
	// base section
	PidFile          string        `goconf:"base:pidfile"`
	Dir              string        `goconf:"base:dir"`
	Log              string        `goconf:"base:log"`
	MaxProc          int           `goconf:"base:maxproc"`
	HTTPAddrs        []string      `goconf:"base:http.addrs:,"`
	HTTPReadTimeout  time.Duration `goconf:"base:http.read.timeout:time"`
	HTTPWriteTimeout time.Duration `goconf:"base:http.write.timeout:time"`
	BucketChannel    int           `goconf:"bucket:channel"`
	SvrProto         int           `goconf:"proto:svr.proto"`
	CliProto         int           `goconf:"proto:cli.proto"`

	// app
	AppId     string `goconf:"app:appid"`
	AppSecret string `goconf:"app:appsecret"`

	Debug bool `goconf:"base:debug"`

	// push
	PushChan     int `goconf:"push:chan"`
	PushChanSize int `goconf:"push:chan.size"`

	// websocket
	WebsocketBind []string `goconf:"websocket:bind:,"`

	// kafka
	KafkaAddrs   []string `goconf:"kafka:kafka.list:,"`
	SaslEnable   bool     `goconf:"kafka:sasl.enable"`
	SaslUser     string   `goconf:"kafka:sasl.user"`
	SaslPassword string   `goconf:"kafka:sasl.password"`
	KafkaTopic   string   `goconf:"kafka:topic"`
	Group        string   `goconf:"kafka:group"`
}

func NewConfig() *Config {
	return &Config{
		PidFile:       "/tmp/intoyun-enterprise-demo-go.pid",
		Dir:           "./",
		Log:           "./intoyun-enterprise-demo-go-log.xml",
		MaxProc:       runtime.NumCPU(),
		HTTPAddrs:     []string{"8080"},
		WebsocketBind: []string{"0.0.0.0:8080"},
		CliProto:      5,
		SvrProto:      80,
		BucketChannel: 1024,
		Debug:         true,
		PushChan:      10,
		PushChanSize:  100,
	}
}

func InitConfig() (err error) {
	Conf = NewConfig()
	gconf = goconf.New()
	if err = gconf.Parse(confFile); err != nil {
		return err
	}
	if err := gconf.Unmarshal(Conf); err != nil {
		return err
	}
	return nil
}

func ReloadConfig() (*Config, error) {
	conf := NewConfig()
	ngconf, err := gconf.Reload()
	if err != nil {
		return nil, err
	}
	if err := ngconf.Unmarshal(conf); err != nil {
		return nil, err
	}
	gconf = ngconf
	return conf, nil
}
