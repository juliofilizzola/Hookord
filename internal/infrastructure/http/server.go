package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const maxWebhookBodySize = 25 << 20

type Config struct {
	Port      string
	APIKey    string
	RateLimit RateLimitConfig
}

type Server struct {
	srv *http.Server
}

func NewServer(cfg Config, webhook http.Handler) *Server {
	limiter := NewRateLimiter(cfg.RateLimit)
	protect := func(h http.Handler) http.Handler {
		return limiter.Middleware(APIKey(cfg.APIKey, h))
	}

	if cfg.APIKey == "" {
		log.Warn().Msg("API_KEY not set: /webhook and /metrics are not protected by api key")
	}

	mux := http.NewServeMux()
	mux.Handle("/webhook", protect(MaxBodySize(maxWebhookBodySize, webhook)))
	mux.Handle("/metrics", protect(promhttp.Handler()))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	})

	return &Server{
		srv: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	log.Info().Str("addr", s.srv.Addr).Msg("starting server")
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("shutting down server...")
	return s.srv.Shutdown(ctx)
}
