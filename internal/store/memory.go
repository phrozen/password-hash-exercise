package store

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Memory implements an in-memory key/value store with
// integer autoincrement keys and delayed writes
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

func (m *Memory) Get(id int) ([]byte, error) {
	if val, ok := m.data[id]; ok {
		return val, nil
	}
	return nil, errors.New("Not Found")
}

func (m *Memory) Set(value []byte) (int, error) {
	// Atomically increment the counter to get a
	// consistent index snapshot
	index := atomic.AddInt64(&m.count, 1)
	// For tracking pending writes
	m.Add(1)
	go func(key int64, val []byte) {
		// Sleep(delay) as per the requirements
		time.Sleep(m.delay)
		// Lock during the write for memory safety
		// due to concurrency (just in case)
		m.Lock()
		m.data[int(key)] = val
		m.Unlock()
		m.Done()
	}(index, value)
	return int(index), nil
}

func (m *Memory) Close() error {
	m.Wait()
	return nil
}
