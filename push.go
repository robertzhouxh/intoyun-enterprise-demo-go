package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"intoyun-enterprise-demo-go/libs/crypto/aes"
	"intoyun-enterprise-demo-go/libs/define"
	"intoyun-enterprise-demo-go/libs/proto"
	"math/rand"

	log "github.com/thinkboy/log4go"
)

type pushArg struct {
	Code int32
	//Body []byte
	Body string
}

var (
	pushChs []chan *pushArg
)

func InitPush() error {
	pushChs = make([]chan *pushArg, Conf.PushChan)
	for i := 0; i < Conf.PushChan; i++ {
		pushChs[i] = make(chan *pushArg, Conf.PushChanSize)
		go processPush(pushChs[i])
	}
	return nil
}

// push routine
func processPush(ch chan *pushArg) {
	var (
		arg *pushArg
	)

	for {
		arg = <-ch
		process(arg)
	}
}

func push(msg []byte) (err error) {
	m := &proto.KafkaMsg{}
	if err = json.Unmarshal(msg, m); err != nil {
		log.Error("json.Unmarshal(%s) error(%s)", msg, err)
		return
	}

	// pushChs[rand.Int()%Conf.PushChan] act as load balancer
	switch m.Code {
	case define.WIFI_GPRS_META:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{Code: m.Code, Body: m.Body}
	case define.LORA_GATE_META:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{Code: m.Code, Body: m.Body}
	case define.LORA_NODE_META:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{Code: m.Code, Body: m.Body}
	case define.TCP_WS_META:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{Code: m.Code, Body: m.Body}
	case define.WIFI_GPRS_RX:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{Code: m.Code, Body: m.Body}
	case define.LORA_GATE_RX:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{Code: m.Code, Body: m.Body}
	case define.LORA_NODE_RX:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{Code: m.Code, Body: m.Body}
	case define.TCP_WS_RX:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{Code: m.Code, Body: m.Body}
	case define.ONLINE_CODE:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{Code: m.Code, Body: m.Body}
	default:
		log.Error("unknown operation:%s", m.Code)
	}
	return
}

// decrypt the ciphertext into plaintext
func process(arg *pushArg) {
	//str := string(arg.Body[:])
	str := arg.Body
	ciphertext, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		log.Error("cbcdecrypter failed: err(%v)", err)
		return
	}
	key, _ := hex.DecodeString(Conf.AppSecret)
	plaintext, err := aes.CBCDecrypter(key, ciphertext)
	if err != nil {
		log.Error("cbcdecrypter failed: err(%v)", err)
	}

	pl := aes.PKCS7UPad([]byte(plaintext))
	//log.Debug("原始的实时Body: %v", []byte(pl))
	//dp := map[string]string{}
	//err = json.Unmarshal([]byte(plaintext), &dp)
	//if err != nil {
	//	log.Error("Body 解析失败 %v", err)
	//}

	for _, ch := range Buckets.Channels() {
		//body, _ := json.Marshal(pl)
		//err = ch.Push(&proto.Proto{Operation: arg.Code, Body: body})
		err = ch.Push(&proto.Proto{Operation: arg.Code, Body: []byte(pl)})
		if err != nil {
			log.Error("userCh push failed err: %v", err)
		}
	}
}
