package _net

// worker 本质就是处理请求的 channel
type worker chan Request

type workerPool struct {
	// worker 池
	workerPool []worker
	// 每个 worker 最大缓存任务数量
	taskSize int
	// 任务处理回调
	request func(Request)
	// worker 池长度
	size int
}

// 初始化 worker 池
func newWorkerPool(poolSize int, taskSize int, request func(Request)) *workerPool {
	return &workerPool{
		workerPool: make([]worker, poolSize),
		taskSize:   taskSize,
		request:    request,
		size:       poolSize,
	}
}

// 启动 worker 池
func (_this *workerPool) start() {
	for i := 0; i < _this.size; i++ {
		_this.workerPool[i] = make(chan Request, _this.taskSize)

		go func(idx int) {
			for {
				select {
				case request := <-_this.workerPool[idx]:
					_this.request(request)
				}
			}
		}(i)
	}
}

// 添加任务
func (_this *workerPool) addTask(task Request) {
	// 每一条连接绑定一个 worker, 单条连接的消息处理是有序的
	_this.workerPool[task.Conn.GetConnID()%_this.size] <- task
}
