package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/felipegenef/post-graduation-exercise-rate-limiter/config"
	"github.com/felipegenef/post-graduation-exercise-rate-limiter/limiter"
	"github.com/felipegenef/post-graduation-exercise-rate-limiter/middleware"
)

func overrideEnv(ipLimit, tokenLimit, blockSecs int) {
	os.Setenv("RATE_LIMIT_IP", strconv.Itoa(ipLimit))
	os.Setenv("RATE_LIMIT_TOKEN", strconv.Itoa(tokenLimit))
	os.Setenv("BLOCK_DURATION_SECONDS", strconv.Itoa(blockSecs))
}

func createTestServer(cfg *config.Config, store limiter.RateLimiter) *httptest.Server {
	handler := middleware.RateLimitMiddleware(store, cfg)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}),
	)
	return httptest.NewServer(handler)
}

// TestRateLimiterConcurrentTokensAndIPs
//
// EN: This test checks if the rate limiter handles multiple concurrent requests from different IPs and tokens.
// It ensures that the rate limits are enforced even under parallel load.
//
// PT: Este teste verifica se o rate limiter lida corretamente com múltiplas requisições concorrentes
// vindas de IPs e tokens diferentes. Ele garante que os limites sejam aplicados mesmo sob carga paralela.
func TestRateLimiterConcurrentTokensAndIPs(t *testing.T) {
	overrideEnv(5, 10, 3)
	cfg := config.Load()
	redisStore := limiter.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword)
	_ = redisStore.Client.FlushDB(redisStore.Ctx)

	server := createTestServer(cfg, redisStore)
	defer server.Close()

	var wg sync.WaitGroup
	client := &http.Client{}
	numRequests := 15

	tokens := []string{}
	ips := []string{}
	for i := 0; i < 10; i++ {
		tokens = append(tokens, "token-"+strconv.Itoa(i))
		ips = append(ips, "192.168.0."+strconv.Itoa(i))
	}

	for i := 0; i < 10; i++ {
		token := tokens[i]
		ip := ips[i]

		wg.Add(1)
		go func(token, ip string) {
			defer wg.Done()
			blockedCount := 0
			for j := 0; j < numRequests; j++ {
				req, _ := http.NewRequest("GET", server.URL, nil)
				req.Header.Set("API_KEY", token)
				req.RemoteAddr = ip + ":12345"

				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("Request failed: %v", err)
					continue
				}

				if j >= cfg.RateLimitToken && resp.StatusCode != http.StatusTooManyRequests {
					t.Errorf("Expected 429 for token %s at request %d, got %d", token, j+1, resp.StatusCode)
				}
				if resp.StatusCode == http.StatusTooManyRequests {
					blockedCount++
				}
				resp.Body.Close()

				time.Sleep(50 * time.Millisecond)
			}
			t.Logf("Token %s blocked %d times", token, blockedCount)
		}(token, ip)
	}

	wg.Wait()
}

// TestRateLimiterIPOnly
//
// EN: This test ensures that rate limiting works based solely on the client's IP address,
// without considering any tokens.
//
// PT: Este teste garante que o rate limiting funcione apenas com base no endereço IP do cliente,
// sem considerar nenhum token.
func TestRateLimiterIPOnly(t *testing.T) {
	overrideEnv(3, 10, 3)
	cfg := config.Load()
	redisStore := limiter.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword)
	_ = redisStore.Client.FlushDB(redisStore.Ctx)

	server := createTestServer(cfg, redisStore)
	defer server.Close()

	ip := "127.0.0.1"
	client := &http.Client{}

	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.RemoteAddr = ip + ":1111"

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if i >= cfg.RateLimitIP && resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected 429 for IP %s at request %d, got %d", ip, i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// TestRateLimiterTokenBlockingDuration
//
// EN: This test checks that a token is temporarily blocked after exceeding the rate limit,
// and that it is unblocked after the configured block duration.
//
// PT: Este teste verifica se um token é temporariamente bloqueado após exceder o limite,
// e se é desbloqueado após a duração de bloqueio configurada.
func TestRateLimiterTokenBlockingDuration(t *testing.T) {
	overrideEnv(1000, 3, 2)
	cfg := config.Load()
	redisStore := limiter.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword)
	_ = redisStore.Client.FlushDB(redisStore.Ctx)

	server := createTestServer(cfg, redisStore)
	defer server.Close()

	client := &http.Client{}
	token := "token-block-test"
	ip := "10.0.0.1"

	for i := 0; i < 4; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header.Set("API_KEY", token)
		req.RemoteAddr = ip + ":1234"
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if i < 3 && resp.StatusCode != http.StatusOK {
			t.Errorf("Expected OK for request %d, got %d", i+1, resp.StatusCode)
		}
		if i == 3 && resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected 429 on request %d, got %d", i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("API_KEY", token)
	req.RemoteAddr = ip + ":1234"
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected 429 during block, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	time.Sleep(time.Duration(cfg.BlockDurationSecs)*time.Second + 500*time.Millisecond)

	req, _ = http.NewRequest("GET", server.URL, nil)
	req.Header.Set("API_KEY", token)
	req.RemoteAddr = ip + ":1234"
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected OK after block period, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestRateLimiterIPBlockingDuration
//
// EN: Similar to the previous test, but this one focuses on blocking based on IP address
// instead of a token.
//
// PT: Semelhante ao teste anterior, mas este foca no bloqueio com base no endereço IP
// ao invés de token.
func TestRateLimiterIPBlockingDuration(t *testing.T) {
	overrideEnv(3, 1000, 2)
	cfg := config.Load()
	redisStore := limiter.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword)
	_ = redisStore.Client.FlushDB(redisStore.Ctx)

	server := createTestServer(cfg, redisStore)
	defer server.Close()

	client := &http.Client{}
	ip := "10.0.0.2"

	for i := 0; i < 4; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.RemoteAddr = ip + ":4321"
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if i < 3 && resp.StatusCode != http.StatusOK {
			t.Errorf("Expected OK for request %d, got %d", i+1, resp.StatusCode)
		}
		if i == 3 && resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected 429 on request %d, got %d", i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.RemoteAddr = ip + ":4321"
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected 429 during block, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	time.Sleep(time.Duration(cfg.BlockDurationSecs)*time.Second + 500*time.Millisecond)

	req, _ = http.NewRequest("GET", server.URL, nil)
	req.RemoteAddr = ip + ":4321"
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected OK after block period, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestRateLimiterTokenOverridesIPLimit
//
// EN: This test ensures that if a token with a higher limit is used, it overrides the stricter IP-based limit.
//
// PT: Este teste garante que, se um token com limite mais alto for usado, ele sobrepõe o limite mais restritivo baseado em IP.
func TestRateLimiterTokenOverridesIPLimit(t *testing.T) {
	overrideEnv(2, 5, 3)
	cfg := config.Load()
	redisStore := limiter.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword)
	_ = redisStore.Client.FlushDB(redisStore.Ctx)

	server := createTestServer(cfg, redisStore)
	defer server.Close()

	client := &http.Client{}
	token := "token-high-limit"
	ip := "10.1.1.1"

	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header.Set("API_KEY", token)
		req.RemoteAddr = ip + ":1234"

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// TestRateLimiterOnlyIPNoToken
//
// EN: This test verifies the behavior when no token is provided and only the IP limit is enforced.
//
// PT: Este teste verifica o comportamento quando nenhum token é fornecido e apenas o limite de IP é aplicado.
func TestRateLimiterOnlyIPNoToken(t *testing.T) {
	overrideEnv(3, 1000, 2)
	cfg := config.Load()
	redisStore := limiter.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword)
	_ = redisStore.Client.FlushDB(redisStore.Ctx)

	server := createTestServer(cfg, redisStore)
	defer server.Close()

	client := &http.Client{}
	ip := "10.2.2.2"

	for i := 0; i < 4; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.RemoteAddr = ip + ":5678"

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if i < 3 && resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 before hitting limit, got %d", resp.StatusCode)
		}
		if i == 3 && resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected 429 after hitting limit, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// TestRateLimiterResponseBodyOnLimit
//
// EN: This test checks the response body when the rate limit is exceeded,
// ensuring that the message returned is clear and expected.
//
// PT: Este teste verifica o corpo da resposta quando o limite é excedido,
// garantindo que a mensagem retornada seja clara e esperada.
func TestRateLimiterResponseBodyOnLimit(t *testing.T) {
	overrideEnv(1, 1, 2)
	cfg := config.Load()
	redisStore := limiter.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword)
	_ = redisStore.Client.FlushDB(redisStore.Ctx)

	server := createTestServer(cfg, redisStore)
	defer server.Close()

	client := &http.Client{}
	ip := "10.4.4.4"
	token := "test-body"

	req1, _ := http.NewRequest("GET", server.URL, nil)
	req1.Header.Set("API_KEY", token)
	req1.RemoteAddr = ip + ":1111"
	resp1, _ := client.Do(req1)
	resp1.Body.Close()

	req2, _ := http.NewRequest("GET", server.URL, nil)
	req2.Header.Set("API_KEY", token)
	req2.RemoteAddr = ip + ":1111"
	resp2, _ := client.Do(req2)

	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("Expected 429, got %d", resp2.StatusCode)
	}
	defer resp2.Body.Close()

	body, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := "you have reached the maximum number of requests or actions allowed within a certain time frame"
	if string(body) != expected {
		t.Errorf("Expected response body %q, got %q", expected, string(body))
	}
}
