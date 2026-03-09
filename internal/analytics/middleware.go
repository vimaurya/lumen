package analytics

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vimaurya/lumen/internal/config"
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
	Method          string
	RequestSize     int
	SessionId       string
	IsBot           bool
}

type statusWriter struct {
	http.ResponseWriter
	Status int
}

func generateHash(ip, ua string) string {
	salt := time.Now().Format("2006-01-02")
	hash := sha256.Sum256([]byte(ip + ua + salt))
	return fmt.Sprintf("%x", hash)
}

func AnalyticsMiddleware(next http.Handler, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		start := time.Now()

		sw := &statusWriter{ResponseWriter: w, Status: http.StatusOK}

		next.ServeHTTP(sw, r)

		// if targetURL, ok := r.Context().Value("lumen-target").(string); ok {
		// 	prefix, _ := r.Context().Value("lumen-prefix").(string)
		// 	fmt.Println("this is the prefix : ", prefix)
		// 	balancer.RecordStatus(prefix, targetURL, sw.Status)
		// }

		duration := time.Since(start).Milliseconds()

		ua := r.Header.Get("User-Agent")
		// ip := ngrokextractIP(r)

		ip := ExtractIP(r.RemoteAddr)

		ref := r.Header.Get("Referer")
		method := r.Method

		requestSize := r.ContentLength
		go func() {
			visitorId := generateHash(ip, ua)

			client := ClientParser.Parse(ua)

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
				Method:          method,
				RequestSize:     int(requestSize),
				SessionId:       getSessionId(ip, ua),
				IsBot:           isBot(ua),
			})
		}()
	})
}

func (w *statusWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}
