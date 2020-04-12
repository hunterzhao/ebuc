package ebuc

/*
#include <errno.h>
#include <stdint.h>
#include <sys/timerfd.h>
#include <stdio.h>
int CreateTimerfd(int msec) {
	int tfd = timerfd_create(CLOCK_MONOTONIC, 0);
    if (tfd == -1) {
		printf("timerfd_create() failed: errno=%d\n", errno);
		return -1;
	}

	struct itimerspec ts;
	ts.it_interval.tv_sec = 0;
	ts.it_interval.tv_nsec = 0;
	ts.it_value.tv_sec = msec / 1000;
    ts.it_value.tv_nsec = (msec % 1000) * 1000000;
	if (timerfd_settime(tfd, 0, &ts, NULL) < 0) {
		printf("timerfd_settime() failed: errno=%d\n", errno);
		close(tfd);
		return -2;
    }
	return tfd;
}
*/
import "C"

import (
	"fmt"
	"syscall"
)

type EventLoop struct {
	pfd      int
	eventors map[int32]*Eventor
}

func NewEventLoop() *EventLoop {
	epollfd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	eventLoop := EventLoop{
		eventors: make(map[int32]*Eventor),
		pfd:      epollfd,
	}
	return &eventLoop
}

func (*EventLoop) Post(task func()) {
}

func (*EventLoop) Stop() {
}

func (loop *EventLoop) AddEvent(e *Eventor) {
	loop.eventors[int32(e.fd)] = e
	e.SetEventLoop(loop)
	if err := syscall.EpollCtl(loop.pfd, syscall.EPOLL_CTL_ADD, e.fd,
		&syscall.EpollEvent{Fd: int32(e.fd), Events: e.event}); err != nil {
		panic(err)
	}
}

func (loop *EventLoop) UpdateEvent(e *Eventor) {
	loop.eventors[int32(e.fd)] = e
	if err := syscall.EpollCtl(loop.pfd, syscall.EPOLL_CTL_MOD, e.fd,
		&syscall.EpollEvent{Fd: int32(e.fd), Events: e.event}); err != nil {
		panic(err)
	}
}

func (loop *EventLoop) RemoveEvent(e *Eventor) {
	if err := syscall.EpollCtl(loop.pfd, syscall.EPOLL_CTL_DEL, e.fd,
		&syscall.EpollEvent{Fd: int32(e.fd), Events: e.event}); err != nil {
		panic(err)
	}
	syscall.Close(e.fd)
	_, ok := loop.eventors[int32(e.fd)]
	if ok {
		delete(loop.eventors, int32(e.fd))
	}
}

func (loop *EventLoop) Poll() {
	events := make([]syscall.EpollEvent, 64)
	n, err := syscall.EpollWait(loop.pfd, events, -1)
	if err != nil && err != syscall.EINTR {
		panic(err)
	}

	for i := 0; i < n; i++ {
		ev := events[i]
		eventor, ok := loop.eventors[ev.Fd]
		if !ok {
			continue
		}

		if ev.Events&syscall.EPOLLIN != 0 {
			fmt.Println("new in events")
		}
		if ev.Events&syscall.EPOLLOUT != 0 {
			fmt.Println("new out events")
		}

		if (eventor.event & ev.Events) == 0 {
			continue
		}

		eventor.cb(ev.Events, eventor)
	}
}

func (loop *EventLoop) Loop() {
	for {
		loop.Poll()
	}
}

func (loop *EventLoop) NotifyClient(clientfd int32, out []byte) {
	eventor, ok := loop.eventors[clientfd]
	if !ok {
		fmt.Println("client not exist")
		return
	}

	eventor.out = append(eventor.out, out...)
	if len(eventor.out) > 0 {
		eventor.EnableWrite()
		eventor.loop.UpdateEvent(eventor)
	}
}

// time event
type TimeId int

func (loop *EventLoop) RunAfter(task func(uint32, *Eventor), time int) {
	timerfd := C.CreateTimerfd(C.int(time))
	timer := Eventor{
		loop: loop,
		fd:   int(timerfd),
		cb:   task,
	}
	timer.EnableRead()
	loop.AddEvent(&timer)
}

func (*EventLoop) RunPeriodic(task func(), time int) TimeId {
	return 1
}

func (*EventLoop) CancelTimer(timeId TimeId) {
}
