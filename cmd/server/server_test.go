package main

import (
	"sync"
	"testing"
	"time"

	"github.com/phrozen/password-hash-exercise/internal/service"
)

// This test only purpose is to test shutdown
// signaling into the server. For actual service
// tests, refer to the internal/service package.
func TestGracefulShutdown(t *testing.T) {
	// Create a new mock service for testing
	svc := service.NewMockService()
	// Let's grab the service signal channel to test
	quit := svc.Shutdown()
	s := NewHTTPServer(svc)
	// Use a wait group to track the server execution
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Run the server
		s.Run("3000")
		wg.Done()
	}()
	// Wait a bit for server initialization to complete
	time.Sleep(100 * time.Millisecond)
	// Send the shutdown signal
	quit <- true
	close(quit)
	// Wait for server to shutdown
	wg.Wait()
	// Test is successful if it does not hang and timeout
}
