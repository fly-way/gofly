package logs

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	logDebug = iota
	logLogic
	logMoney
	logSystem
	logError
	logPanic
)

var logName = []string{
	"debug",
	"logic",
	"money",
	"system",
	"error",
	"panic",
}

var (
	logs  []*logger
	debug bool
)

func init() {
	logs = make([]*logger, len(logName))
	debug = true
	for k, v := range logName {
		logs[k] = new(lStdFlags, "[" + strings.ToUpper(v) + "]", 2)
	}

	SetOutput = func(dir string) {
		for k, v := range logName {
			logs[k].setOutput(dir, v)
		}
	}

	Debug = func(v ...interface{}) {
		if debug {
			logs[logDebug].output(fmt.Sprintln(v...))
		}
	}

	Logic = func(v ...interface{}) {
		logs[logLogic].output(fmt.Sprintln(v...))
	}

	Money = func(v ...interface{}) {
		logs[logMoney].output(fmt.Sprintln(v...))
	}

	System = func(v ...interface{}) {
		logs[logSystem].output(fmt.Sprintln(v...))
	}

	Error = func(v ...interface{}) {
		logs[logError].output(fmt.Sprintln(v...))
	}

	Panic = func(v ...interface{}) {
		s := fmt.Sprintln(v...)
		logs[logPanic].output(s)
		panic(s)
	}

	Stack = func(v ...interface{}) {
		s := fmt.Sprint(v...)
		s += "\n"
		buf := make([]byte, 1024*1024)
		n := runtime.Stack(buf, true) //得到当前堆栈信息
		s += string(buf[:n])
		s += "\n"
		logs[logPanic].output(s)
	}
}



