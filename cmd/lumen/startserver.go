package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/vimaurya/lumen/internal/analytics"
	"github.com/vimaurya/lumen/internal/config"
	"github.com/vimaurya/lumen/internal/security"
	"github.com/vimaurya/lumen/internal/ui"
)

var server *http.Server

func startServer(cfg *config.Config) {
	target, _ := url.Parse(cfg.Proxy[0].Target)
	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

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
		proxy.ServeHTTP(w, r)
	})

	wrappedMux := analytics.AnalyticsMiddleware(mux, cfg)

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
