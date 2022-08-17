package store

// Store defines an interface for a store of any byte slice that
// tracks the elements with an integer id in incremental fashion.
// Close is added as it is a common practice for other non-trivial
// implementations to perform tear down processes like graceful shutdown.
type Store interface {
	Get(int) ([]byte, error)
	Set([]byte) (int, error)
	Close() error
}
