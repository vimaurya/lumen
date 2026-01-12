package main

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Hit struct {
	Path         string
	HashedUserId string
	Referrer     string
	Timestamp    int64
}

func generateHash(ip, ua string) string {
	salt := time.Now().Format("2006-01-02")
	hash := sha256.Sum256([]byte(extractIP(ip) + ua + salt))
	return fmt.Sprintf("%x", hash)
}

func extractIP(ip string) string {
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip
	}

	return host
}

func AnalyticsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		ua := r.Header.Get("User-Agent")
		ip := r.RemoteAddr
		ref := r.Header.Get("Referrer")

		visitorId := generateHash(ip, ua)

		func() {
			Collect(Hit{
				Path:         path,
				HashedUserId: visitorId,
				Referrer:     ref,
				Timestamp:    time.Now().Unix(),
			})
		}()
		next.ServeHTTP(w, r)
	})
}
