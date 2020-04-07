package ebuc

import (
	"fmt"
	"syscall"
)

type EventLoop struct {
	pfd      int
	eventors map[int32]Eventor
}

func NewEventLoop() *EventLoop {
	epollfd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	eventLoop := EventLoop{
		eventors: make(map[int32]Eventor),
		pfd:      epollfd,
	}
	return &eventLoop
}

func (*EventLoop) Post(task func()) {
}

func (*EventLoop) Stop() {
}

func (loop *EventLoop) AddEvent(e Eventor) {
	loop.eventors[int32(e.fd)] = e
	if err := syscall.EpollCtl(loop.pfd, syscall.EPOLL_CTL_ADD, e.fd,
		&syscall.EpollEvent{Fd: int32(e.fd), Events: e.event}); err != nil {
		panic(err)
	}
}

func (loop *EventLoop) UpdateEvent(e Eventor) {
	loop.eventors[int32(e.fd)] = e
	if err := syscall.EpollCtl(loop.pfd, syscall.EPOLL_CTL_MOD, e.fd,
		&syscall.EpollEvent{Fd: int32(e.fd), Events: e.event}); err != nil {
		panic(err)
	}
}

func (loop *EventLoop) RemoveEvent(e Eventor) {
	if err := syscall.EpollCtl(loop.pfd, syscall.EPOLL_CTL_DEL, e.fd,
		&syscall.EpollEvent{Fd: int32(e.fd), Events: e.event}); err != nil {
		panic(err)
	}
}

func (loop *EventLoop) Pool() {
	events := make([]syscall.EpollEvent, 64)
	n, err := syscall.EpollWait(loop.pfd, events, -1)
	if err != nil && err != syscall.EINTR {
		panic(err)
	}

	fmt.Println("new events")
	for i := 0; i < n; i++ {
		ev := events[i]
		eventor, ok := loop.eventors[ev.Fd]
		if !ok {
			continue
		}

		if (eventor.event & ev.Events) == 0 {
			continue
		}

		eventor.cb(ev.Events, eventor)
	}
}

func (loop *EventLoop) Loop() {
	for {
		loop.Pool()
	}
}

// time event
type TimeId int

func (*EventLoop) RunAfter(task func(), time int64) TimeId {
	return 1
}

func (*EventLoop) RunPeriodic(task func(), time int64) TimeId {
	return 1
}

func (*EventLoop) CancelTimer(timeId TimeId) {
}
