package main

import (
	"example.com/montyzhao/ebuc"
	"fmt"
)

func main() {
	tcpServer := ebuc.NewTcpServer(4)
	tcpServer.Open = func(n int) {
		fmt.Printf("connection %d connected\n", n)
	}
	tcpServer.Process = func(in []byte, clientId int32, loop *ebuc.EventLoop) ([]byte, int) {
		loop.RunAfter(func(event uint32, eventor *ebuc.Eventor) {
			loop.NotifyClient(clientId, []byte("time is up, bro\n"))
		}, 3000)
		return in, len(in)
	}
	tcpServer.Start("127.0.0.1")

	fmt.Println("start loop")
}
