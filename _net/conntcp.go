package _net

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gofly/logs"
	"gofly/utils/encrypt"
	"io"
	"net"
	"sync"
)

type connTcp struct {
	id       int
	serve    *socket
	conn     *net.TCPConn
	attr     sync.Map
	msgChan  chan []byte
	headPool []byte
	exit     chan bool
}

type msgTcp struct {
	len  uint32
	id   uint32
	data interface{}
}

func newConnTcp(id int, serve *socket, conn *net.TCPConn) IConn {
	return &connTcp{
		id:       id,
		serve:    serve,
		conn:     conn,
		msgChan:  make(chan []byte, 4096),
		headPool: make([]byte, 8 /*len uint32 + id uint32*/),
		exit:     make(chan bool),
	}
}

func (_this *connTcp) Start() {
	if _this.serve.connStart != nil {
		_this.serve.connStart(_this)
	}

	go _this.reader()
	go _this.writer()
}

func (_this *connTcp) Stop() {
	if _this.serve.connStop != nil {
		_this.serve.connStop(_this)
	}

	close(_this.exit)
	close(_this.msgChan)
}

func (_this *connTcp) GetConnID() int {
	return _this.id
}

func (_this *connTcp) GetRemoteAddr() net.Addr {
	return _this.conn.RemoteAddr()
}

func (_this *connTcp) WriteMsg(msg []byte) {
	select {
	case <-_this.exit:
		return
	default:
		_this.msgChan <- msg
	}
}

func (_this *connTcp) SendMsg(id int, data interface{}) {
	v, ok := data.([]byte)
	if !ok {
		logs.Error("msg data not []byte, data:", data)
		return
	}

	msgTcp := &msgTcp{len: uint32(len(v)), id: uint32(id), data: v}
	if _this.serve.key != "" {
		if msg, err := encrypt.AesEncrypt(v, []byte(_this.serve.key), []byte(_this.serve.iv)); err != nil {
			logs.Error("sendMsg aes encrypt err:", err)
			return
		} else {
			msgTcp.data = msg
		}
	}

	msg, err := _this.pack(msgTcp)
	if err != nil {
		logs.Error("pack err:", err, "msg id:", id)
		return
	}

	_this.WriteMsg(msg)
}

func (_this *connTcp) QueryAttr() *sync.Map {
	return &_this.attr
}

func (_this *connTcp) reader() {
	defer _this.Stop()

	for {
		if _, err := io.ReadFull(_this.conn, _this.headPool); err != nil {
			return
		}

		msg, err := _this.unPack(_this.headPool)
		if err != nil {
			logs.Error("unpack err:", err)
		}

		var data []byte
		if msg.len > 0 {
			data = make([]byte, msg.len)
			if _, err := io.ReadFull(_this.conn, data); err != nil {
				logs.Error("read msg data err:", err)
				return
			}
		}

		if _this.serve.key != "" {
			if data, err = encrypt.AesDecrypt(data, []byte(_this.serve.key), []byte(_this.serve.iv)); err != nil {
				logs.Error("Decrypt err, close connection ... data:", string(data), "err:", err)
				return
			}
		}

		go _this.serve.workers.addTask(Request{Conn: _this, ID: msg.id, Data: data})
	}
}

func (_this *connTcp) writer() {
	for {
		select {
		case <-_this.exit:
			return
		default:
			if _, err := _this.conn.Write(<-_this.msgChan); err != nil {
				logs.Error("send data err:", err, " conn writer exit")
				return
			}
		}
	}
}

func (_this *connTcp) pack(msg *msgTcp) ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})

	if err := binary.Write(buff, _this.serve.byteOrder, msg.len); err != nil {
		return nil, err
	}

	if err := binary.Write(buff, _this.serve.byteOrder, msg.id); err != nil {
		return nil, err
	}

	if err := binary.Write(buff, _this.serve.byteOrder, msg.data); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (_this *connTcp) unPack(data []byte) (*msgTcp, error) {
	var (
		msg      = msgTcp{}
		dataBuff = bytes.NewReader(data)
	)

	if err := binary.Read(dataBuff, _this.serve.byteOrder, &msg.len); err != nil {
		return nil, err
	}

	if _this.serve.packetMaxSize > 0 && int(msg.len) > _this.serve.packetMaxSize {
		return nil, errors.New("msg data long")
	}

	if err := binary.Read(dataBuff, _this.serve.byteOrder, &msg.id); err != nil {
		return nil, err
	}

	return &msg, nil
}
