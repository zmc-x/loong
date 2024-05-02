package trafficgate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
	shutdownTimeOut time.Duration
}

func NewServer(rawCfg any) (*Server, error) {
	cfg := Config{}
	err := json.Unmarshal(rawCfg.([]byte), &cfg)
	if err != nil {
		return nil, err
	}
	return &Server{
		server: &http.Server{
			Addr: fmt.Sprintf(":%d", cfg.Port),
			ReadTimeout: 5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler: http.NewServeMux(),
		},
		shutdownTimeOut: 10 * time.Second,
	}, nil
}

func (s *Server) RegisterHandler(path string, handler http.Handler) {
	s.server.Handler.(*http.ServeMux).Handle(path, handler)
}

func (s *Server) StartServer() error {
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err 
	}
	return nil
}

func (s *Server) ShutdownServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeOut)
	defer cancel()
	return s.server.Shutdown(ctx)
}