package main

import (
	"bytes"
	"encoding/binary"
	"github.com/fly-way/gofly/_net"
	"github.com/fly-way/gofly/logs"
	"net"
	"time"
)

func server() {
	srv := _net.NewServer()
	defer srv.Stop()

	ws := srv.TcpServe()
	ws.InitWorkerPool(1, 256, func(request _net.Request) {
		request.Conn.SendMsg(1000, []byte("pong"))
		logs.Debug("server receive [id,data]:", request.ID, string(request.Data.([]byte)))
	})

	ws.Listen("127.0.0.1", 9999, "tcp4")

	srv.Start()
}

func client() {
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		logs.Panic(err)
		return
	}

	go func() {
		for {
			msg := pack(1, []byte("ping"))
			_, err := conn.Write(msg)
			if err != nil {
				logs.Panic(err)
			}
			time.Sleep(time.Second * 2)
		}
	}()

	for {
		// 先读取 head (dataLen + msgID) (uint32 + uint32)
		head := make([]byte, 8)
		_, err := conn.Read(head)
		if err != nil {
			logs.Panic(err)
		}

		// 再读取 body
		size, id := unPack(head)
		data := make([]byte, size)
		if size != 0 {
			_, err := conn.Read(data)
			if err != nil {
				logs.Panic(err)
			}
		}
		logs.Debug("client receive [id,data]:", id, string(data))
	}
}

func pack(id uint32, data []byte) []byte {
	buff := bytes.NewBuffer([]byte{})
	if err := binary.Write(buff, binary.BigEndian, uint32(len(data))); err != nil {
		return nil
	}
	if err := binary.Write(buff, binary.BigEndian, id); err != nil {
		return nil
	}
	if err := binary.Write(buff, binary.BigEndian, data); err != nil {
		return nil
	}
	return buff.Bytes()
}

func unPack(data []byte) (size uint32, id uint32) {
	var dataBuff = bytes.NewReader(data)
	if err := binary.Read(dataBuff, binary.BigEndian, &size); err != nil {
		return 0,0
	}
	if err := binary.Read(dataBuff, binary.BigEndian, &id); err != nil {
		return 0,0
	}
	return size, id
}

func main() {
	go server()
	time.Sleep(1*time.Second)
	go client()

	for {
		time.Sleep(1 * time.Second)
	}
}