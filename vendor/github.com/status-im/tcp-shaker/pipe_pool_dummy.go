package tcp

type pipePoolDummy struct{}

func newPipePoolDummy() *pipePoolDummy {
	return &pipePoolDummy{}
}

func (*pipePoolDummy) getPipe() chan error {
	return make(chan error, 1)
}

func (*pipePoolDummy) putBackPipe(pipe chan error) {}
