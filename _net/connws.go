package _net

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/fly-way/gofly/logs"
	"github.com/fly-way/gofly/utils/encrypt"
	"github.com/gorilla/websocket"
	"net"
	"sync"
)

type connWebsocket struct {
	id      int
	serve   *socket
	conn    *websocket.Conn
	attr    sync.Map
	msgChan chan []byte
	exit    chan bool
}

type msgWs struct {
	id   uint32
	data interface{}
}

func newConnWebsocket(id int, serve *socket, conn *websocket.Conn) IConn {
	return &connWebsocket{
		id:      id,
		serve:   serve,
		conn:    conn,
		msgChan: make(chan []byte, 4096),
		exit:    make(chan bool),
	}
}

func (_this *connWebsocket) Start() {
	if _this.serve.connStart != nil {
		_this.serve.connStart(_this)
	}

	go _this.reader()
	go _this.writer()
}

func (_this *connWebsocket) Stop() {
	if _this.serve.connStop != nil {
		_this.serve.connStop(_this)
	}

	close(_this.exit)
	close(_this.msgChan)
}

func (_this *connWebsocket) GetConnID() int {
	return _this.id
}

func (_this *connWebsocket) GetRemoteAddr() net.Addr {
	return _this.conn.RemoteAddr()
}

func (_this *connWebsocket) WriteMsg(msg []byte) {
	select {
	case <-_this.exit:
		return
	default:
		_this.msgChan <- msg
	}
}

func (_this *connWebsocket) SendMsg(id int, data interface{}) {
	msg, err := _this.pack(&msgWs{uint32(id), data})
	if err != nil {
		logs.Error("pack err:", err, "msg id:", id)
		return
	}

	if _this.serve.key != "" {
		if msg, err = encrypt.AesEncrypt(msg, []byte(_this.serve.key), []byte(_this.serve.iv)); err != nil {
			logs.Error("sendMsg aes encrypt err:", err)
			return
		}
	}

	_this.WriteMsg(msg)
}

func (_this *connWebsocket) QueryAttr() *sync.Map {
	return &_this.attr
}

func (_this *connWebsocket) reader() {
	defer _this.Stop()

	var (
		data []byte
		err  error
	)

	for {
		if _, data, err = _this.conn.ReadMessage(); err != nil {
			if netErr, ok := err.(net.Error); ok {
				if !netErr.Timeout() {
					logs.Debug("ReadMessage error:", err)
				}
			}
			return
		}

		if _this.serve.key != "" {
			if data, err = encrypt.AesDecrypt(data, []byte(_this.serve.key), []byte(_this.serve.iv)); err != nil {
				logs.Error("Decrypt err, close connection ... data:", string(data), "err:", err)
				return
			}
		}

		var msg *msgWs
		if msg, err = _this.unPack(data); err != nil {
			logs.Error("Unpack err:", err)
			return
		}

		go _this.serve.workers.addTask(Request{Conn: _this, ID: msg.id, Data: msg.data})
	}
}

func (_this *connWebsocket) writer() {
	for {
		select {
		case <-_this.exit:
			return
		default:
			if err := _this.conn.WriteMessage(websocket.BinaryMessage, <-_this.msgChan); err != nil {
				return
			}
		}
	}
}

func (_this *connWebsocket) pack(msg *msgWs) ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})

	if err := binary.Write(buff, _this.serve.byteOrder, msg.id); err != nil {
		return nil, err
	}

	if err := binary.Write(buff, _this.serve.byteOrder, msg.data); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (_this *connWebsocket) unPack(data []byte) (*msgWs, error) {
	size := len(data)
	if size < 4 {
		return nil, errors.New("msg data short")
	}

	if _this.serve.packetMaxSize > 0 && size > _this.serve.packetMaxSize {
		return nil, errors.New("msg data long")
	}

	return &msgWs{_this.serve.byteOrder.Uint32(data[:4]), data[4:]}, nil
}
