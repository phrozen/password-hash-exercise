package app

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/phrozen/password-hash-exercise/internal/store"
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

// Test with random input sizes and data, assert the resulting
// hash is always the expected length, regardless of input.
func TestHash(t *testing.T) {
	app := &App{} // avoid Shutdown
	for i := 0; i < 100; i++ {
		password := make([]byte, rand.Intn(64)+1)
		rand.Read(password)
		hash := app.hash(password)
		equal(t, HASH_LENGTH, len(hash))
	}
	// Test for zero and nil values
	var pass []byte
	hash := app.hash(pass)
	equal(t, HASH_LENGTH, len(hash))
	hash = app.hash(nil)
	equal(t, HASH_LENGTH, len(hash))
}

// Kinda redundant with just a single Store implementation,
// just in case more Stores are added for comparison
func TestSetGet(t *testing.T) {
	app := New(store.NewMemory(0))
	defer app.Close()
	for i := 1; i <= 100; i++ {
		input := fmt.Sprintf("password-%d", i)
		index, err := app.SetHash(input)
		equal(t, err, nil)
		equal(t, index, i)
		// Even with 0 delay, scheduler needs time for
		// context switching, 1ms should be enough
		// but we use higher values for slow machines
		time.Sleep(10 * time.Millisecond)
		output, err := app.GetHash(i)
		equal(t, err, nil)
		equal(t, HASH_LENGTH, len(output))
	}
}

// Shows consistent performance up to 64 byte inputs (common use case)
// Implementation has consistent allocs and memory usage (deterministic)
func BenchmarkHash(b *testing.B) {
	app := &App{} // avoid Shutdown
	for s := 8; s <= 1024; s = s << 1 {
		password := make([]byte, s)
		rand.Read(password)
		b.Run(fmt.Sprintf("%d_bytes_input", s), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				app.hash(password)
			}
		})
	}
}
