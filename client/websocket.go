package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"intoyun-enterprise-demo-go/libs/define"
	"intoyun-enterprise-demo-go/libs/proto"
	"math"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	countDown int64
	countUp   int64
)

const (
	BOOL  = 0x00
	NUMB  = 0x01
	ENUME = 0x02
	STR   = 0x03
	TRANS = 0x04
)

func result() {
	var (
		lastTimes   int64
		lastTimesUp int64
		diff        int64
		diffUp      int64
		nowCount    int64
		nowCountUp  int64
		timer       = int64(80)
	)

	for {
		nowCount = atomic.LoadInt64(&countDown)
		nowCountUp = atomic.LoadInt64(&countUp)
		diff = nowCount - lastTimes
		diffUp = nowCountUp - lastTimesUp
		lastTimes = nowCount
		lastTimesUp = nowCountUp
		fmt.Println(fmt.Sprintf("%s down:%d down/s:%d", time.Now().Format("2006-01-02 15:04:05"), nowCount, diff/timer))
		fmt.Println(fmt.Sprintf("%s up:%d up/s:%d", time.Now().Format("2006-01-02 15:04:05"), nowCountUp, diffUp/timer))

		t := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("\n\n----------------------------------------统计信息: 在线设备总数[%d]-----------------------------------------------------------------\n\n", statistics.OnCnt)
		for id, dev := range statistics.Devices {
			//fmt.Print("统计Devid: %s Now: %d Dev'sOnAt: %d", time.Now().Unix(), dev.OnAt)
			mins := (time.Now().Unix() - dev.OnAt) / 60
			if mins == 0 {
				mins++
			}
			fmt.Printf("\n设备Id:[%s]-上线时间[%s]-当前时间[%s]-总分钟数[%d]-上传次数[%d]-平均次数[%f]\n", id, dev.Online, t, mins, dev.RxCnt, float64(dev.RxCnt)/float64(mins))
		}
		fmt.Printf("\n\n--------------------------------------------------------------------------------------------------------------------------------------\n\n")
		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func client(key string) {
	u := url.URL{Scheme: "ws", Host: wsAddr, Path: "/sub"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	proto0 := &proto.Proto{}
	proto0.Operation = define.OP_AUTH
	proto0.Body, _ = json.Marshal(map[string]string{"key": key})

	if err = proto0.WriteWebsocket(conn); err != nil {
		fmt.Println("发送认证失败:%v", err)
		return
	}
	atomic.AddInt64(&countUp, 1)
	fmt.Println("连接成功!!!")

	if err = proto0.ReadWebsocket(conn); err != nil {
		fmt.Println("读取认证应答失败:%v", err)
		return
	} else {
		atomic.AddInt64(&countDown, 1)
		fmt.Println("认证成功!!!")
		time.Sleep(time.Second * 1)
	}

	// writer
	go func() {
		proto1 := &(proto.Proto{})
		for {
			// heartbeat
			proto1.Operation = define.OP_HEARTBEAT
			proto1.Body = nil
			if err = proto1.WriteWebsocket(conn); err != nil {
				return
			}
			atomic.AddInt64(&countUp, 1)
			if debug == true {
				fmt.Println("发送心跳===>")
			}
			time.Sleep(time.Second * time.Duration(freq))
		}
	}()

	// reader
	proto2 := &proto.Proto{}
	//rtdata := &RtData{}
	rtdata := map[string]interface{}{}
	ondata := map[string]interface{}{}

	for {
		if err = proto2.ReadWebsocket(conn); err != nil {
			fmt.Println("读取消息失败: %v\n\n", err)
			return
		}

		atomic.AddInt64(&countDown, 1)

		if proto2.Operation == define.OP_HEARTBEAT_REPLY {
			// 每收到一次心跳就重置读取超时时间
			if err = conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
				return
			}
			if debug == true {
				fmt.Println("收到心跳<===")
			}
		} else {
			if debug == true {
				fmt.Printf("\n\n收到推送消息\n---------------------------------------\nCode=%d\nBody=%v\n---------------------------------------\n\n", proto2.Operation, []byte(proto2.Body))
			}

			if proto2.Operation == define.ONLINE_CODE {
				err = json.Unmarshal(proto2.Body, &rtdata)
				if err != nil {
					fmt.Printf("json 解析失败: %v\n\n", err)
				}
				//fmt.Printf("\n---------解析之后的上下线数据-------------\n")
				rtd, _ := rtdata["data"].(string)
				//ondps, _ := base64.StdEncoding.DecodeString(string(rtdata["data"]))
				ondps, _ := base64.StdEncoding.DecodeString(rtd)

				//fmt.Printf("ONLINE: %v", ondps)
				err = json.Unmarshal(ondps, &ondata)
				if err != nil {
					fmt.Printf("json 解析失败: %v\n\n", err)
				}

				dev := new(Device)
				if ondata["online"] == true {
					fmt.Printf("\n\n%s 设备上线 ===> %v\n\n", ondata["devId"], ondata["online"])
					dev.Online = time.Now().Format("2006-01-02 15:04:05")
					dev.OnAt = time.Now().Unix()
					dev.RxCnt = 1
					dev.Average = 0
					statistics.OnCnt++
				} else {
					fmt.Printf("\n\n%s 设备离线 ===> %v\n\n", ondata["devId"], ondata["online"])
					dev.Offline = time.Now().Format("2006-01-02 15:04:05")
					dev.OffAt = time.Now().Unix()
					dev.RxCnt = 0
					dev.RxCnt = 0
					dev.Average = 0
					statistics.OnCnt--
				}
				devId, _ := ondata["devId"].(string)

				statistics.cLock.Lock()
				statistics.Devices[devId] = dev
				statistics.cLock.Unlock()

			}

			if proto2.Operation == define.WIFI_GPRS_RX || proto2.Operation == define.LORA_NODE_RX {
				err = json.Unmarshal(proto2.Body, &rtdata)
				if err != nil {
					fmt.Printf("json 解析失败: %v\n\n", err)
				}

				//rxData type: {"devId": <DeviceId>, "prdId": <ProductId>, "stoId": <StoreId>, "data": <mqtt_payload_after_base64_encode>}
				if debug {
					fmt.Printf("解析结果: %v\n\n", rtdata)
				}

				statistics.cLock.Lock()
				did, _ := rtdata["devId"].(string)
				if statistics.Devices[did] != nil {
					statistics.Devices[did].RxCnt++
				} else {
					fmt.Printf("\n!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! LORA 设备上线[%s] 初始化开始 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n", did)
					dev := new(Device)
					dev.Online = time.Now().Format("2006-01-02 15:04:05")
					dev.OnAt = time.Now().Unix()
					dev.RxCnt = 1
					dev.Average = 0
					statistics.Devices[did] = dev
					statistics.OnCnt++
					fmt.Printf("\n!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! LORA 设备[%s] 初始化结束 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n", did)
				}
				statistics.cLock.Unlock()

				if debug == true {
					//rtdps, _ := base64.StdEncoding.DecodeString(string(rtdata["data"]))
					rtd, _ := rtdata["data"].(string)
					rtdps, _ := base64.StdEncoding.DecodeString(rtd)

					//fmt.Printf("实时数据raw： %v\n\n", rtdps)
					dps := parse(rtdps)
					// app/web 客户端展示数据点
					//prdId := rtdata["prdId"]
					prdId, _ := rtdata["prdId"].(string)
					prdDps := prdMap[prdId].Datapoints

					fmt.Printf("\n---------解析之后的实时数据-------------\n")
					for _, item := range dps {
						dpItem := GetDpItem(prdDps, item.DpId)
						if item.DpType == NUMB {
							precision, _ := dpItem.Resolution.Int64()
							dpVal := ConvNumb(item.DpVal, dpItem.Min, int(precision))
							fmt.Printf("\n数据点Id: %d\n数据点类型%d\n数据点值:%g\n", item.DpId, item.DpType, dpVal)
						} else if item.DpType == BOOL {
							dpVal := "false"
							if bytes2int(item.DpVal) == uint64(1) {
								dpVal = "true"
							}
							fmt.Printf("\n数据点Id: %d\n数据点类型%d\n数据点值:%s\n", item.DpId, item.DpType, dpVal)
						}
					}
					fmt.Printf("\n---------------------------------------\n")
				}
			}
		}
	}
}

//func int16(b []byte) uint16 { return uint16(b[1]) | uint16(b[0])<<8 }

func parse(data []byte) []Dp {
	var (
		idx    uint16
		dpId   uint16
		dpLen  uint16
		dpType uint16
		dpVal  []byte
	)

	dps := make([]Dp, 0)

	if len(data) == 0 || data[0] != 0x31 {
		fmt.Printf("data len: %d, 首字节不是 0x31: %d", len(data), data[0])
		return nil
	}

	for idx = 1; idx < uint16(len(data)); idx += dpLen {
		if data[idx]&0x80 != 0 {
			dpId = binary.BigEndian.Uint16(data[idx : idx+2])
			idx += 2
		} else {
			dpId = uint16(data[idx])
			idx += 1
		}

		dpType = uint16(data[idx])
		idx += 1

		if data[idx]&0x80 != 0 {
			dpLen = binary.BigEndian.Uint16(data[idx : idx+2])
			idx += 2
		} else {
			dpLen = uint16(data[idx])
			idx += 1
		}

		dpVal = data[idx : idx+dpLen]
		dp := Dp{DpId: dpId, DpType: dpType, DpVal: dpVal}
		dps = append(dps, dp)
	}

	return dps
}

func ConvNumb(numb []byte, min int, precision int) float64 {
	dpVal1 := bytes2int(numb)
	dpVal2 := (float64(dpVal1)/(math.Pow(float64(10), float64(precision))) + float64(min))
	return dpVal2
}

// bytes2int returns the int value it represents.
func bytes2int(data []byte) uint64 {
	n, val := len(data), uint64(0)
	if n > 8 {
		panic("data too long")
	}

	for i, b := range data {
		val += uint64(b) << uint64((n-i-1)*8)
	}
	return val
}

// int2bytes returns the byte array it represents.
func int2bytes(val uint64) []byte {
	data, j := make([]byte, 8), -1
	for i := 0; i < 8; i++ {
		shift := uint64((7 - i) * 8)
		data[i] = byte((val & (0xff << shift)) >> shift)

		if j == -1 && data[i] != 0 {
			j = i
		}
	}

	if j != -1 {
		return data[j:]
	}
	return data[:1]
}

func GetDpItem(dps []DpInfo, id uint16) DpInfo {
	for _, item := range dps {
		if item.DpId == int(id) {
			return item
		}
	}

	return DpInfo{}
}
