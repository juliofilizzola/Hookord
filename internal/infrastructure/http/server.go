package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/juliofiliizzola/hookord/internal/application"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type Server struct {
	srv *http.Server
}

func NewServer(port string, webhookService *application.WebhookService) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", webhookService.HandleWebhook)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	})
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{
		srv: &http.Server{
			Addr:    ":" + port,
			Handler: mux,
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
