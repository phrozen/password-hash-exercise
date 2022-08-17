package main

import (
	"flag"
	"os"
	"time"

	"github.com/phrozen/password-hash-exercise/internal/service"
)

func main() {
	// Flags needed for runtime configuration, could abstract into
	// a config package if needed, but kept it here for readability.
	// Sensible defaults are either requirements or common practice.
	delay := flag.Duration("d", 5*time.Second, "Delay for writes")
	port := flag.String("p", os.Getenv("PORT"), "Listening port")
	logs := flag.Bool("l", true, "Enables logging")
	flag.Parse()
	// Create a new Hashing Service and feed it to the http server
	server := NewHTTPServer(service.NewHashingService(*delay, *logs))
	// Run will perform graceful shutdown
	server.Run(*port)
}
