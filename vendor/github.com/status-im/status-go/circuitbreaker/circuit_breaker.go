package circuitbreaker

import (
	"fmt"
	"strings"

	"github.com/afex/hystrix-go/hystrix"
)

type FallbackFunc func() ([]any, error)

type CommandResult struct {
	res []any
	err error
}

func (cr CommandResult) Result() []any {
	return cr.res
}

func (cr CommandResult) Error() error {
	return cr.err
}

type Command struct {
	functors []*Functor
}

func NewCommand(functors []*Functor) *Command {
	return &Command{
		functors: functors,
	}
}

func (cmd *Command) Add(ftor *Functor) {
	cmd.functors = append(cmd.functors, ftor)
}

func (cmd *Command) IsEmpty() bool {
	return len(cmd.functors) == 0
}

type Config struct {
	CommandName            string
	Timeout                int
	MaxConcurrentRequests  int
	RequestVolumeThreshold int
	SleepWindow            int
	ErrorPercentThreshold  int
}

type CircuitBreaker struct {
	config Config
}

func NewCircuitBreaker(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
	}
}

type Functor struct {
	Exec FallbackFunc
}

func NewFunctor(exec FallbackFunc) *Functor {
	return &Functor{
		Exec: exec,
	}
}

// This a blocking function
func (eh *CircuitBreaker) Execute(cmd Command) CommandResult {
	resultChan := make(chan CommandResult, 1)
	var result CommandResult

	for i := 0; i < len(cmd.functors); i += 2 {
		f1 := cmd.functors[i]
		var f2 *Functor
		if i+1 < len(cmd.functors) {
			f2 = cmd.functors[i+1]
		}

		circuitName := fmt.Sprintf("%s_%d", eh.config.CommandName, i)
		if hystrix.GetCircuitSettings()[circuitName] == nil {
			hystrix.ConfigureCommand(circuitName, hystrix.CommandConfig{
				Timeout:                eh.config.Timeout,
				MaxConcurrentRequests:  eh.config.MaxConcurrentRequests,
				RequestVolumeThreshold: eh.config.RequestVolumeThreshold,
				SleepWindow:            eh.config.SleepWindow,
				ErrorPercentThreshold:  eh.config.ErrorPercentThreshold,
			})
		}

		// If circuit is the same for all functions, in case of len(cmd.functors) > 2,
		// main and fallback providers are different next run if first two fail,
		// which causes health issues for both main and fallback and ErrorPercentThreshold
		// is reached faster than it should be.
		errChan := hystrix.Go(circuitName, func() error {
			res, err := f1.Exec()
			// Write to resultChan only if success
			if err == nil {
				resultChan <- CommandResult{res: res, err: err}
			}
			return err
		}, func(err error) error {
			// In case of concurrency, we should not execute the fallback
			if f2 == nil || err == hystrix.ErrMaxConcurrency {
				return err
			}
			res, err := f2.Exec()
			if err == nil {
				resultChan <- CommandResult{res: res, err: err}
			}
			return err
		})

		select {
		case result = <-resultChan:
			if result.err == nil {
				return result
			}
		case err := <-errChan:
			result = CommandResult{err: err}

			// In case of max concurrency, we should delay the execution and stop iterating over fallbacks
			// No error unwrapping here, so use strings.Contains
			if strings.Contains(err.Error(), hystrix.ErrMaxConcurrency.Error()) {
				return result
			}
		}
	}

	return result
}
