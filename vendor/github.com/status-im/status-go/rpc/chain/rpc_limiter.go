package chain

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultMaxRequestsPerSecond = 100
	minRequestsPerSecond        = 20
	requestsPerSecondStep       = 10

	tickerInterval = 1 * time.Second
)

var (
	ErrRequestsOverLimit = fmt.Errorf("number of requests over limit")
)

type callerOnWait struct {
	requests int
	ch       chan bool
}

type RPCLimiter struct {
	uuid uuid.UUID

	maxRequestsPerSecond      int
	maxRequestsPerSecondMutex sync.RWMutex

	requestsMadeWithinSecond      int
	requestsMadeWithinSecondMutex sync.RWMutex

	callersOnWaitForRequests      []callerOnWait
	callersOnWaitForRequestsMutex sync.RWMutex

	quit chan bool
}

func NewRPCLimiter() *RPCLimiter {

	limiter := RPCLimiter{
		uuid:                 uuid.New(),
		maxRequestsPerSecond: defaultMaxRequestsPerSecond,
		quit:                 make(chan bool),
	}

	limiter.start()

	return &limiter
}

func (rl *RPCLimiter) ReduceLimit() {
	rl.maxRequestsPerSecondMutex.Lock()
	defer rl.maxRequestsPerSecondMutex.Unlock()
	if rl.maxRequestsPerSecond <= minRequestsPerSecond {
		return
	}
	rl.maxRequestsPerSecond = rl.maxRequestsPerSecond - requestsPerSecondStep
}

func (rl *RPCLimiter) start() {
	ticker := time.NewTicker(tickerInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				{
					rl.requestsMadeWithinSecondMutex.Lock()
					oldrequestsMadeWithinSecond := rl.requestsMadeWithinSecond
					if rl.requestsMadeWithinSecond != 0 {
						rl.requestsMadeWithinSecond = 0
					}
					rl.requestsMadeWithinSecondMutex.Unlock()
					if oldrequestsMadeWithinSecond == 0 {
						continue
					}
				}

				rl.callersOnWaitForRequestsMutex.Lock()
				numOfRequestsToMakeAvailable := rl.maxRequestsPerSecond
				for {
					if numOfRequestsToMakeAvailable == 0 || len(rl.callersOnWaitForRequests) == 0 {
						break
					}

					var index = -1
					for i := 0; i < len(rl.callersOnWaitForRequests); i++ {
						if rl.callersOnWaitForRequests[i].requests <= numOfRequestsToMakeAvailable {
							index = i
							break
						}
					}

					if index == -1 {
						break
					}

					callerOnWait := rl.callersOnWaitForRequests[index]
					numOfRequestsToMakeAvailable -= callerOnWait.requests
					rl.callersOnWaitForRequests = append(rl.callersOnWaitForRequests[:index], rl.callersOnWaitForRequests[index+1:]...)

					callerOnWait.ch <- true
				}
				rl.callersOnWaitForRequestsMutex.Unlock()

			case <-rl.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (rl *RPCLimiter) Stop() {
	rl.quit <- true
	close(rl.quit)
	for _, callerOnWait := range rl.callersOnWaitForRequests {
		close(callerOnWait.ch)
	}
	rl.callersOnWaitForRequests = nil
}

func (rl *RPCLimiter) WaitForRequestsAvailability(requests int) error {
	if requests > rl.maxRequestsPerSecond {
		return ErrRequestsOverLimit
	}

	{
		rl.requestsMadeWithinSecondMutex.Lock()
		if rl.requestsMadeWithinSecond+requests <= rl.maxRequestsPerSecond {
			rl.requestsMadeWithinSecond += requests
			rl.requestsMadeWithinSecondMutex.Unlock()
			return nil
		}
		rl.requestsMadeWithinSecondMutex.Unlock()
	}

	callerOnWait := callerOnWait{
		requests: requests,
		ch:       make(chan bool),
	}

	{
		rl.callersOnWaitForRequestsMutex.Lock()
		rl.callersOnWaitForRequests = append(rl.callersOnWaitForRequests, callerOnWait)
		rl.callersOnWaitForRequestsMutex.Unlock()
	}

	<-callerOnWait.ch

	close(callerOnWait.ch)

	rl.requestsMadeWithinSecondMutex.Lock()
	rl.requestsMadeWithinSecond += requests
	rl.requestsMadeWithinSecondMutex.Unlock()

	return nil
}
