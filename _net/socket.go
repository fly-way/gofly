package _net

import (
	"encoding/binary"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/fly-way/gofly/logs"
	"net"
	"net/http"
	"strconv"
)

type socket struct {
	network       string
	key, iv       string
	packetMaxSize int
	byteOrder     binary.ByteOrder
	connStart     func(IConn)
	connStop      func(IConn)
	workers       *workerPool
}

func newSocket(network string) *socket {
	logs.System("socket init:", network)

	return &socket{
		network:       network,
		packetMaxSize: 4096,
		byteOrder:     binary.BigEndian,
	}
}

func (_this *socket) AesEncrypt(key string, iv string) {
	_this.key = key
	_this.iv = iv
}

func (_this *socket) SetConnStartCall(connStart func(IConn)) {
	_this.connStart = connStart

}

func (_this *socket) SetConnStopCall(connStop func(IConn)) {
	_this.connStop = connStop
}

func (_this *socket) InitWorkerPool(poolSize int, taskSize int, request func(Request)) {
	_this.workers = newWorkerPool(poolSize, taskSize, request)
}

func (_this *socket) SetPacketMaxSize(size int) {
	_this.packetMaxSize = size
}

func (_this *socket) SetByteOrder(bigOrder bool) {
	if bigOrder {
		_this.byteOrder = binary.BigEndian
	} else {
		_this.byteOrder = binary.LittleEndian
	}
}

func (_this *socket) Listen(host string, port int, param string) {
	if _this.workers == nil {
		_this.workers = newWorkerPool(1, 256, func(r Request){
			logs.Debug("request[conn,msgID,msgData]:", r.Conn.GetConnID(), r.ID, r.Data)
		})
	}
	_this.workers.start()

	switch _this.network {
	case "tcp":
		switch param {
		case "tcp", "tcp4", "tcp6", "":
			logs.System("tcp listen, host:", host, "port:", port, "version:", param)

			addr, err := net.ResolveTCPAddr(param, fmt.Sprintf("%s:%d", host, port))
			if err != nil {
				logs.Panic("Resolve tcp addr err: ", err)
				return
			}

			listener, err := net.ListenTCP(param, addr)
			if err != nil {
				logs.Panic("listen", param, "err", err)
				return
			}

			for {
				tcpConn, err := listener.AcceptTCP()
				if err != nil {
					logs.Error("accept err ", err)
					continue
				}

				go func() {
					conn := newConnTcp(1, _this, tcpConn)
					conn.Start()
				}()
			}
		default:
			logs.Panic(fmt.Sprintf("tcp socket unknown param: %s", param))
			return
		}
	case "websocket":
		logs.System("websocket listen, host:", host, "port:", port, "pattern:", param)

		upGrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc(param, func(writer http.ResponseWriter, request *http.Request) {
				var (
					wsConn *websocket.Conn
					err    error
				)

				if wsConn, err = upGrader.Upgrade(writer, request, nil); err != nil {
					return
				}

				go func() {
					conn := newConnWebsocket(1, _this, wsConn)
					conn.Start()
				}()
			})

			addr := host + ":" + strconv.Itoa(port)
			if err := http.ListenAndServe(addr, mux); err != nil {
				panic(err)
			}
		}()

	default:
		logs.Panic(fmt.Sprintf("unknown network: %s", _this.network))
	}
}
