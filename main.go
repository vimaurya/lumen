package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	err := InitDB()
	if err != nil {
		log.Panicf("failed to init db : %w", err)
	}

	done := make(chan bool)
	go func() {
		StartWorker()
		done <- true
	}()

	server := &http.Server{
		Addr:    ":8080",
		Handler: AnalyticsMiddleware(http.DefaultServeMux),
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
