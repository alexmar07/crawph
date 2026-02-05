package ratelimit

import (
	"testing"
	"time"
)

func TestRegistryReturnsSameLimiterForSameDomain(t *testing.T) {
	r := NewRegistry(1.0)
	l1 := r.GetLimiter("example.com")
	l2 := r.GetLimiter("example.com")
	if l1 != l2 {
		t.Error("expected same limiter for same domain")
	}
}

func TestRegistryReturnsDifferentLimiterForDifferentDomain(t *testing.T) {
	r := NewRegistry(1.0)
	l1 := r.GetLimiter("example.com")
	l2 := r.GetLimiter("other.com")
	if l1 == l2 {
		t.Error("expected different limiter for different domain")
	}
}

func TestSetDomainRateUpdatesExistingLimiter(t *testing.T) {
	r := NewRegistry(10.0)
	l1 := r.GetLimiter("slow.com")
	r.SetDomainRate("slow.com", 0.5)
	l2 := r.GetLimiter("slow.com")
	if l1 != l2 {
		t.Error("expected same limiter object after SetDomainRate")
	}
	start := time.Now()
	l2.Wait()
	l2.Wait()
	elapsed := time.Since(start)
	if elapsed < 1*time.Second {
		t.Errorf("expected ~2s delay at 0.5 rps, got %v", elapsed)
	}
}

func TestDefaultRate(t *testing.T) {
	r := NewRegistry(10.0)
	l := r.GetLimiter("fast.com")
	start := time.Now()
	for i := 0; i < 5; i++ {
		l.Wait()
	}
	elapsed := time.Since(start)
	if elapsed > 1*time.Second {
		t.Errorf("expected <1s at 10 rps for 5 requests, got %v", elapsed)
	}
}
