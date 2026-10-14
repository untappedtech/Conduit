package http

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/untappedtech/conduit/internal/auth"
	"github.com/untappedtech/conduit/internal/domain"
	"github.com/untappedtech/conduit/internal/errors"
	"golang.org/x/time/rate"
)

type clientLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64 // unix nano timestamp
}

// RateLimiter manages per-client Token Bucket rate limiters keyed by IP or token.
type RateLimiter struct {
	config          domain.RateLimitConfig
	responseEncoder *ResponseEncoder
	tokenExtractor  auth.TokenExtractor
	clients         sync.Map // string -> *clientLimiterEntry
	stopSweep       chan struct{}
	sweepInterval   time.Duration
	staleTimeout    time.Duration
	sweepOnce       sync.Once
}

// NewRateLimiter creates a new RateLimiter instance.
func NewRateLimiter(config domain.RateLimitConfig, responseEncoder *ResponseEncoder, tokenExtractor ...auth.TokenExtractor) *RateLimiter {
	var extractor auth.TokenExtractor
	if len(tokenExtractor) > 0 && tokenExtractor[0] != nil {
		extractor = tokenExtractor[0]
	} else {
		extractor = auth.NewDefaultTokenExtractor()
	}

	rl := &RateLimiter{
		config:          config,
		responseEncoder: responseEncoder,
		tokenExtractor:  extractor,
		stopSweep:       make(chan struct{}),
		sweepInterval:   time.Minute,
		staleTimeout:    3 * time.Minute,
	}

	if config.Enabled {
		go rl.sweepLoop()
	}

	return rl
}

// SetSweepTimes allows configuring custom intervals for testing or specialized environments.
func (rl *RateLimiter) SetSweepTimes(sweepInterval, staleTimeout time.Duration) {
	rl.sweepInterval = sweepInterval
	rl.staleTimeout = staleTimeout
}

// Close gracefully stops the background sweeper goroutine.
func (rl *RateLimiter) Close() {
	rl.sweepOnce.Do(func() {
		close(rl.stopSweep)
	})
}

func (rl *RateLimiter) sweepLoop() {
	ticker := time.NewTicker(rl.sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopSweep:
			return
		case <-ticker.C:
			rl.Sweep()
		}
	}
}

// Sweep removes inactive limiters that have not been seen within staleTimeout.
func (rl *RateLimiter) Sweep() {
	now := time.Now()
	rl.clients.Range(func(key, val any) bool {
		entry := val.(*clientLimiterEntry)
		lastSeen := time.Unix(0, entry.lastSeen.Load())
		if now.Sub(lastSeen) > rl.staleTimeout {
			rl.clients.Delete(key)
		}
		return true
	})
}

// ActiveLimitersCount returns the number of active limiters currently in the cache.
func (rl *RateLimiter) ActiveLimitersCount() int {
	count := 0
	rl.clients.Range(func(key, val any) bool {
		count++
		return true
	})
	return count
}

// ExtractClientKey determines the rate-limiting key for the request (token if present, else client IP).
func (rl *RateLimiter) ExtractClientKey(r *http.Request) string {
	if rl.tokenExtractor != nil {
		if token := rl.tokenExtractor.ExtractToken(r); token != "" {
			return "token:" + token
		}
	}
	return "ip:" + ExtractClientIP(r)
}

// ExtractClientIP retrieves the real client IP address from proxy headers or RemoteAddr.
func ExtractClientIP(r *http.Request) string {
	// 1. Check X-Forwarded-For (first entry is the original client IP)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 2. Check X-Real-IP
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		ip := strings.TrimSpace(xrip)
		if ip != "" {
			return ip
		}
	}

	// 3. Fallback to RemoteAddr (strip port if present)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

// GetLimiter returns or creates a *rate.Limiter for the given key.
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	now := time.Now().UnixNano()
	if val, ok := rl.clients.Load(key); ok {
		entry := val.(*clientLimiterEntry)
		entry.lastSeen.Store(now)
		return entry.limiter
	}

	limit := rate.Limit(rl.config.RequestsPerSecond)
	burst := rl.config.Burst
	if burst <= 0 {
		burst = int(rl.config.RequestsPerSecond)
		if burst <= 0 {
			burst = 1
		}
	}
	limiter := rate.NewLimiter(limit, burst)

	newEntry := &clientLimiterEntry{
		limiter: limiter,
	}
	newEntry.lastSeen.Store(now)

	actual, _ := rl.clients.LoadOrStore(key, newEntry)
	entry := actual.(*clientLimiterEntry)
	entry.lastSeen.Store(now)
	return entry.limiter
}

// Middleware returns an HTTP middleware enforcing rate limits.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			key := rl.ExtractClientKey(r)
			limiter := rl.GetLimiter(key)

			if !limiter.Allow() {
				if rl.responseEncoder != nil {
					rl.responseEncoder.EncodeError(w, r, errors.ErrTooManyRequests)
				} else {
					http.Error(w, "Rate limit exceeded.", http.StatusTooManyRequests)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware provides a standalone middleware constructor for rate limiting.
func RateLimitMiddleware(config domain.RateLimitConfig, responseEncoder *ResponseEncoder, tokenExtractor ...auth.TokenExtractor) func(http.Handler) http.Handler {
	rl := NewRateLimiter(config, responseEncoder, tokenExtractor...)
	return rl.Middleware()
}
