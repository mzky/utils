package main

import (
	"fmt"

	"utils/net"
)

func main() {
	fmt.Println(net.GetNetworkInfo())
	fmt.Println(net.GetRealAdapter())
}
