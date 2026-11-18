package middleware

import (
	"sync"

	"Afaisal94/first-app/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter menyimpan rate limiter per alamat IP.
type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

// newIPLimiter membuat instance rate limiter per IP dengan rate dan burst yang ditentukan.
func newIPLimiter(rps float64, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// get mengambil (atau buat) limiter untuk IP tertentu.
func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(l.rps, l.burst)
		l.limiters[ip] = limiter
	}
	return limiter
}

// RateLimit membatasi jumlah request per IP (token bucket).
// rps = request per detik yang diizinkan (steady state).
// burst = jumlah request instan yang diizinkan sebelum throttle.
//
// Contoh: RateLimit(5, 10) → 5 req/detik sustained, burst 10 req sekaligus.
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	limiter := newIPLimiter(rps, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.get(ip).Allow() {
			response.JSON(c, 429, false, "Too many requests. Please slow down.", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
