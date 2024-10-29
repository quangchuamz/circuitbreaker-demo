package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/hello", helloHandler)
	log.Println("Service B starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate occasional failures
	if time.Now().Unix()%3 == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Service B Error")
		return
	}
	fmt.Fprint(w, "Hello from Service B")
}
