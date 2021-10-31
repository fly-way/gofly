package _net

import (
	"github.com/fly-way/gofly/logs"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"
)


type server struct {
	profPort      int
	signals       []os.Signal
	signalCall    func(os.Signal)
}

func (_this *server) TcpServe() ISocket {
	return newSocket("tcp")
}

func (_this *server) WebsocketServe() ISocket {
	return newSocket("websocket")
}


func (_this *server) RpcServe() IRpc {
	return newRpc(false)
}


func (_this *server) RpcJsonServe() IRpc {
	return newRpc(true)
}

func (_this *server) RegProfListen(port int) {
	_this.profPort = port
}

func (_this *server) RegSignal(callback func(os.Signal), sig ...os.Signal) {
	_this.signalCall = callback
	_this.signals = sig
}


func (_this *server) Start() {
	logs.System("server start!")

	// prof 性能分析
	if _this.profPort != 0 {
		// 开启对锁调用的跟踪
		runtime.SetMutexProfileFraction(1)
		// 开启对阻塞操作的跟踪
		runtime.SetBlockProfileRate(1)
		addr := "127.0.0.1" + ":" + strconv.Itoa(_this.profPort)
		go func() {
			http.ListenAndServe(addr, nil)
		}()
	}

	// 信号处理
	if _this.signalCall != nil {
		go func() {
			c := make(chan os.Signal)
			signal.Notify(c, _this.signals...)
			for sig := range c {
				_this.signalCall(sig)
				logs.System("Signal:", sig)
			}
		}()
	}

	// 阻塞主线程
	for {
		time.Sleep(60 * time.Second)
	}
}

func (_this *server) Stop() {
	logs.System("server stop!")
}

