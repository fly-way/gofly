package main

import (
	"fmt"
	_ "net/http/pprof" // 必须 import 这个包
)

func test(sz []int) {
	sz = append(sz, 1)
	fmt.Println(sz)

}

func main() {
	a := make([]int, 0)
	a = append(a, 1)
	a = append(a, 2)
	a = append(a, 3)
	a = append(a, 4)
	a = append(a, 5)

	test(a)
	fmt.Println(a)
}
