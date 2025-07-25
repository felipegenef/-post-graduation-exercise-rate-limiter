package main

import (
	"log"
	"net/http"

	"github.com/felipegenef/post-graduation-exercise-rate-limiter/config"
	"github.com/felipegenef/post-graduation-exercise-rate-limiter/limiter"
	"github.com/felipegenef/post-graduation-exercise-rate-limiter/middleware"
)

func main() {
	cfg := config.Load()
	store := limiter.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome!"))
	})

	handler := middleware.RateLimitMiddleware(store, cfg)(mux)

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
