package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/phrozen/password-hash-exercise/internal/app"
	"github.com/phrozen/password-hash-exercise/internal/stats"
)

// WARNING: Don't use this, use testify instead!
// https://github.com/stretchr/testify
// This is only good if you are limited to std library
// and know what you are doing...
func equal(t *testing.T, want, have any) {
	if want != have {
		_, f, l, _ := runtime.Caller(1)
		t.Errorf("\n%s:%d\n\t%s\texpected: %v - got: %v", f, l, t.Name(), want, have)
	}
}

// Function alias to improve readability
var request = httptest.NewRequest

// Helper function to avoid code duplication and improve readability
func serve(svc Service, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	svc.Handler().ServeHTTP(rec, req)
	return rec
}

func TestSuccess(t *testing.T) {
	s := NewHashingService(0, false)
	defer s.Close()
	// Decide on a number of rounds for the test
	rounds := 100
	// Test for POST success and index value
	for i := 1; i <= rounds; i++ {
		body := fmt.Sprintf("password=value%d", i)
		req := request(http.MethodPost, "/hash", strings.NewReader(body))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res := serve(s, req)
		equal(t, http.StatusOK, res.Result().StatusCode)
		equal(t, fmt.Sprintf("%d", i), res.Body.String())
	}
	// Make sure all pending writes are done
	time.Sleep(25 * time.Millisecond)
	// Test for GET success on the same index
	for i := 1; i <= rounds; i++ {
		res := serve(s, request(http.MethodGet, fmt.Sprintf("/hash/%d", i), nil))
		equal(t, http.StatusOK, res.Result().StatusCode)
		equal(t, app.HASH_LENGTH, len(res.Body.String()))
	}
	// Check stats
	res := serve(s, request(http.MethodGet, "/stats", nil))
	equal(t, http.StatusOK, res.Result().StatusCode)
	response := stats.Response{}
	err := json.Unmarshal(res.Body.Bytes(), &response)
	equal(t, nil, err)
	equal(t, rounds, int(response.Total))
}

func TestBadRequest(t *testing.T) {
	s := NewHashingService(0, false)
	defer s.Close()

	casesPost := map[string]string{
		"/hash/123": "password=valid", // Valid body
		"/hash/":    "password=",      // Valid path
		"/hash":     "invalid",        // Invalid payload
	}

	for path, payload := range casesPost {
		req := request(http.MethodPost, path, strings.NewReader(payload))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res := serve(s, req)
		equal(t, http.StatusBadRequest, res.Result().StatusCode)
	}

	casesGet := []string{
		"/hash",     // No id
		"/hash/",    // No id
		"/hash/abc", // Invalid id
		"/hash/-1",  // Invalid id
		"/hash/1.1", // Invalid id
		"/hash/2_3", // Invalid id
	}

	for _, path := range casesGet {
		res := serve(s, request(http.MethodGet, path, nil))
		equal(t, http.StatusBadRequest, res.Result().StatusCode)
	}
}

func TestNotAllowed(t *testing.T) {
	s := NewHashingService(0, false)
	defer s.Close()

	cases := map[string]string{
		"/stats":    http.MethodPost,
		"/hash/":    http.MethodPut,
		"/hash/1":   http.MethodDelete,
		"/shutdown": http.MethodPatch,
	}

	for path, method := range cases {
		res := serve(s, request(method, path, nil))
		equal(t, http.StatusMethodNotAllowed, res.Result().StatusCode)
	}
}

func TestNotFound(t *testing.T) {
	s := NewHashingService(0, false)
	defer s.Close()

	cases := map[string]string{
		"/hash/1": http.MethodGet, // Valid path
		"/foo":    http.MethodGet,
		"/bar":    http.MethodPost,
		"/baz":    http.MethodDelete,
	}

	for path, method := range cases {
		res := serve(s, request(method, path, nil))
		equal(t, http.StatusNotFound, res.Result().StatusCode)
	}
}

func TestShutdown(t *testing.T) {
	s := NewHashingService(0, false)
	defer s.Close()
	// Wait for the shutdown signal
	quit := false
	go func() {
		quit = <-s.Shutdown()
	}()
	// Let the go routine start
	time.Sleep(10 * time.Millisecond)
	// Shutdown the service
	res := serve(s, request(http.MethodGet, "/shutdown", nil))
	// Let the scheduler return the channel value to check results
	time.Sleep(10 * time.Millisecond)
	equal(t, http.StatusOK, res.Result().StatusCode)
	equal(t, true, quit)
	// Call again to get the blocking state of the channel write
	res = serve(s, request(http.MethodGet, "/shutdown", nil))
	equal(t, http.StatusConflict, res.Result().StatusCode)
}
