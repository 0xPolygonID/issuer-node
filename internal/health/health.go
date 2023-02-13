package health

import (
	"context"
	"sync"
	"time"
)

const (
	DefaultPingPeriod = 5 * time.Second // DefaultPingPeriod is a recommendation to ping any service
)

// Status struct
type Status struct {
	sync.RWMutex
	monitors     Monitors
	lastStatuses map[string]bool
}

// Pinger is a function that return error if cannot ping. False otherwise
type Pinger func(ctx context.Context) error

// Monitors represents a map of Pingers identified by it's human name
type Monitors map[string]Pinger

// New returns a Health instance
func New(m Monitors) *Status {
	return &Status{
		monitors:     m,
		lastStatuses: make(map[string]bool, len(m)),
	}
}

// Run starts a monitor that will check each service every t duration.
func (s *Status) Run(ctx context.Context, t time.Duration) {
	go func() {
		timer := time.NewTicker(t)
		s.checkStatus(ctx)
		for {
			select {
			case <-timer.C:
				s.checkStatus(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Status returns a map of service_name, bool telling whether some service is down
// true: Ok. false: Cannot ping
func (s *Status) Status() map[string]bool {
	s.RLock()
	defer s.RUnlock()
	return s.lastStatuses
}

func (s *Status) checkStatus(ctx context.Context) {
	s.Lock()
	defer s.Unlock()
	for service, ping := range s.monitors {
		s.lastStatuses[service] = ping(ctx) == nil
	}
}
