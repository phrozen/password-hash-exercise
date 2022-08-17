package store

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Memory implements an in-memory key/value store with
// integer autoincrement keys and delayed writes.
// Internally uses a map for data storage instead of
// a slice for different reasons:
// 		+ Ease of implementation (deterministic index)
// 		+ Memory allocation time is predictable
//		+ Future proofing (deletes)
// Slices and channels would be the straight forward,
// solution but implementation has a lot of caveats:
// 		+ When waiting time.Sleep inside goroutines is
// 		non deterministic, scheduler might finish a
//		routine faster and thus append in the wrong
// 		order. This can be solved by tracking the size
//		of the slice and grow on demand, runtime should
//		be responsible for this, not the developer.
// 		+ Working with buffered channels means, ideally,
//		the write to the channel is done inside a select
//		clause with a default statement to avoid blocking,
//		I would personally use this approach if latency
//		was unpredictable or we want to have a limit on
//		pending writes, but we have fixed delay and the
//		sync.WaitGroup can easily be customized to achieve
//		a similar limiter functionality.
//		+ If deletes become a requirement, slices just
//		fall way behind in terms of performance due to
//		memory management.
// The benefits of slices do not outweight the added complexity
// go maps have O(1) amortized complexity for insert and lookups
// on finite space domains (integer), making them performant.
type Memory struct {
	sync.WaitGroup
	sync.RWMutex
	count int64
	data  map[int][]byte
	delay time.Duration
}

// NewMemory creates a new store with 'delay' writes.
// It is useful to allow the caller to setup the delay,
// specially for testing as we avoid mocking.
func NewMemory(delay time.Duration) *Memory {
	return &Memory{
		data:  make(map[int][]byte),
		delay: delay,
	}
}

// Get returns the value at index id or an error otherwise
func (m *Memory) Get(id int) ([]byte, error) {
	// [FIX] Use the mutex to avoid reading on concurrent writes
	m.Lock()
	defer m.Unlock()
	if val, ok := m.data[id]; ok {
		return val, nil
	}
	return nil, errors.New("Not Found")
}

// Set saves the value and returns the index where data will
// be written after delay. It does not block but fires a
// routine and keeps track of it via sync.WaitGroup, it uses
// atomic increase on the index, and a sync.Mutex when writing
// on a map for memory safety (concurrency).
func (m *Memory) Set(value []byte) (int, error) {
	// Atomically increment the counter to get a
	// consistent index snapshot
	index := atomic.AddInt64(&m.count, 1)
	// For tracking pending writes
	m.Add(1)
	go func(key int64, val []byte) {
		defer m.Done()
		// Sleep(delay) as per the requirements
		time.Sleep(m.delay)
		// Lock during the write for memory safety
		// due to concurrency (just in case)
		m.Lock()
		m.data[int(key)] = val
		m.Unlock()
	}(index, value)
	return int(index), nil
}

// Close blocks until all pending write operations are done
// Useful if data would be persisted, otherwise just a nice
// "to have" in case other implementations are done.
func (m *Memory) Close() error {
	m.Wait()
	return nil
}
