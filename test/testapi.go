package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var port string
	if len(os.Args) > 1 {
		command := os.Args[1]

		switch command {
		case "port":
			port = os.Args[2]
		default:
			log.Printf("unknown command: %s", command)
			return
		}
	}

	host := "localhost:" + port
	lumensecret := "verysecret45"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received: %s %s", r.Method, r.URL.Path)
		secret := r.Header.Get("X-Lumen-Secret")
		log.Printf("Secret Header: %s", secret)

		if secret != lumensecret {
			http.Error(w, "access denied, requests must go through lumen", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"message": "Hello from Test API",
			"received_path": "%s",
			"lumen_secret": "%s",
			"all_headers": %+v
		}`, r.URL.Path, r.Header.Get("X-Lumen-Secret"), r.Header)
	})

	log.Printf("Test API listening on %s", host)
	log.Fatal(http.ListenAndServe(host, nil))
}
