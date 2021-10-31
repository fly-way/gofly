package logs

var (
	// dir 日志路径
	// 若未执行 SetOutput, logs 默认输出 os.Stderr
	SetOutput func(dir string)

	// 是否输出debug日志
	SetDebug  func(bool)

	// 日志输出
	Debug     func(...interface{})
	Logic     func(...interface{})
	Money     func(...interface{})
	System    func(...interface{})
	Error     func(...interface{})
	Panic     func(...interface{})
	Stack     func(...interface{})
)
