package service

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/phrozen/password-hash-exercise/internal/app"
	"github.com/phrozen/password-hash-exercise/internal/middleware/logger"
	"github.com/phrozen/password-hash-exercise/internal/stats"
	"github.com/phrozen/password-hash-exercise/internal/store"
)

var (
	getHashRe = regexp.MustCompile(`^\/hash\/(\d+)$`)
	setHashRe = regexp.MustCompile(`^\/hash[\/]*$`)
)

// HashingService implements Service and provides all request handlers
type HashingService struct {
	application *app.App
	logging     bool
	quit        chan bool
	router      *http.ServeMux
	statistics  *stats.Stats
}

func NewHashingService(delay time.Duration, logging bool) *HashingService {
	s := &HashingService{
		application: app.New(store.NewMemory(delay)),
		logging:     logging,
		quit:        make(chan bool),
		router:      http.NewServeMux(),
		statistics:  stats.New(),
	}
	s.setup()
	return s
}

// Handler will return the service http handler wrapped
// around with a logger middleware based on configuration
func (s *HashingService) Handler() http.Handler {
	if s.logging {
		return logger.Logger(s.router)
	}
	return s.router
}

// Shutdown returns the service's shutdown signaling channel
func (s *HashingService) Shutdown() chan bool {
	return s.quit
}

// Close performs teardown operations for the service
func (s *HashingService) Close() error {
	close(s.quit)
	return s.application.Close()
}

// Setup routes and handlers for the service
func (s *HashingService) setup() {
	// Fix for 301 Redirect for POST requests without trailing slash
	// On POST /hash Go default mux will try to redirect to /hash/
	// but then the POST call gets downgraded to GET and Body is lost
	s.router.HandleFunc("/hash", s.hashHandler)
	s.router.HandleFunc("/hash/", s.hashHandler)
	s.router.HandleFunc("/stats", s.statsHandler)
	s.router.HandleFunc("/shutdown", s.shutdownHandler)
}

func (s *HashingService) hashHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /hash/<id:int>
		if !getHashRe.MatchString(r.URL.Path) {
			http.Error(w, "GET /hash/<id:int>", http.StatusBadRequest)
			return
		}
		s.getHash(w, r)
	case http.MethodPost:
		//POST /hash Form(password:<string>)
		if !setHashRe.MatchString(r.URL.Path) {
			http.Error(w, "POST /hash Form(password=<string>)", http.StatusBadRequest)
			return
		}
		// Calculate statistics of time to process POST requests
		// Requirements are ambiguous, what does "process" mean in this context
		// Does delay account in process time? only successful requests?
		// Checking all requests to POST /hash for statistics to avoid tight coupling
		// It might be a better idea to move it to application layer if requirements change
		start := time.Now()
		s.postHash(w, r)
		s.statistics.Add(start)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *HashingService) getHash(w http.ResponseWriter, r *http.Request) {
	// Regular expression matching should prevent first two checks from ever failing
	matches := getHashRe.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		// Should never fail due to regexp matching
		http.Error(w, "GET /hash/<id:int>", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(matches[1])
	if err != nil {
		// Should never fail due to regexp matching
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Only error that can be returned is "Not Found"
	hash, err := s.application.GetHash(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write([]byte(hash))
}

func (s *HashingService) postHash(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	if password == "" {
		http.Error(w, "POST /hash Form(password=<string>)", http.StatusBadRequest)
		return
	}
	id, err := s.application.SetHash(password)
	if err != nil {
		// Never fails with memory store
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%d", id)
}

func (s *HashingService) statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	data, err := s.statistics.JSON()
	if err != nil {
		// Should never fail
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Sends the shutdown signal to the quit channel, in this case to the HTTP Server
// to beging the graceful shutdown process.
func (s *HashingService) shutdownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// We write inside a select clause to the unbuffered channel quit, so on subsecuent
	// requests we don't block the operation (channel write with no reader) and return
	// an error instead. Most likely the server will stop incoming connections before
	// that ever happens.
	select {
	case s.quit <- true:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Shutting down the server."))
	default:
		http.Error(w, "Shutdown in progress...", http.StatusConflict)
	}
}
