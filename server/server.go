package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

// Server is a test server.
type Server struct {
	logger     *slog.Logger
	port       string
	httpServer *http.Server
}

// New creates a new server.
func New(logger *slog.Logger, port string) *Server {
	srv := Server{
		logger: logger,
		port:   port,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/test", srv.testHandler)

	srv.httpServer = &http.Server{
		Addr:    net.JoinHostPort("", port),
		Handler: mux,
	}
	return &srv
}

// Start starts the server.
func (srv *Server) Start(ctx context.Context) error {
	go func() {
		if err := srv.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srv.logger.Error("could not listen and serve", slog.String("error", err.Error()))
		}
	}()
	return nil
}

// Stop stops the server.
func (s *Server) Stop(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("could not shutdown http server: %w", err)
	}
	return nil
}

type testHandlerResponse struct {
	Items []string `json:"items"`
}

var testHandlerResp = testHandlerResponse{
	Items: []string{
		"foo",
		"bar",
		"baz",
	},
}

// testHandler handles requests to the /test endpoint.
func (srv *Server) testHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}

	jsonResp, err := json.Marshal(testHandlerResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
