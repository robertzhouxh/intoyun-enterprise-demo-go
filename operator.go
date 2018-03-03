package main

import (
	"intoyun-enterprise-demo-go/libs/define"
	"intoyun-enterprise-demo-go/libs/proto"
)

type Operator interface {
	Operate(p *proto.Proto) error
	Connect(p *proto.Proto) error
	Disconnect(p *proto.Proto) error
}

// realize the Operator interface
type DefaultOperator struct {
}

func (operator *DefaultOperator) Operate(p *proto.Proto) (err error) {
	if p.Operation == define.OP_HEARTBEAT {
		p.Operation = define.OP_HEARTBEAT_REPLY
		p.Body = nil
	}
	return
}

func (operator *DefaultOperator) Connect(p *proto.Proto) (err error) {
	// TODO
	return
}

func (operator *DefaultOperator) Disconnect(p *proto.Proto) (err error) {
	// TODO
	return
}
