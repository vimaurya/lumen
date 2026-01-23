package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := "localhost:8081"

	lumensecret := "verysecret5"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received: %s %s", r.Method, r.URL.Path)
		secret := r.Header.Get("X-Lumen-Secret")
		log.Printf("Secret Header: %s", secret)

		if secret != lumensecret {
			w.Write([]byte("{error : access denied. requests must go through lumen}"))
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

	log.Printf("Test API listening on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
