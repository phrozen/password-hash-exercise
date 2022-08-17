package service

import "net/http"

// Service provides the adapter (abstraction) to be run
// by the HTTP server in the main package, it is a very
// simple wrapper around common service operations needed
// for this example.
type Service interface {
	// Should take care of all the routeing and handling
	Handler() http.Handler
	// Returns the channel the service can use to signal
	// the server for graceful shutdown.
	Shutdown() chan bool
	// Should peform all service teardown operations,
	// in real life this closes databases, files, etc...
	Close() error
}

// MockService for testing purposes, modify as needed
type MockService struct {
	quit chan bool
}

func NewMockService() *MockService {
	return &MockService{quit: make(chan bool)}
}

func (s *MockService) Handler() http.Handler {
	return http.NewServeMux()
}

func (s *MockService) Shutdown() chan bool {
	return s.quit
}

func (s *MockService) Close() error {
	return nil
}
