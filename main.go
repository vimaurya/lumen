package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"
	"os/signal"
)

//go:embed resources/visit.jpg
var pixelByte []byte

func main() {
	err := InitDB()
	if err != nil {
		log.Panicf("failed to init db : %v", err)
	}

	done := make(chan bool)
	go func() {
		StartWorker()
		done <- true
	}()

	mux := http.DefaultServeMux

	mux.Handle("/visit.jpg", AnalyticsMiddleware(http.HandlerFunc(pixelHandler)))
	mux.HandleFunc("/admin", DashboardHandler)

	wrappedMux := AnalyticsMiddleware(mux)

	server := &http.Server{
		Addr:    ":8080",
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
