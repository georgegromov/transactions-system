package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type httpServer struct {
	httpServer *http.Server
}

func NewHttpServer(
	port, maxHeaderBytes int,
	readTimeout, writeTimeout, idleTimeout time.Duration,
	handler http.Handler,
) *httpServer {
	addr := fmt.Sprintf(":%d", port)
	return &httpServer{
		httpServer: &http.Server{
			Addr:           addr,
			Handler:        handler,
			ReadTimeout:    readTimeout,
			WriteTimeout:   writeTimeout,
			IdleTimeout:    idleTimeout,
			MaxHeaderBytes: maxHeaderBytes << 20,
		},
	}
}

func (s *httpServer) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *httpServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
