package main

import (
	"fmt"

	"zspace.cn/stunserver/servers"
)

func main() {
	fmt.Println("stunserver started!")
	servers.Instance().Start()
}
