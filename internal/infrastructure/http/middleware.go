package http

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

const (
	APIKeyHeader     = "X-API-Key"
	APIKeyQueryParam = "api_key"

	visitorTTL = 3 * time.Minute
)

func APIKey(key string, next http.Handler) http.Handler {
	if key == "" {
		return next
	}

	expected := []byte(key)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provided := []byte(extractAPIKey(r))
		if subtle.ConstantTimeCompare(provided, expected) != 1 {
			log.Warn().Str("path", r.URL.Path).Str("remote", r.RemoteAddr).Msg("invalid api key")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractAPIKey(r *http.Request) string {
	if key := r.Header.Get(APIKeyHeader); key != "" {
		return key
	}
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return r.URL.Query().Get(APIKeyQueryParam)
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
	TrustProxy        bool
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	cfg         RateLimitConfig
	mu          sync.Mutex
	visitors    map[string]*visitor
	lastCleanup time.Time
}

func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	return &RateLimiter{cfg: cfg, visitors: make(map[string]*visitor), lastCleanup: time.Now()}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.lastCleanup) > visitorTTL {
		for key, v := range rl.visitors {
			if now.Sub(v.lastSeen) > visitorTTL {
				delete(rl.visitors, key)
			}
		}
		rl.lastCleanup = now
	}

	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rate.Limit(rl.cfg.RequestsPerSecond), rl.cfg.Burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = now

	return v.limiter.Allow()
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	if rl.cfg.RequestsPerSecond <= 0 {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r, rl.cfg.TrustProxy)
		if !rl.Allow(ip) {
			log.Warn().Str("path", r.URL.Path).Str("ip", ip).Msg("rate limit exceeded")
			w.Header().Set("Retry-After", "1")
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip, _, _ := strings.Cut(forwarded, ",")
			return strings.TrimSpace(ip)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func MaxBodySize(limit int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		next.ServeHTTP(w, r)
	})
}
