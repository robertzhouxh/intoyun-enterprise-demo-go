package main

import (
	"encoding/json"
	"errors"
	"intoyun-enterprise-demo-go/libs/define"
	"intoyun-enterprise-demo-go/libs/proto"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/thinkboy/log4go"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func InitWebsocket(addrs []string) (err error) {
	var (
		bind         string
		listener     *net.TCPListener
		addr         *net.TCPAddr
		httpServeMux = http.NewServeMux()
		server       *http.Server
	)

	httpServeMux.HandleFunc("/sub", ServeWebSocket)

	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		server = &http.Server{Handler: httpServeMux}
		if Debug {
			log.Debug("start websocket listen: \"%s\"", bind)
		}
		go func(host string) {
			if err = server.Serve(listener); err != nil {
				log.Error("server.Serve(\"%s\") error(%v)", host, err)
				panic(err)
			}
		}(bind)
	}
	return
}

func ServeWebSocket(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Error("Websocket Upgrade error(%v), userAgent(%s)", err, req.UserAgent())
		return
	}
	defer ws.Close()
	var (
		lAddr = ws.LocalAddr()
		rAddr = ws.RemoteAddr()
	)
	log.Debug("start websocket serve \"%s\" with \"%s\"", lAddr, rAddr)
	DefaultServer.serveWebsocket(ws)
}

// Reader
func (server *Server) serveWebsocket(conn *websocket.Conn) {
	var (
		err error
		p   *proto.Proto
		key string // clientID or sessionId
		//hb  time.Duration
		ch = NewChannel(Conf.CliProto, Conf.SvrProto)
	)

	if p, err = ch.CliProto.Set(); err == nil {
		if key, _, err = server.authWebsocket(conn, p); err == nil {
			err = Buckets.Put(key, ch)
		} else {
			conn.Close()
			log.Error("handshake failed error(%v)", err)
			return
		}
	}

	go server.dispatchWebsocket(conn, ch)

	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		if err = p.ReadWebsocket(conn); err != nil {
			break
		}

		// process message, and then operate the p
		if err = server.operator.Operate(p); err != nil {
			break
		}

		ch.CliProto.SetAdv()

		// notify the Writer
		ch.Signal()
	}

	log.Error("key: %s server websocket failed error(%v)", key, err)
	conn.Close()
	ch.Close()
	Buckets.Del(key)

	return
}

// Writer
func (server *Server) dispatchWebsocket(conn *websocket.Conn, ch *Channel) {
	var (
		err error
		p   *proto.Proto
	)
	for {
		p = ch.Ready()
		switch p {
		case proto.ProtoFinish:
			goto failed
		case proto.ProtoReady:
			for {
				if p, err = ch.CliProto.Get(); err != nil {
					err = nil // must be empty error
					break
				}
				if err = p.WriteWebsocket(conn); err != nil {
					goto failed
				}
				p.Body = nil // avoid memory leak
				ch.CliProto.GetAdv()
			}
		default:
			// not from reader but from kafka push
			log.Error("push msg here >>>>>>>>>>>>>>>>>: %v", p)
			if err = p.WriteWebsocket(conn); err != nil {
				log.Error("p.WriteWebsocket Err:>>>>>>>>>>>>>>>>>: %v", err)
				goto failed
			}
		}

	}

failed:
	if err != nil {
	}
	conn.Close()

	// must ensure all channel message discard, for reader won't blocking Signal
	for {
		if p == proto.ProtoFinish {
			break
		}
		p = ch.Ready()
	}
	return
}

func (server *Server) authWebsocket(conn *websocket.Conn, p *proto.Proto) (key string, heartbeat time.Duration, err error) {
	if err = p.ReadWebsocket(conn); err != nil {
		log.Error("p.ReadWebsocket err: %v!!!!!!", err)
		return
	}
	if p.Operation != define.OP_AUTH {
		err = errors.New("Invalid Operation")
		return
	}
	data := map[string]string{}
	json.Unmarshal(p.Body, &data)
	log.Debug("authwebsocket data %v", data)
	key = data["key"]
	heartbeat = 300

	p.Body = nil
	p.Operation = define.OP_AUTH_REPLY
	err = p.WriteWebsocket(conn)

	return
}
