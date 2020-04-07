package main

import (
	"example.com/montyzhao/ebuc"
	"fmt"
)

func main() {
	tcpServer := ebuc.NewTcpServer(4)
	tcpServer.Start("127.0.0.1")

	fmt.Println("start loop")
}
