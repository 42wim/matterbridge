package timesource

import (
	"bytes"
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/beevik/ntp"
	"go.uber.org/zap"
)

const (
	// DefaultMaxAllowedFailures defines how many failures will be tolerated.
	DefaultMaxAllowedFailures = 1

	// FastNTPSyncPeriod period between ntp synchronizations before the first
	// successful connection.
	FastNTPSyncPeriod = 2 * time.Minute

	// SlowNTPSyncPeriod period between ntp synchronizations after the first
	// successful connection.
	SlowNTPSyncPeriod = 1 * time.Hour

	// DefaultRPCTimeout defines write deadline for single ntp server request.
	DefaultRPCTimeout = 2 * time.Second
)

// DefaultServers will be resolved to the closest available,
// and with high probability resolved to the different IPs
var DefaultServers = []string{
	"0.pool.ntp.org",
	"1.pool.ntp.org",
	"2.pool.ntp.org",
	"3.pool.ntp.org",
}
var errUpdateOffset = errors.New("failed to compute offset")

type ntpQuery func(string, ntp.QueryOptions) (*ntp.Response, error)

type queryResponse struct {
	Offset time.Duration
	Error  error
}

type multiRPCError []error

func (e multiRPCError) Error() string {
	var b bytes.Buffer
	b.WriteString("RPC failed: ")
	more := false
	for _, err := range e {
		if more {
			b.WriteString("; ")
		}
		b.WriteString(err.Error())
		more = true
	}
	b.WriteString(".")
	return b.String()
}

func computeOffset(timeQuery ntpQuery, servers []string, allowedFailures int) (time.Duration, error) {
	if len(servers) == 0 {
		return 0, nil
	}
	responses := make(chan queryResponse, len(servers))
	for _, server := range servers {
		go func(server string) {
			response, err := timeQuery(server, ntp.QueryOptions{
				Timeout: DefaultRPCTimeout,
			})
			if err == nil {
				err = response.Validate()
			}
			if err != nil {
				responses <- queryResponse{Error: err}
				return
			}
			responses <- queryResponse{Offset: response.ClockOffset}
		}(server)
	}
	var (
		rpcErrors multiRPCError
		offsets   []time.Duration
		collected int
	)
	for response := range responses {
		if response.Error != nil {
			rpcErrors = append(rpcErrors, response.Error)
		} else {
			offsets = append(offsets, response.Offset)
		}
		collected++
		if collected == len(servers) {
			break
		}
	}
	if lth := len(rpcErrors); lth > allowedFailures {
		return 0, rpcErrors
	} else if lth == len(servers) {
		return 0, rpcErrors
	}
	sort.SliceStable(offsets, func(i, j int) bool {
		return offsets[i] > offsets[j]
	})
	mid := len(offsets) / 2
	if len(offsets)%2 == 0 {
		return (offsets[mid-1] + offsets[mid]) / 2, nil
	}
	return offsets[mid], nil
}

// NewNTPTimesource creates a timesource that uses NTP
func NewNTPTimesource(ntpServers []string, log *zap.Logger) *NTPTimeSource {
	return &NTPTimeSource{
		servers:           ntpServers,
		allowedFailures:   DefaultMaxAllowedFailures,
		fastNTPSyncPeriod: FastNTPSyncPeriod,
		slowNTPSyncPeriod: SlowNTPSyncPeriod,
		timeQuery:         ntp.QueryWithOptions,
		log:               log.Named("timesource"),
	}
}

// NTPTimeSource provides source of time that tries to be resistant to time skews.
// It does so by periodically querying time offset from ntp servers.
type NTPTimeSource struct {
	servers           []string
	allowedFailures   int
	fastNTPSyncPeriod time.Duration
	slowNTPSyncPeriod time.Duration
	timeQuery         ntpQuery // for ease of testing
	log               *zap.Logger

	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu           sync.RWMutex
	latestOffset time.Duration
}

// Now returns time adjusted by latest known offset
func (s *NTPTimeSource) Now() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Now().Add(s.latestOffset)
}

func (s *NTPTimeSource) updateOffset() error {
	offset, err := computeOffset(s.timeQuery, s.servers, s.allowedFailures)
	if err != nil {
		s.log.Error("failed to compute offset", zap.Error(err))
		return errUpdateOffset
	}
	s.log.Info("Difference with ntp servers", zap.Duration("offset", offset))
	s.mu.Lock()
	s.latestOffset = offset
	s.mu.Unlock()
	return nil
}

// runPeriodically runs periodically the given function based on NTPTimeSource
// synchronization limits (fastNTPSyncPeriod / slowNTPSyncPeriod)
func (s *NTPTimeSource) runPeriodically(ctx context.Context, fn func() error) error {
	var period time.Duration

	s.log.Info("starting service")

	// we try to do it synchronously so that user can have reliable messages right away
	s.wg.Add(1)
	go func() {
		for {
			select {
			case <-time.After(period):
				if err := fn(); err == nil {
					period = s.slowNTPSyncPeriod
				} else if period != s.slowNTPSyncPeriod {
					period = s.fastNTPSyncPeriod
				}

			case <-ctx.Done():
				s.log.Info("stopping service")
				s.wg.Done()
				return
			}
		}
	}()

	return nil
}

// Start runs a goroutine that updates local offset every updatePeriod.
func (s *NTPTimeSource) Start(ctx context.Context) error {
	s.wg.Wait() // Waiting for other go routines to stop
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	return s.runPeriodically(ctx, s.updateOffset)
}

// Stop goroutine that updates time source.
func (s *NTPTimeSource) Stop() {
	if s.cancel == nil {
		return
	}
	s.cancel()
	s.wg.Wait()
}
