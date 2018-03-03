package main

import (
	"intoyun-enterprise-demo-go/libs/proto"
)

type Channel struct {
	CliProto Ring
	signal   chan *proto.Proto
}

func NewChannel(cli, svr int) *Channel {
	c := new(Channel)
	c.CliProto.Init(cli)
	c.signal = make(chan *proto.Proto, svr)
	return c
}

func (c *Channel) Push(p *proto.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
	}
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() *proto.Proto {
	return <-c.signal
}

// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {
	c.signal <- proto.ProtoReady
}

// Close close the channel.
func (c *Channel) Close() {
	c.signal <- proto.ProtoFinish
}
