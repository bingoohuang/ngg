package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	body := flag.String("body", "{}", "JSON body to return")
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintln(w, *body)
	})

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
}
