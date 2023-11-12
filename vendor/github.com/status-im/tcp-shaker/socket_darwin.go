// +build darwin

package tcp

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

const maxKEvents = 32

// createSocket creates a socket with necessary options set.
func createSocketZeroLinger(zeroLinger bool) (fd int, err error) {
	// Create socket
	fd, err = _createNonBlockingSocket()
	if err == nil {
		if zeroLinger {
			err = _setZeroLinger(fd)
		}
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

	//unix.CloseOnExec(fd)
	return fd, err
}

// setSockOpts sets SOCK_NONBLOCK and TCP_NODELAY for given fd
func _setSockOpts(fd int) error {
	err := unix.SetNonblock(fd, true)
	if err != nil {
		return err
	}
	return unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, 0)
}

var zeroLinger = unix.Linger{Onoff: 1, Linger: 0}

// setLinger sets SO_Linger with 0 timeout to given fd
func _setZeroLinger(fd int) error {
	return unix.SetsockoptLinger(fd, unix.SOL_SOCKET, unix.SO_LINGER, &zeroLinger)
}

func createPoller() (kq int, err error) {
	kq, err = unix.Kqueue()
	if err != nil {
		err = os.NewSyscallError("kqueue", err)
	}
	return kq, err
}

// registerEvents registers given fd with read events.
func registerEvents(kq, fd int) error {
	eventFilter := unix.Kevent_t{}
	unix.SetKevent(&eventFilter, fd, unix.EVFILT_WRITE, unix.EV_ADD|unix.EV_ONESHOT)
	// doesn't block, just sets the events we are interested in
	_, err := unix.Kevent(kq, []unix.Kevent_t{eventFilter}, []unix.Kevent_t{}, nil)
	if err != nil {
		return os.NewSyscallError("kevent", err)
	}
	return nil
}

func pollEvents(kq int, timeout time.Duration) ([]event, error) {
	tsTimeout := unix.NsecToTimespec(timeout.Nanoseconds())
	rEvents := make([]unix.Kevent_t, maxKEvents)
	// this blocks, waiting for socket events
	nEvents, err := unix.Kevent(kq, []unix.Kevent_t{}, rEvents, &tsTimeout)
	if err != nil {
		if err == unix.EINTR {
			return nil, nil
		}
		return nil, os.NewSyscallError("kevent", err)
	}

	var events = make([]event, 0, nEvents)

	for i := 0; i < nEvents; i++ {
		var fd = int(rEvents[i].Ident)
		var evt = event{Fd: fd, Err: nil}

		errCode, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err != nil {
			evt.Err = os.NewSyscallError("getsockopt", err)
		}
		if errCode != 0 {
			evt.Err = newErrConnect(errCode)
		}
		events = append(events, evt)
	}
	return events, nil
}
