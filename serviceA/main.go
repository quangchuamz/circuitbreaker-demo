package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
)

var (
	cb     *gobreaker.CircuitBreaker
	client *redis.Client
)

func init() {
	// Initialize Redis client
	client = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	var st gobreaker.Settings
	st.Name = "ServiceB"
	st.MaxRequests = 3
	st.Interval = time.Duration(5) * time.Second
	st.Timeout = time.Duration(10) * time.Second
	st.ReadyToTrip = func(counts gobreaker.Counts) bool {
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 3 && failureRatio >= 0.6
	}

	cb = gobreaker.NewCircuitBreaker(st)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/call-service-b", callServiceBHandler)
	log.Printf("Service A starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func callServiceBHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Check if the circuit is open in Redis
	isOpen, err := client.Get(ctx, "circuit:ServiceB:open").Bool()
	if err != nil && err != redis.Nil {
		log.Printf("Error checking circuit state: %v", err)
	}

	if isOpen {
		http.Error(w, "Service B is unavailable (Circuit Open)", http.StatusServiceUnavailable)
		return
	}

	result, err := cb.Execute(func() (interface{}, error) {
		return callServiceB()
	})
	if err != nil {
		log.Printf("Circuit breaker error: %v", err)
		// Set circuit to open in Redis
		err = client.Set(ctx, "circuit:ServiceB:open", true, 10*time.Second).Err()
		if err != nil {
			log.Printf("Error setting circuit state: %v", err)
		}
		http.Error(w, "Service B is unavailable", http.StatusServiceUnavailable)
		return
	}

	// Reset circuit state in Redis
	err = client.Set(ctx, "circuit:ServiceB:open", false, 0).Err()
	if err != nil {
		log.Printf("Error resetting circuit state: %v", err)
	}

	fmt.Fprint(w, result)
}

func callServiceB() (interface{}, error) {
	resp, err := http.Get("http://service-b:8081/hello")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service B returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return string(body), nil
}
