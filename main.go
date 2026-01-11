package main

import "net/http"

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("this is analytics!"))
	})

	wrappedMux := AnalyticsMiddleware(mux)

	http.ListenAndServe(":8080", wrappedMux)
}
