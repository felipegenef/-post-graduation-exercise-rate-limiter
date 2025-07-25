package limiter

type RateLimiter interface {
	Allow(key string, limit int, blockSeconds int) (bool, error)
}
