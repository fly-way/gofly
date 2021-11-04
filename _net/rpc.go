package _net

import (
	"github.com/fly-way/gofly/logs"
	"github.com/fly-way/gofly/utils"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
)

type rpcServe struct {
	json   bool
	client map[string]bool
	iFunc  []interface{}
	state  int
	done   *chan bool
}

func newRpc(json bool) *rpcServe {
	return &rpcServe{
		json:   json,
		client: make(map[string]bool),
		iFunc:  make([]interface{}, 0)}
}


func (_this *rpcServe) RegRcv(rcv interface{}) {
	_this.iFunc = append(_this.iFunc, rcv)
}

func (_this *rpcServe) RegAllowIps(ips ...string) {
	for _, v := range ips {
		_this.client[v] = true
	}
}

func (_this *rpcServe) Listen(host string, port int, done *chan bool) {
	_this.done = done

	addr := host + ":" + utils.ToString(port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logs.Panic("rpc listen error:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logs.Panic("rpc accept error:", err)
		}

		vecAddr := strings.Split(conn.RemoteAddr().String(), ":")
		if len(vecAddr) != 2 {
			logs.Panic("rpc remote addr error, addr:", vecAddr)
		}

		if (len(_this.client) > 0) && (!_this.client[vecAddr[0]]) {
			logs.Error("rpc unknown client ip:", vecAddr[0])
			return
		}

		logs.System("rpc serve, rmt addr:", conn.RemoteAddr(), "local addr:", conn.LocalAddr())

		if _this.json {
			for _, v := range _this.iFunc {
				rpc.Register(v)
			}
			jsonrpc.ServeConn(conn)
		} else {
			p := rpc.NewServer()
			for _, v := range _this.iFunc {
				p.Register(v)
			}
			p.ServeConn(conn)
		}
	}
}

func (_this *rpcServe) CloseDone() {
	if _this.done != nil {
		close(*_this.done)
	}
}
