package netpoll

type FDOperator struct {
	// FD is file descriptor, poll will bind when register.
	FD int

	// Desc provides three callbacks for fd's reading, writing or hanging events.
	OnRead  func(*FDOperator)
	OnWrite func(*FDOperator)
	OnHup   func(*FDOperator)

	// Epoll is the registered location of the file descriptor.
	Epoll *Epoll

	Data any

	// private, used by operatorCache
	next  *FDOperator
	index int32 // index in operatorCache
}

func (operator *FDOperator) reset() {
	operator.FD = 0
	operator.OnRead, operator.OnWrite, operator.OnHup = nil, nil, nil
	operator.Epoll = nil
	operator.Data = nil
}
