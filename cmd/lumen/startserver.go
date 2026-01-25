package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/vimaurya/lumen/internal/analytics"
	"github.com/vimaurya/lumen/internal/balancer"
	"github.com/vimaurya/lumen/internal/config"
	"github.com/vimaurya/lumen/internal/security"
	"github.com/vimaurya/lumen/internal/ui"
)

var server *http.Server

func startServer(cfg *config.Config) {
	balancers := make(map[string]*balancer.Balancer)

	for _, p := range cfg.Proxy {
		b := &balancer.Balancer{}
		for _, target := range p.Targets {
			u, err := url.Parse(target)
			if err != nil {
				log.Fatalf("Invalid target URL %s : %v", target, err)
			}
			b.Targets = append(b.Targets, u)
		}
		balancers[p.Prefix] = b
	}

	rp := &httputil.ReverseProxy{}

	rp.Director = func(req *http.Request) {
		var matched *config.Proxy

		for i := range cfg.Proxy {
			if strings.HasPrefix(req.URL.Path, cfg.Proxy[i].Prefix) {
				matched = &cfg.Proxy[i]
				break
			}
		}

		if matched == nil {
			return
		}

		target := balancers[matched.Prefix].Next()

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		req.Header.Set("X-Lumen-Secret", cfg.Security.LumenSecret)

		if !cfg.Proxy[0].PreservePath {
			newPath := strings.TrimPrefix(req.URL.Path, cfg.Proxy[0].Prefix)
			if !strings.HasPrefix(newPath, "/") {
				newPath = "/" + newPath
			}
			req.URL.Path = newPath
		}

		req.Host = target.Host
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/visit", visitHandler)
	mux.HandleFunc(cfg.Server.AdminPath, security.PasswordProtection(cfg, ui.DashboardHandler))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rp.ServeHTTP(w, r)
	})

	analyticsMux := analytics.AnalyticsMiddleware(mux, cfg)

	wrappedMux := security.RateLimiter(analyticsMux, cfg)

	security.CleanUpVisitors()

	server = &http.Server{
		Addr:    "localhost:" + strconv.Itoa(cfg.Server.Port),
		Handler: wrappedMux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()
}
