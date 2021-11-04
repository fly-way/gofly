package main

import (
	"github.com/fly-way/gofly/_net"
	"github.com/fly-way/gofly/logs"
	"net"
	"net/rpc/jsonrpc"
	"time"
)

type JsonMathUtil struct {
}

func (mu *JsonMathUtil) Square(arg int, ret *int) error {
	*ret = arg * arg
	logs.Debug(*ret)
	return nil
}

func jsonServer() {
	srv := _net.NewServer()
	defer srv.Stop()

	rpcServe := srv.RpcJsonServe()
	rpcServe.RegRcv(new(JsonMathUtil))
	rpcServe.Listen("127.0.0.1", 9999, nil)

	srv.Start()
}

func jsonClient() {
	client, err := net.DialTimeout("tcp", "127.0.0.1:9999", 1000*1000*1000*30) // 30秒超时时间
	if err != nil {
		panic(err.Error())
	}
	defer client.Close()

	clientRpc := jsonrpc.NewClient(client)

	var ret = 0
	for i := 0; i <= 20; i++ {
		clientRpc.Call("JsonMathUtil.Square", i, &ret)
	}
}

func main() {
	go jsonServer()
	time.Sleep(1*time.Second)
	go jsonClient()

	for {
		time.Sleep(1 * time.Second)
	}
}
