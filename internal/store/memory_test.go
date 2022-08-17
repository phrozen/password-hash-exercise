package store

import (
	"bytes"
	"fmt"
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

// Tests store for correct set/get ops
func TestSetGet(t *testing.T) {
	store := NewMemory(0)
	defer store.Close()
	for i := 1; i <= 100; i++ {
		input := []byte(fmt.Sprintf("%d", i))
		index, err := store.Set(input)
		equal(t, err, nil)
		equal(t, index, i)
		// Even with 0 delay, scheduler needs time for
		// context switching, 1ms should be enough
		// but we use higher values for slow machines
		time.Sleep(25 * time.Millisecond)
		output, err := store.Get(i)
		equal(t, err, nil)
		equal(t, 0, bytes.Compare(input, output))
	}
}

func TestDelay(t *testing.T) {
	// Start with 100ms delay writes
	store := NewMemory(100 * time.Millisecond)
	defer store.Close()
	input := []byte("test")
	index, err := store.Set(input)
	equal(t, err, nil)
	equal(t, index, 1)
	// Expect index not to be there yet
	output, err := store.Get(index)
	equal(t, true, err != nil)
	equal(t, 0, bytes.Compare([]byte(nil), output))
	// Wait more than delay...if it fails...
	// try larger value on slow machines
	time.Sleep(125 * time.Millisecond)
	// Expect index to be there
	output, err = store.Get(index)
	equal(t, nil, err)
	equal(t, 0, bytes.Compare(input, output))
}
