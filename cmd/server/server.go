package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/phrozen/password-hash-exercise/internal/service"
)

// Server abstracts the HTTP operation layer and provides
// transparent graceful shutdown and teardown.
type Server struct {
	service service.Service
}

// NewHTTPServer creates a new HTTP server with the given service backend
func NewHTTPServer(svc service.Service) *Server {
	return &Server{service: svc}
}

// Run makes the server listen for requests asynchronously on port, and also
// handles all signaling required for graceful shutdown.
// This is the only place in the entire project which prints messages to the
// console, as it should be, and errors are logged normally instead of stoping
// execution (log.Fatal) to avoid side effects from the shutdown process.
func (s *Server) Run(port string) {
	// Create a new HTTP Server with the given port and service handler
	server := &http.Server{
		Addr:    ":" + port,
		Handler: s.service.Handler(),
	}
	// Start listening for http requests on a go routine
	go func() {
		log.Println("Server listening on port:", port)
		if err := server.ListenAndServe(); err != nil {
			log.Println("Shutting down the server...", err)
		}
	}()
	// Create a new notification channel to listen to os.Signal
	// interruptions like Ctrl+C to gracefully shutdown
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)
	// Block on either an OS interrupt or the service.Shutdown signal
	// to start graceful shutdown.
	select {
	case <-interrupt:
		log.Println("Received OS shutdown signal (Ctrl+C)")
	case <-s.service.Shutdown():
		log.Println("Received service shutdown signal: (GET /shutdown)")
	}
	// Shutdown the server gracefully, stop incoming connections and use
	// context cancelation to wait for at most 10 seconds for pending connection
	// to finish before forcefully closing them
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Println("Server shutdown forcefully after timeout:", err)
	}
	// Let's not forget to run the service teardown process which will run
	// the Close() chain. Service -> Application -> Store
	log.Println("Service teardown in progress...")
	if err := s.service.Close(); err != nil {
		log.Println("Service close error:", err)
	}
	log.Println("Done!")
}
