package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Limiter struct {
	mu         sync.Mutex
	tokens     float64
	burst      float64
	rate       float64
	lastRefill time.Time
}

func NewLimiter(r float64, burst int) *Limiter {
	return &Limiter{
		tokens:     float64(burst),
		burst:      float64(burst),
		rate:       r,
		lastRefill: time.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.tokens += elapsed * l.rate

	if l.tokens > l.burst {
		l.tokens = l.burst
	}
	l.lastRefill = now

	if l.tokens >= 1 {
		l.tokens--
		return true
	}

	return false
}

var limiter = NewLimiter(1, 3)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func RateLimiter(w http.ResponseWriter, r *http.Request) {

	if !limiter.Allow() {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{
			"error": "too many requests",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "hello! your request was allowed",
	})
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", RateLimiter)

	fmt.Println("Server running on http://localhost:8080")

	err := http.ListenAndServe(":8080", router)

	if err != nil {
		slog.Error("error found", slog.String("error", err.Error()))
	}

}
