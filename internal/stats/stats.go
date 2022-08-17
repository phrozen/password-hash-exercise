package stats

import (
	"encoding/json"
	"sync/atomic"
	"time"
)

// Stats keeps track of the total number of requests and the
// total elapsed time in microseconds to calculate an average
type Stats struct {
	requests int64
	elapsed  int64
}

type response struct {
	Total   int64 `json:"total"`
	Average int64 `json:"average"`
}

// New returns Stats with default zero values can use &Stats{} instead
func New() *Stats {
	return &Stats{}
}

// Add calculates the delta in time between Now and start
// and atomically increases the stats counters
func (s *Stats) Add(start time.Time) {
	delta := time.Now().UnixMicro() - start.UnixMicro()
	atomic.AddInt64(&s.elapsed, delta)
	atomic.AddInt64(&s.requests, 1)
}

// JSON returns the JSON representation of the stats to
// be consumed by a handler as per the requirements
func (s *Stats) JSON() ([]byte, error) {
	avg := int64(0)
	// avoid division by zero
	if s.requests > 0 {
		avg = s.elapsed / s.requests
	}
	res := response{
		Total:   s.requests,
		Average: avg,
	}
	return json.Marshal(res)
}
