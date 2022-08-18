package app

import (
	"crypto/sha512"
	"encoding/base64"

	"github.com/phrozen/password-hash-exercise/internal/store"
)

// 512 bits / 8 = 64 bytes * 8 / 6 = 86
// plus padding for a multiple of 4 = 88
const HASH_LENGTH = 88

// App (application) implements the core business logic based on requirements
// which is to save hashed passwords and retrieve them later.
type App struct {
	store store.Store
}

// New creates an application with the given Store implementation
func New(s store.Store) *App {
	return &App{store: s}
}

// GetHash returns the hash at the given id from the Store
func (app *App) GetHash(id int) (string, error) {
	hash, err := app.store.Get(id)
	return string(hash), err
}

// SetHash receives a password to be hashed with SHA512 algorithm and
// then converted to base64 encoding and saved to the Store, returns
// the id where the hash is/will be saved.
func (app *App) SetHash(password string) (int, error) {
	hash := app.hash([]byte(password))
	return app.store.Set(hash)
}

// Close runs all tear down operations like closing the Store
func (app *App) Close() error {
	return app.store.Close()
}

// hashes any input with SHA512 and encodes to standard base64
// returned []byte has ALWAYS 88 bytes length and should never fail
func (app *App) hash(input []byte) []byte {
	hash := sha512.Sum512(input)
	// Deterministic buffer size will improve
	// performance, if slow, try sync.Pool
	output := make([]byte, HASH_LENGTH)
	// convert fixed size array 'hash' by slicing it
	base64.StdEncoding.Encode(output, hash[:])
	return output
}
