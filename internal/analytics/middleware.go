package analytics

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
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

func generateHash(ip, ua string) string {
	salt := time.Now().Format("2006-01-02")
	hash := sha256.Sum256([]byte(extractIP(ip) + ua + salt))
	return fmt.Sprintf("%x", hash)
}

func AnalyticsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		ignoredExtensions := map[string]bool{
			".js":    true,
			".css":   true,
			".map":   true,
			".png":   true,
			".jpg":   true,
			".jpeg":  true,
			".ico":   true,
			".svg":   true,
			".woff":  true,
			".woff2": true,
			".json":  true,
		}

		ext := ""
		if dot := strings.LastIndex(path, "."); dot != -1 {
			ext = strings.ToLower(path[dot:])
		}

		if path == "/admin" ||
			ignoredExtensions[ext] {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		sw := &statusWriter{ResponseWriter: w, Status: http.StatusOK}

		next.ServeHTTP(sw, r)

		log.Printf("path : %s status : %v", path, sw.Status)

		duration := time.Since(start).Milliseconds()

		ua := r.Header.Get("User-Agent")
		ip := r.RemoteAddr
		ref := r.Header.Get("Referrer")

		visitorId := generateHash(ip, ua)

		client := ClientParser.Parse(ua)

		go func() {
			Collect(Hit{
				Path:            path,
				HashedUserId:    visitorId,
				Referrer:        ref,
				Timestamp:       time.Now().Unix(),
				Browser:         client.UserAgent.Family,
				OperatingSystem: client.Os.Family,
				Device:          client.Device.Family,
				Country:         getCountry(extractIP(ip)),
				Status:          sw.Status,
				Duration:        duration,
			})
		}()
	})
}

func (w *statusWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}
