package main

import (
	"github.com/fly-way/gofly/logs"
	"time"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			for {
				logs.Error("5555")
				time.Sleep(1*time.Second)
			}
		}
	}()

	// 这里未设置 log 路径, 因此还是 os.Stderr 输出
	logs.Logic("0000")
	logs.Debug("1111")
	logs.System("2222")

	// 这里设置了 log 路径, 后续会输出在 ./log 目录下
	logs.SetOutput("./log")
	logs.Money("3333")
	logs.Debug("1111")
	logs.Debug("2222")

	logs.SetDebug(false)
	logs.Debug("3333")

	logs.Panic("4444")
}
