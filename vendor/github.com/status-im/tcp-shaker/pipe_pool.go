package tcp

type pipePool interface {
	getPipe() chan error
	putBackPipe(chan error)
}
