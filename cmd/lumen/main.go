package main

import (
	_ "embed"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"

	"github.com/oschwald/geoip2-golang"
	"github.com/ua-parser/uap-go/uaparser"
	"github.com/vimaurya/lumen/internal/analytics"
	"github.com/vimaurya/lumen/internal/storage"
	"github.com/vimaurya/lumen/internal/ui"
)

func main() {
	err := storage.InitDB()
	if err != nil {
		log.Panicf("failed to init db : %v", err)
	}

	analytics.GeoDB, err = geoip2.Open("./resources/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}

	defer analytics.GeoDB.Close()

	analytics.ClientParser, err = uaparser.New()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	go func() {
		storage.StartWorker()
		done <- true
	}()

	target, _ := url.Parse("http://localhost:8081")
	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/visit", visitHandler)
	mux.HandleFunc("/admin", ui.DashboardHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	wrappedMux := analytics.AnalyticsMiddleware(mux)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: wrappedMux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("shutting down...")

	server.Close()

	close(analytics.HitBuffer)

	<-done

	log.Println("flushed...")
}

func visitHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello visitor"))
}
