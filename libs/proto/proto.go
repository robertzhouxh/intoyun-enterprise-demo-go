package proto

import (
	"errors"
	"fmt"
	"intoyun-enterprise-demo-go/libs/define"

	"github.com/gorilla/websocket"
)

var (
	emptyProto    = Proto{}
	emptyJSONBody = []byte("{}")

	ErrProtoPackLen   = errors.New("default server codec pack length error")
	ErrProtoHeaderLen = errors.New("default server codec header length error")
)

var (
	ProtoReady  = &Proto{Operation: define.OP_PROTO_READY}
	ProtoFinish = &Proto{Operation: define.OP_PROTO_FINISH}
)

type Proto struct {
	Operation int32  `json:"op"`   // operation for request
	Body      []byte `json:"body"` // binary body bytes(json.RawMessage is []byte)
	//Body      json.RawMessage `json:"body"` // binary body bytes(json.RawMessage is []byte)
	//Body *json.RawMessage `json:"body"` // binary body bytes(json.RawMessage is []byte)
}

func (p *Proto) Reset() {
	*p = emptyProto
}

func (p *Proto) String() string {
	return fmt.Sprintf("\n-------- proto --------\nop: %d\nbody: %v\n-----------------------", p.Operation, p.Body)
}

func (p *Proto) ReadWebsocket(conn *websocket.Conn) (err error) {
	err = conn.ReadJSON(p)
	return
}

func (p *Proto) WriteWebsocket(conn *websocket.Conn) (err error) {
	if p.Body == nil {
		p.Body = emptyJSONBody
	}
	err = conn.WriteJSON(p)
	return
}

//func (p *Proto) WriteWebsocket(wr *websocket.Conn) (err error) {
//	if p.Body == nil {
//		p.Body = emptyJSONBody
//	}
//	err = wr.WriteJSON([]*Proto{p})
//	return
//}

//======================================================
type KafkaMsg struct {
	Code int32  `json:"code"`
	Ts   int32  `json:"ts"`
	Sign string `json:"sign"`
	//Body json.RawMessage `json:"body"`
	Body string `json:"body"`
}
