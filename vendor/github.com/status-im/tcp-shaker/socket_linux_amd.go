//go:build android && amd64
// +build android,amd64

package tcp

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

const maxEpollEvents = 32

// createSocket creates a socket with necessary options set.
func createSocketZeroLinger(zeroLinger bool) (fd int, err error) {
	// Create socket
	fd, err = _createNonBlockingSocket()
	if err == nil && zeroLinger {
		err = _setZeroLinger(fd)
	}
	return
}

// createNonBlockingSocket creates a non-blocking socket with necessary options all set.
func _createNonBlockingSocket() (int, error) {
	// Create socket
	fd, err := _createSocket()
	if err != nil {
		return 0, err
	}
	// Set necessary options
	err = _setSockOpts(fd)
	if err != nil {
		unix.Close(fd)
	}
	return fd, err
}

// createSocket creates a socket with CloseOnExec set
func _createSocket() (int, error) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		return 0, os.NewSyscallError("socket", err)
	}

	unix.CloseOnExec(fd)
	return fd, err
}

// setSockOpts sets SOCK_NONBLOCK and TCP_QUICKACK for given fd
func _setSockOpts(fd int) error {
	err := unix.SetNonblock(fd, true)
	if err != nil {
		return err
	}
	return unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_QUICKACK, 0)
}

var zeroLinger = unix.Linger{Onoff: 1, Linger: 0}

// setLinger sets SO_Linger with 0 timeout to given fd
func _setZeroLinger(fd int) error {
	return unix.SetsockoptLinger(fd, unix.SOL_SOCKET, unix.SO_LINGER, &zeroLinger)
}

func createPoller() (fd int, err error) {
	fd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		err = os.NewSyscallError("epoll_create1", err)
	}
	return fd, err
}

// registerEvents registers given fd with read and write events.
func registerEvents(pollerFd int, fd int) error {
	var event unix.EpollEvent
	event.Events = unix.EPOLLOUT | unix.EPOLLIN | unix.EPOLLET
	event.Fd = int32(fd)
	if err := unix.EpollCtl(pollerFd, unix.EPOLL_CTL_ADD, fd, &event); err != nil {
		return os.NewSyscallError(fmt.Sprintf("epoll_ctl(%d, ADD, %d, ...)", pollerFd, fd), err)
	}
	return nil
}

func pollEvents(pollerFd int, timeout time.Duration) ([]event, error) {
	eventCh := make(chan event)
	errorCh := make(chan error)
	doneCh := make(chan bool)

	go func(fd int) {
		for {
			select {
			case <-doneCh:
				return
			default:
				n, _, err := unix.Recvfrom(fd, nil, unix.MSG_DONTWAIT|unix.MSG_PEEK)
				if err != nil && err != unix.EAGAIN && err != unix.EWOULDBLOCK {
					errorCh <- os.NewSyscallError("recvfrom", err)
					return
				}
				if n > 0 {
					eventCh <- event{Fd: fd, Err: nil}
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}(pollerFd)

	var events []event

	timer := time.NewTimer(timeout)
	defer timer.Stop()

Loop:
	for {
		select {
		case evt := <-eventCh:
			events = append(events, evt)
		case err := <-errorCh:
			return nil, err
		case <-timer.C:
			break Loop
		}
	}

	// Signal the goroutine to stop.
	close(doneCh)

	return events, nil
}
