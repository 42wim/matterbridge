package tcp

type resultPipes interface {
	popResultPipe(int) (chan error, bool)
	deregisterResultPipe(int)
	registerResultPipe(int, chan error)
}
