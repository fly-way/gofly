package _mysql

func NewMysqlConn() IMysqlConn {
	return &mysqlConn{}
}

type IMysqlConn interface {
	// 设置连接池信息
	// idleConn 连接池最多保留的空闲连接数量
	//		(这些连接并不会因为空闲而导致关闭, 它们会在下次使用时打开, 通过牺牲内存来减少程序运行时重新建立mysql连接的开销)
	// openConn 连接池大小
	//		(偶尔需要大量连接, 但是 idleConn 数量的连接都处于繁忙状态时, 系统会额外建立一些临时连接, 这些临时连接+常驻连接的数量最大值
	//		不会超过 openConn, 并且在完成任务后会释放这些资源)
	SetConnPoolInfo(idleConn, openConn int)

	// 开启连接
	// ip port user pwd 登录信息
	// dbName 数据库名称
	// charset 编码格式, 如 utf8mb4
	// workSize 缓存的任务数量
	//		(可以理解为任务队列的长度, 处理任务的线程会一直从任务队列中取出任务, 并从连接池中随机取出一条连接, 交给mysql服务处理)
	StartConn(ip string, port int, user string, pwd string, dbName string, charset string, workSize int)

	// 关闭连接
	CloseConn()

	// 任意写 如:
	//	ExecChan("INSERT INTO tb_player_info VALUES(?,?,?,?,?)", 10086, "独孤求败", 100, 1000, "cn")
	//	ExecChan("Delete FROM tb_player_info WHERE uid = 10086")
	//	ExecChan("TRUNCATE TABLE tb_player_info;")
	ExecChan(query string, args ...interface{})

	// 查询单条数据 如:
	//	var playerData = player{} 需定义一个 player 结构体作为数据容器
	//	同步写法: <-QueryChan("SELECT * FROM tb_player_info where uid = 10086", &playerData)
	//	异步写法: goroutine 里执行
	QueryChan(query string, dest interface{}) chan bool

	// 查询多条数据 如:
	//	var playerSlice []player 需定义一个 player 结构的切片作为数据容器
	//	同步写法: <-QueryMoreChan("SELECT * FROM tb_player_info", &playerSlice)
	//	异步写法: goroutine 里执行
	QueryMoreChan(query string, dest interface{}) chan bool
}
