package balancer

import (
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Target struct {
	URL              *url.URL
	FailureThreshold int
	failureCount     int
	lastFailure      time.Time
	Timeout          time.Duration
	state            int
	mu               sync.Mutex
}

type Balancer struct {
	Targets []*Target
	current uint64
}

var Balancers = make(map[string]*Balancer)

func (b *Balancer) Next() *url.URL {
	n := len(b.Targets)
	if n == 0 {
		return nil
	}

	failCount := 0
	for failCount < n {
		idx := atomic.AddUint64(&b.current, 1)

		target := b.Targets[(idx-1)%uint64(n)]

		target.mu.Lock()
		s := target.state
		lastfail := target.lastFailure
		target.mu.Unlock()

		if s == 0 {
			return target.URL
		}

		if s == 1 {
			if time.Since(lastfail) > target.Timeout {

				target.mu.Lock()
				target.state = 2
				target.mu.Unlock()

				return target.URL
			}
			failCount++
		}
	}
	return nil
}

var targetUrlDirectory = make(map[string]int)

func RecordStatus(prefix, targetUrl string, status int) {
	if b, ok := Balancers[prefix]; ok {
		targetIdx := -1
		if target, ok := targetUrlDirectory[targetUrl]; ok {
			targetIdx = target
		} else {
			for idx := range b.Targets {
				if b.Targets[idx].URL.String() == targetUrl {
					targetIdx = idx
					break
				}
			}
		}

		if targetIdx > -1 {
			targetServer := b.Targets[targetIdx]

			targetServer.mu.Lock()

			if status >= 500 {

				targetServer.failureCount++
				targetServer.lastFailure = time.Now()

				if targetServer.failureCount > targetServer.FailureThreshold {
					targetServer.state = 1
				}

			} else {
				targetServer.state = 0
				targetServer.failureCount = 0
			}
		}
	}
}
