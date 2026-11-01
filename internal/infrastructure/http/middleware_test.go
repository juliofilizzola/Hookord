package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestAPIKey(t *testing.T) {
	handler := APIKey("s3cret", okHandler)

	tests := []struct {
		name  string
		setup func(r *http.Request)
		want  int
	}{
		{"missing key", func(r *http.Request) {}, http.StatusUnauthorized},
		{"wrong key", func(r *http.Request) { r.Header.Set(APIKeyHeader, "nope") }, http.StatusUnauthorized},
		{"header", func(r *http.Request) { r.Header.Set(APIKeyHeader, "s3cret") }, http.StatusOK},
		{"bearer", func(r *http.Request) { r.Header.Set("Authorization", "Bearer s3cret") }, http.StatusOK},
		{"query param", func(r *http.Request) { r.URL.RawQuery = APIKeyQueryParam + "=s3cret" }, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
			tt.setup(req)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != tt.want {
				t.Errorf("status = %d, want %d", rr.Code, tt.want)
			}
		})
	}
}

func TestAPIKey_EmptyKeyDisablesCheck(t *testing.T) {
	rr := httptest.NewRecorder()
	APIKey("", okHandler).ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/webhook", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRateLimiter_BlocksAfterBurst(t *testing.T) {
	handler := NewRateLimiter(RateLimitConfig{RequestsPerSecond: 0.001, Burst: 2}).Middleware(okHandler)

	codes := make([]int, 0, 3)
	for range 3 {
		req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		codes = append(codes, rr.Code)
	}

	if codes[0] != http.StatusOK || codes[1] != http.StatusOK || codes[2] != http.StatusTooManyRequests {
		t.Errorf("codes = %v, want [200 200 429]", codes)
	}
}

func TestRateLimiter_PerClient(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{RequestsPerSecond: 0.001, Burst: 1})
	if !rl.Allow("1.1.1.1") {
		t.Fatal("first request of client A should pass")
	}
	if rl.Allow("1.1.1.1") {
		t.Error("second request of client A should be limited")
	}
	if !rl.Allow("2.2.2.2") {
		t.Error("client B should have its own bucket")
	}
}

func TestRateLimiter_DisabledWithZeroRate(t *testing.T) {
	handler := NewRateLimiter(RateLimitConfig{RequestsPerSecond: 0}).Middleware(okHandler)
	for range 5 {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rr.Code)
		}
	}
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:5555"
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 10.0.0.2")

	if got := clientIP(req, false); got != "10.0.0.1" {
		t.Errorf("clientIP(untrusted) = %q, want 10.0.0.1", got)
	}
	if got := clientIP(req, true); got != "203.0.113.7" {
		t.Errorf("clientIP(trusted) = %q, want 203.0.113.7", got)
	}
}

func TestServer_ProtectedRoutes(t *testing.T) {
	srv := NewServer(Config{Port: "0", APIKey: "k"}, okHandler)
	ts := httptest.NewServer(srv.srv.Handler)
	defer ts.Close()

	for path, want := range map[string]int{
		"/health":            http.StatusOK,
		"/metrics":           http.StatusUnauthorized,
		"/metrics?api_key=k": http.StatusOK,
		"/webhook":           http.StatusUnauthorized,
		"/webhook?api_key=k": http.StatusOK,
	} {
		resp, err := ts.Client().Get(ts.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != want {
			t.Errorf("GET %s = %d, want %d", path, resp.StatusCode, want)
		}
	}
}
