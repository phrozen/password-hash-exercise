package app

import (
	"crypto/sha512"
	"encoding/base64"

	"github.com/phrozen/password-hash-exercise/internal/store"
)

// 512 bits / 8 = 64 bytes * 8 / 6 = 86
// plus padding for a multiple of 4 = 88
const HASH_LENGTH = 88

type App struct {
	store store.Store
}

func New(s store.Store) *App {
	return &App{store: s}
}

func (app *App) GetHash(id int) (string, error) {
	hash, err := app.store.Get(id)
	return string(hash), err
}

func (app *App) SetHash(password string) (int, error) {
	hash := app.hash([]byte(password))
	return app.store.Set(hash)
}

// Close runs all tear down operations like closing the Store
func (app *App) Close() error {
	return app.store.Close()
}

func (app *App) hash(input []byte) []byte {
	hash := sha512.Sum512(input)
	// Deterministic buffer size will improve
	// performance, if slow, try sync.Pool
	output := make([]byte, HASH_LENGTH)
	// convert fixed size array 'hash' by slicing it
	base64.StdEncoding.Encode(output, hash[:])
	return output
}
