package ebuc

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
)

type TcpServer struct {
	acceptor Eventor
	loops    []*EventLoop
	wg       sync.WaitGroup
	connNum  int32

	// recall when new connection come
	Open    func(n int)
	Process func(in []byte) []byte
}

func NewTcpServer(loopNum int) *TcpServer {
	tcpServer := TcpServer{
		loops: make([]*EventLoop, loopNum),
	}

	for i := 0; i < loopNum; i++ {
		tcpServer.loops[i] = NewEventLoop()
	}

	// add acceptor eventor to first loop
	fd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}

	tcpServer.acceptor = Eventor{
		loop: tcpServer.loops[0],
		fd:   fd,
		cb:   tcpServer.accepted,
	}
	tcpServer.acceptor.EnableRead()

	tcpServer.loops[0].AddEvent(&tcpServer.acceptor)
	return &tcpServer
}

func (server *TcpServer) Start(addr string) {
	tcpAddr := syscall.SockaddrInet4{}
	tcpAddr.Port = 50001
	copy(tcpAddr.Addr[:], net.ParseIP(addr).To4())
	syscall.Bind(server.acceptor.fd, &tcpAddr)
	syscall.Listen(server.acceptor.fd, 10)
	fmt.Println("listening on 50001")

	server.wg.Add(len(server.loops))
	for _, loop := range server.loops {
		go loop.Loop()
	}
	server.wg.Wait()
}

func (server *TcpServer) accepted(event uint32, eventor *Eventor) {
	if (event & syscall.EPOLLIN) == 0 {
		fmt.Println("no a good event")
		return
	}

	// accept
	connFd, _, err := syscall.Accept(eventor.fd)
	if err != nil {
		return
	}

	// RoundRobin
	idx := int(atomic.LoadInt32(&server.connNum)) % len(server.loops)

	// new event
	connector := NewEventor(func(event uint32, eventor *Eventor) {
		if (event & syscall.EPOLLIN) != 0 {
			in := make([]byte, 0xffff)
			n, err := syscall.Read(eventor.fd, in)
			if n == 0 || err != nil {
				if err == syscall.EAGAIN {
					return
				}
				eventor.loop.RemoveEvent(eventor)
				eventor.loop = nil
				eventor.out = nil
				return
			}

			eventor.out = server.Process(in)
			eventor.EnableWrite()
			eventor.loop.UpdateEvent(eventor)
		}

		if (event & syscall.EPOLLOUT) != 0 {
			n, err := syscall.Write(eventor.fd, eventor.out)
			if err != nil {
				if err == syscall.EAGAIN {
					return
				}
				panic(err)
			}

			if n == len(eventor.out) {
				if cap(eventor.out) > 4096 {
					eventor.out = nil
				} else {
					eventor.out = eventor.out[:0]
				}
			} else {
				eventor.out = eventor.out[n:]
			}

			if len(eventor.out) == 0 {
				eventor.DisableWrite()
				eventor.loop.UpdateEvent(eventor)
			}
		}
	})
	connector.fd = connFd
	if err := syscall.SetNonblock(connector.fd, true); err != nil {
		panic(err)
	}

	connector.EnableRead()

	// add event
	server.loops[idx].AddEvent(&connector)
	atomic.AddInt32(&server.connNum, 1)

	server.Open(idx)
}
