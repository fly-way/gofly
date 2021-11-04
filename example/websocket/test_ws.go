package main

import (
	"bytes"
	"encoding/binary"
	"github.com/fly-way/gofly/_net"
	"github.com/fly-way/gofly/logs"
	"github.com/gorilla/websocket"
	"time"
)

var content = []string{
	"Hello!",
	"Hi!",
	"Nice to see you!",
}

func server() {
	srv := _net.NewServer()
	defer srv.Stop()

	ws := srv.WebsocketServe()

	ws.InitWorkerPool(1, 256, func(request _net.Request) {
		request.Conn.SendMsg(1000, []byte("pong"))
		logs.Debug("server receive [id,data]:", request.ID, string(request.Data.([]byte)))
	})
	ws.Listen("127.0.0.1", 9999, "/ws")

	srv.Start()
}

func pack(id uint32, data interface{}) []byte {
	buff := bytes.NewBuffer([]byte{})
	if binary.Write(buff, binary.BigEndian, id) != nil {
		return nil
	}

	if binary.Write(buff, binary.BigEndian, data) != nil {
		return nil
	}

	return buff.Bytes()
}

func client() {
	url := "ws://127.0.0.1:9999/ws" //服务器地址
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		logs.Panic(err)
	}

	go func() {
		for {
			msg := pack(1, []byte("ping"))
			err := ws.WriteMessage(websocket.BinaryMessage, msg)
			if err != nil {
				logs.Panic(err)
			}
			time.Sleep(time.Second * 2)
		}
	}()

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			logs.Panic(err)
		}

		logs.Debug("client receive [id,data]:", binary.BigEndian.Uint32(data[:4]), string(data[4:]))
	}
}

func main() {
	go server()
	time.Sleep(1*time.Second)
	go client()

	for {
		time.Sleep(1 * time.Second)
	}
}

