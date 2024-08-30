package main

import (
	"fmt"

	"github.com/mzky/utils/net"
)

func main() {
	fmt.Println(net.GetNetworkInfo())
	fmt.Println(net.GetRealAdapter())
}
