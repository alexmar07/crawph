package ratelimit

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

type Limiter struct {
	limiter *rate.Limiter
}

func (l *Limiter) Wait() {
	_ = l.limiter.Wait(context.Background())
}

type Registry struct {
	defaultRPS float64
	limiters   map[string]*Limiter
	mu         sync.Mutex
}

func NewRegistry(defaultRPS float64) *Registry {
	return &Registry{
		defaultRPS: defaultRPS,
		limiters:   make(map[string]*Limiter),
	}
}

func (r *Registry) GetLimiter(domain string) *Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if l, ok := r.limiters[domain]; ok {
		return l
	}
	l := &Limiter{
		limiter: rate.NewLimiter(rate.Limit(r.defaultRPS), 1),
	}
	r.limiters[domain] = l
	return l
}

// SetDomainRate updates the rate on the existing limiter for a domain,
// or creates a new one. Uses SetLimit so existing references stay valid.
func (r *Registry) SetDomainRate(domain string, rps float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if l, ok := r.limiters[domain]; ok {
		l.limiter.SetLimit(rate.Limit(rps))
		return
	}
	r.limiters[domain] = &Limiter{
		limiter: rate.NewLimiter(rate.Limit(rps), 1),
	}
}
