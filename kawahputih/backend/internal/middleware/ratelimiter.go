package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
	rps      rate.Limit
	burst    int
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPLimiterStore(rps rate.Limit, burst int) *ipLimiterStore {
	s := &ipLimiterStore{limiters: make(map[string]*limiterEntry), rps: rps, burst: burst}
	go s.evictLoop()
	return s
}

func (s *ipLimiterStore) get(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.limiters[ip]
	if !ok {
		entry = &limiterEntry{limiter: rate.NewLimiter(s.rps, s.burst)}
		s.limiters[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (s *ipLimiterStore) evictLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for ip, entry := range s.limiters {
			if time.Since(entry.lastSeen) > 15*time.Minute {
				delete(s.limiters, ip)
			}
		}
		s.mu.Unlock()
	}
}

func RateLimit(rps float64, burst int) gin.HandlerFunc {
	store := newIPLimiterStore(rate.Limit(rps), burst)
	return func(c *gin.Context) {
		if !store.get(c.ClientIP()).Allow() {
			response.TooManyRequests(c, "")
			c.Abort()
			return
		}
		c.Next()
	}
}

// TieredRateLimit menerapkan tingkatan throughput global: Tamu 60 req/menit (berdasarkan IP), Pengguna 120 req/menit, Admin 300 req/menit (berdasarkan ID pengguna). Ini terpisah dari RateLimit per-endpoint yang lebih ketat di /auth/register, /auth/login, dll; keduanya berlaku bersamaan, pelapisan yang disengaja.
func TieredRateLimit(jwtManager *jwtutil.Manager) gin.HandlerFunc {
	guestStore := newIPLimiterStore(rate.Limit(60.0/60.0), 10)
	userStore := newIPLimiterStore(rate.Limit(120.0/60.0), 20)
	adminStore := newIPLimiterStore(rate.Limit(300.0/60.0), 40)

	return func(c *gin.Context) {
		var limiter *rate.Limiter

		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			if claims, err := jwtManager.ParseAccessToken(strings.TrimPrefix(header, "Bearer ")); err == nil {
				if isAdminRole(claims.Roles) {
					limiter = adminStore.get(claims.UserID)
				} else {
					limiter = userStore.get(claims.UserID)
				}
			}
			// Token tidak valid/kedaluwarsa TIDAK ditolak middleware ini; itu tugas RequireAuth. Middleware ini hanya mencoba mengklasifikasi tier; bila gagal, jatuh ke tier tamu.
		}
		if limiter == nil {
			limiter = guestStore.get(c.ClientIP())
		}

		if !limiter.Allow() {
			response.TooManyRequests(c, "rate limit exceeded, please slow down")
			c.Abort()
			return
		}
		c.Next()
	}
}

// isAdminRole mengenali konstanta role kanonik (lihat internal/domain/role); ADMIN adalah yang sebenarnya dibawa seed dan JWT. Nama huruf kecil lama juga diterima agar token lama tetap terklasifikasi benar.
func isAdminRole(roles []string) bool {
	for _, r := range roles {
		switch r {
		case "ADMIN", "admin", "super_admin":
			return true
		}
	}
	return false
}
