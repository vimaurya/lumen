package security

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vimaurya/lumen/internal/config"
)

type visitor struct {
	count    int
	lastSeen time.Time
	limit    int
}

var (
	counter = make(map[string]*visitor)
	mu      sync.Mutex
)

func upLimit(ip string) bool {
	defaultLimit := 15000
	coolDownPeriod := 3 * time.Second
	bonus := 5000
	mu.Lock()
	defer mu.Unlock()

	v, ok := counter[ip]
	if !ok {
		v = &visitor{
			limit: defaultLimit,
		}
		counter[ip] = v
	}

	if v.count > v.limit {
		log.Print("already rate limited")
		if time.Since(v.lastSeen) > coolDownPeriod {
			v.count -= bonus
			if v.count < 0 {
				v.count = 0
			}

			v.lastSeen = time.Now()
			return false
		}
	}

	v.lastSeen = time.Now()
	v.count++

	return v.count > v.limit
}

func RateLimiter(next http.Handler, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		path := r.URL.Path

		ext := ""
		if dot := strings.LastIndex(path, "."); dot != -1 {
			ext = strings.ToLower(path[dot:])
		}

		if path == cfg.Server.AdminPath ||
			cfg.Security.IgnoredExtensions[ext] {
			next.ServeHTTP(w, r)
			return
		}

		if upLimit(ip) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func getIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	return r.RemoteAddr
}

func CleanUpVisitors() {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			mu.Lock()
			for ip, v := range counter {
				if time.Since(v.lastSeen) > 5*time.Minute {
					delete(counter, ip)
				}
			}
			mu.Unlock()
		}
	}()
}
