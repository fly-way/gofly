package main

import (
	"github.com/fly-way/gofly/_net"
	"github.com/fly-way/gofly/logs"
	"net/rpc"
	"time"
)

type MathUtil struct {
}

func (mu *MathUtil) Square(arg int, ret *int) error {
	*ret = arg * arg
	logs.Debug(*ret)
	return nil
}

func server() {
	srv := _net.NewServer()
	defer srv.Stop()

	rpcServe := srv.RpcServe()
	rpcServe.RegRcv(new(MathUtil))
	rpcServe.Listen("127.0.0.1", 9999, nil)

	srv.Start()
}

func client() {
	client, err := rpc.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		panic(err.Error())
	}
	defer client.Close()

	var ret = 0
	for i := 0; i <= 20; i++ {
		client.Call("MathUtil.Square", i, &ret)
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