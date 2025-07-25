package middleware

import (
	"net/http"
	"strings"

	"github.com/felipegenef/post-graduation-exercise-rate-limiter/config"
	"github.com/felipegenef/post-graduation-exercise-rate-limiter/limiter"
)

func RateLimitMiddleware(store limiter.RateLimiter, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("API_KEY")
			ip := strings.Split(r.RemoteAddr, ":")[0]

			var key string
			var limit int
			//
			if token != "" {
				key = "token:" + token
				limit = cfg.RateLimitToken
			} else {
				key = "ip:" + ip
				limit = cfg.RateLimitIP
			}

			allowed, err := store.Allow(key, limit, cfg.BlockDurationSecs)
			if err != nil || !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
