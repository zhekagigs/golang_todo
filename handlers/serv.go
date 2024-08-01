package handlers

import (
	"context"
	"net/http"
)

type HTTPServer interface {
	ListenAndServe(addr string, handler http.Handler) error
	Shutdown(ctx context.Context) error
}

type RealHTTPServer struct {
	server *http.Server
}

func (s *RealHTTPServer) ListenAndServe(addr string, handler http.Handler) error {

	return http.ListenAndServe(addr, handler)
}

func (s *RealHTTPServer) Shutdown(ctx context.Context) error {

	return s.server.Shutdown(ctx)
}
