package main

import (
	"example.com/montyzhao/ebuc"
	"fmt"
)

func main() {
	tcpServer := ebuc.NewTcpServer(4)
	tcpServer.Open = func(n int) {
		fmt.Printf("connection %d connected", n)
	}
	tcpServer.Process = func(in []byte) []byte {
		return in
	}
	tcpServer.Start("127.0.0.1")

	fmt.Println("start loop")
}
