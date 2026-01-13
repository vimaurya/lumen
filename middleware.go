package main

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type Hit struct {
	Path            string
	HashedUserId    string
	Referrer        string
	Country         string
	Browser         string
	Device          string
	Duration        int64
	OperatingSystem string
	Status          int
	Timestamp       int64
}

type statusWriter struct {
	http.ResponseWriter
	Status int
}

var GeoDB *geoip2.Reader

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

func getCountry(ip string) string {
	parsedIP := net.ParseIP(extractIP(ip))

	if parsedIP == nil {
		return "Unknown"
	}

	record, err := GeoDB.City(parsedIP)
	if err != nil {
		return "Unknown"
	}

	if name, ok := record.Country.Names["en"]; ok {
		return name
	}

	return record.Country.IsoCode
}

func AnalyticsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/admin" ||
			strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".ico") ||
			strings.HasSuffix(path, ".css") {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		sw := &statusWriter{ResponseWriter: w, Status: http.StatusOK}

		next.ServeHTTP(sw, r)

		duration := time.Since(start).Milliseconds()

		ua := r.Header.Get("User-Agent")
		ip := r.RemoteAddr
		ref := r.Header.Get("Referrer")

		visitorId := generateHash(ip, ua)

		client := clientParser.Parse(ua)

		go func() {
			Collect(Hit{
				Path:            path,
				HashedUserId:    visitorId,
				Referrer:        ref,
				Timestamp:       time.Now().Unix(),
				Browser:         client.UserAgent.Family,
				OperatingSystem: client.Os.Family,
				Device:          client.Device.Family,
				Country:         getCountry(ip),
				Status:          sw.Status,
				Duration:        duration,
			})
		}()
	})
}
