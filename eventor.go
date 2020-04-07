package ebuc

import (
	"syscall"
)

type Eventor struct {
	fd    int
	event uint32
	cb    func(uint32, Eventor)
	loop  *EventLoop
}

func (e *Eventor) EnableRead() {
	e.event = e.event | syscall.EPOLLIN | -syscall.EPOLLET
}

func (e *Eventor) EnableWrite() {
	e.event = e.event | syscall.EPOLLOUT | -syscall.EPOLLET
}

func (e *Eventor) DisableRead() {
	// e.event = e.event & ^syscall.EPOLLIN
}

func (e *Eventor) DisableWrite() {
	// e.event = e.event & ^syscall.EPOLLOUT
}

func NewEventor(callBack func(uint32, Eventor), loop *EventLoop) Eventor {
	return Eventor{cb: callBack, loop: loop}
}
