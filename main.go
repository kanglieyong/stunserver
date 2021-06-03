package main

import (
	"fmt"

	"zspace.cn/stunserver/servers"
)

const APP_VERSION = "20210603a"

func main() {
	fmt.Println("stunserver started!")
	servers.Instance().Start()
}
