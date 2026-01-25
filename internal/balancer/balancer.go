package balancer

import (
	"net/url"
	"sync/atomic"
)

type Balancer struct {
	Targets []*url.URL
	current uint64
}

func (b *Balancer) Next() *url.URL {
	if len(b.Targets) == 0 {
		return nil
	}

	idx := atomic.AddUint64(&b.current, 1)
	return b.Targets[(idx-1)%uint64(len(b.Targets))]
}
