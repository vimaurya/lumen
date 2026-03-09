package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vimaurya/lumen/internal/analytics"
	"github.com/vimaurya/lumen/internal/balancer"
	"github.com/vimaurya/lumen/internal/config"
	"github.com/vimaurya/lumen/internal/security"
	"github.com/vimaurya/lumen/internal/ui"
)

var server *http.Server

func startServer(cfg *config.Config) {
	for _, p := range cfg.Proxy {
		b := &balancer.Balancer{}
		for _, target := range p.Targets {
			u, err := url.Parse(target)
			if err != nil {
				log.Fatalf("Invalid target URL %s : %v", target, err)
			}

			target := balancer.Target{
				URL:              u,
				FailureThreshold: 5,
				Timeout:          3600 * time.Second,
			}

			b.Targets = append(b.Targets, &target)
		}
		balancer.Balancers[p.Prefix] = b
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

		target := balancer.Balancers[matched.Prefix].Next()

		ctx := context.WithValue(req.Context(), "lumen-prefix", matched.Prefix)
		ctx = context.WithValue(req.Context(), "lumen-target", target.String())

		*req = *req.WithContext(ctx)
		*req = *req.WithContext(ctx)

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
