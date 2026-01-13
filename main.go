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
)

//go:embed resources/visit.jpg
var pixelByte []byte

var clientParser *uaparser.Parser

func main() {
	err := InitDB()
	if err != nil {
		log.Panicf("failed to init db : %v", err)
	}

	GeoDB, err = geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}

	defer GeoDB.Close()

	clientParser, err = uaparser.New()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	go func() {
		StartWorker()
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

	mux.HandleFunc("/visit.jpg", pixelHandler)
	mux.HandleFunc("/admin", DashboardHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	wrappedMux := AnalyticsMiddleware(mux)

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

	close(HitBuffer)

	<-done

	log.Println("flushed...")
}

func pixelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Write(pixelByte)
}
