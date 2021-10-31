package _net

type rpcServe struct {
	json bool
}

func newRpc(json bool) *rpcServe {
	return &rpcServe{json}
}

func (_this *rpcServe) RegNode(id int, typ string) {

}

func (_this *rpcServe) RegRcv(rcv interface{}) {

}


func (_this *rpcServe) RegAllowIps(ips map[string]bool) {

}


func (_this *rpcServe) Listen(host string, port int) {

}

