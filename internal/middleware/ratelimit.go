package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

// visitor tracks request timestamps for a single IP.
type visitor struct {
	mu       sync.Mutex
	tokens   float64
	lastSeen time.Time
	rate     float64 // tokens per second
	burst    float64 // max tokens
}

func (v *visitor) allow() bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.lastSeen = now

	// Refill tokens
	v.tokens += elapsed * v.rate
	if v.tokens > v.burst {
		v.tokens = v.burst
	}

	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// IPRateLimiter provides per-IP rate limiting.
type IPRateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	rate     float64 // tokens per second
	burst    float64
	cleanup  time.Duration
}

// NewIPRateLimiter creates a rate limiter.
// rate is requests per second, burst is the max burst size.
func NewIPRateLimiter(rate float64, burst int) *IPRateLimiter {
	rl := &IPRateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    float64(burst),
		cleanup:  3 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *IPRateLimiter) getVisitor(ip string) *visitor {
	rl.mu.RLock()
	v, ok := rl.visitors[ip]
	rl.mu.RUnlock()
	if ok {
		return v
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after write lock
	if v, ok := rl.visitors[ip]; ok {
		return v
	}

	v = &visitor{
		tokens:   rl.burst,
		lastSeen: time.Now(),
		rate:     rl.rate,
		burst:    rl.burst,
	}
	rl.visitors[ip] = v
	return v
}

func (rl *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if time.Since(v.lastSeen) > rl.cleanup {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns middleware that limits requests per IP.
func RateLimit(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			if !limiter.getVisitor(ip).allow() {
				slog.Warn("rate_limited", "ip", ip, "path", r.URL.Path)
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// extractIP returns the client IP from X-Forwarded-For (Cloud Run) or RemoteAddr.
func extractIP(r *http.Request) string {
	// Cloud Run sets X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// First IP in the chain is the client
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
