package discovery

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/discv5"
)

// NewMultiplexer creates Multiplexer instance.
func NewMultiplexer(discoveries []Discovery) Multiplexer {
	return Multiplexer{discoveries}
}

// Multiplexer allows to use multiple discoveries behind single Discovery interface.
type Multiplexer struct {
	discoveries []Discovery
}

// Running should return true if at least one discovery is running
func (m Multiplexer) Running() (rst bool) {
	for i := range m.discoveries {
		rst = rst || m.discoveries[i].Running()
	}
	return rst
}

// Start every discovery and stop every started in case if at least one fails.
func (m Multiplexer) Start() (err error) {
	started := []int{}
	for i := range m.discoveries {
		if err = m.discoveries[i].Start(); err != nil {
			break
		}
		started = append(started, i)
	}
	if err != nil {
		for _, i := range started {
			_ = m.discoveries[i].Stop()
		}
	}
	return err
}

// Stop every discovery.
func (m Multiplexer) Stop() (err error) {
	messages := []string{}
	for i := range m.discoveries {
		if err = m.discoveries[i].Stop(); err != nil {
			messages = append(messages, err.Error())
		}
	}
	if len(messages) != 0 {
		return fmt.Errorf("failed to stop discoveries: %s", strings.Join(messages, "; "))
	}
	return nil
}

// Register passed topic and stop channel to every discovery and waits till it will return.
func (m Multiplexer) Register(topic string, stop chan struct{}) error {
	errors := make(chan error, len(m.discoveries))
	for i := range m.discoveries {
		i := i
		go func() {
			errors <- m.discoveries[i].Register(topic, stop)
		}()
	}
	total := 0
	messages := []string{}
	for err := range errors {
		total++
		if err != nil {
			messages = append(messages, err.Error())
		}
		if total == len(m.discoveries) {
			break
		}
	}
	if len(messages) != 0 {
		return fmt.Errorf("failed to register %s: %s", topic, strings.Join(messages, "; "))
	}
	return nil
}

// Discover shares topic and channles for receiving results. And multiplexer periods that are sent to period channel.
func (m Multiplexer) Discover(topic string, period <-chan time.Duration, found chan<- *discv5.Node, lookup chan<- bool) error {
	var (
		periods  = make([]chan time.Duration, len(m.discoveries))
		messages = []string{}
		wg       sync.WaitGroup
		mu       sync.Mutex
	)
	wg.Add(len(m.discoveries) + 1)
	for i := range m.discoveries {
		i := i
		periods[i] = make(chan time.Duration, 2)
		go func() {
			err := m.discoveries[i].Discover(topic, periods[i], found, lookup)
			if err != nil {
				mu.Lock()
				messages = append(messages, err.Error())
				mu.Unlock()
			}
			wg.Done()
		}()
	}
	go func() {
		for {
			newPeriod, ok := <-period
			for i := range periods {
				if !ok {
					close(periods[i])
				} else {
					periods[i] <- newPeriod
				}
			}
			if !ok {
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()
	if len(messages) != 0 {
		return fmt.Errorf("failed to discover topic %s: %s", topic, strings.Join(messages, "; "))
	}
	return nil
}
