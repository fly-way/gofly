package _net

import (
	"net"
	"os"
	"sync"
)

func NewServer() IServer {
	return &server{}
}

type IServer interface {
	// 初始化tcp服务对象
	TcpServe() ISocket

	// 初始化websocket服务对象
	WebsocketServe() ISocket

	// 初始化rpc服务对象
	RpcServe() IRpc

	// 初始化json rpc 服务对象
	RpcJsonServe() IRpc

	// 注册 prof 性能分析监听
	// port 端口
	RegProfListen(port int)

	// 注册系统信号处理
	// callback 处理系统信号的 callback
	// sig 系统信号集合
	RegSignal(callback func(os.Signal), sig ...os.Signal)

	// 启动服务
	Start()

	// 停止服务
	Stop()
}

type ISocket interface {
	// 设置 aes 加密 key 和 向量 iv, key 的 16 位, 24, 32 分别对应 AES-128, AES-192, AES-256
	// 默认为不加密
	AesEncrypt(key string, iv string)

	// 设置与客户端建立连接时的回调函数
	SetConnStartCall(func(IConn))

	// 设置与客户端断开连接时的回调函数
	SetConnStopCall(func(IConn))

	// 设置消息包最大长度
	// 默认为4096
	SetPacketMaxSize(size int)

	// 设置字节顺序
	// 字节数据可以存放在低地址处, 也可以存放在高地址处, 若双端出现字节数据高低位相反, 就要考虑到字节顺序问题
	// 默认为big endian
	SetByteOrder(bigOrder bool)

	// 初始化工作池
	// poolSize 工作池的 worker 数量
	// taskSize 每个 worker 最大缓存任务数量
	// request 设置客户端发起请求时的回调函数
	// 工作池最大任务缓存数量 =  poolSize * taskSize
	InitWorkerPool(poolSize int, taskSize int, request func(Request))

	// 开启端口监听, host ip地址, port 端口 param 扩展参数
	// tcp socket 下, param 表示 tcp 版本("tcp", "tcp4", "tcp6"), 为空表示"tcp"
	// websocket 下, param 表示 pattern 模式, 如 "/ws"
	Listen(host string, port int, param string)
}

type IRpc interface {
	// 注册节点标识
	// id 节点id typ 节点类型
	// 用于区分rpc集群中各个节点
	RegNode(id int, typ string)

	// 注册接口
	// 可注册多个接口
	RegRcv(rcv interface{})

	// 注册允许建立连接的客户端ip集合
	// 未注册时表示允许所有连接
	// 注册后只允许ips集合里的ip进行rpc连接
	// 重复调用会覆盖之前注册的ips
	RegAllowIps(ips map[string]bool)

	// 开启端口监听, host ip地址, port 端口
	Listen(host string, port int)
}

type IConn interface {
	// 启动连接
	Start()
	// 停止连接
	Stop()
	// 获取连接ID
	GetConnID() int
	// 获取连接地址
	GetRemoteAddr() net.Addr
	// 发送消息
	WriteMsg([]byte)
	// 发送消息
	SendMsg(id int, data interface{})
	// 获取属性
	QueryAttr() *sync.Map
}

type Request struct {
	Conn IConn
	ID   uint32
	Data interface{}
}
