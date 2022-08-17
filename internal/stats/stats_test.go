package stats

import (
	"encoding/json"
	"runtime"
	"testing"
	"time"
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

func TestStats(t *testing.T) {
	s := New()
	// Test request counter
	start := time.Now()
	for i := 1; i <= 100; i++ {
		s.Add(start)
		equal(t, int64(i), s.requests)
	}
	// Create a sample response object for
	// comparison with data from above loop
	want := Response{
		Total:   s.requests,
		Average: s.elapsed / s.requests,
	}
	// Get the JSON and Unmarshal for compare
	have := Response{}
	data, err := s.JSON()
	equal(t, true, err == nil)
	err = json.Unmarshal(data, &have)
	equal(t, true, err == nil)
	equal(t, want, have)
}
