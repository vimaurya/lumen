package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/oschwald/geoip2-golang"
	"github.com/ua-parser/uap-go/uaparser"
	"github.com/vimaurya/lumen/internal/analytics"
	"github.com/vimaurya/lumen/internal/config"
	"github.com/vimaurya/lumen/internal/storage"
)

func main() {
	if len(os.Args) > 1 {
		command := os.Args[1]

		switch command {
		case "init":
			runInit()
			return
		case "version":
			if info, ok := debug.ReadBuildInfo(); ok {
				fmt.Printf("Lumen Version: %s\n", info.Main.Version)
				fmt.Printf("Go Version:    %s\n", info.GoVersion)
				return
			}
			fmt.Println("Lumen v0.1.0 (build info unavailable)")
			return
		}
	}

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

	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("error loading config : %v", err)
	}

	startServer(cfg)

	done := make(chan bool)
	go func() {
		storage.StartWorker()
		done <- true
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
