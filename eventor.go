package ebuc

import (
	"syscall"
)

var (
	EPOLLET uint32 = 0x80000000
)

type Eventor struct {
	fd    int
	event uint32
	cb    func(uint32, *Eventor)
	loop  *EventLoop
	out   []byte
	in    []byte
}

func (e *Eventor) EnableRead() {
	e.event = e.event | syscall.EPOLLIN | EPOLLET
}

func (e *Eventor) EnableWrite() {
	e.event = e.event | syscall.EPOLLOUT | EPOLLET
}

func (e *Eventor) DisableRead() {
	e.event = e.event &^ syscall.EPOLLIN
}

func (e *Eventor) DisableWrite() {
	e.event = e.event &^ syscall.EPOLLOUT
}

func (e *Eventor) SetEventLoop(loop *EventLoop) {
	e.loop = loop
}

func NewEventor(callBack func(uint32, *Eventor)) Eventor {
	return Eventor{cb: callBack}
}
