/*
	Logger package provides a very simple logging middleware for testing
	and debugging purposes on the service routes. It implements a response
	observer to track down the response status codes from the service.

	No testing is provided as it is meant as a debug tool, in production,
	a real structured logger should be used and different logging targets
	should be provided based on requirements. But it serves as a real life
	example on how to implement such middleware, it is just out of scope.
*/
package logger

import (
	"log"
	"net/http"
	"time"
)

// ResponseObserver embeds a ResponseWriter for logging purposes
type ResponseObserver struct {
	http.ResponseWriter
	statusCode int
	start      time.Time
}

// WriteHeader implementation to track status code from the response
func (ro *ResponseObserver) WriteHeader(code int) {
	ro.statusCode = code
	ro.ResponseWriter.WriteHeader(code)
}

// Logger middleware for debugging purposes
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ro := &ResponseObserver{w, http.StatusOK, time.Now()}
		next.ServeHTTP(ro, r)
		log.Printf("[%d] %s %s %v", ro.statusCode, r.Method, r.URL.Path, time.Since(ro.start))
	})
}
