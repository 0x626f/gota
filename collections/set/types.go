package set

// Keyable is an interface for types that can provide a comparable key.
// This is useful for maps, dictionaries, and other data structures that require
// unique identification of elements.
//
// Type parameters:
//   - T: The key type, which must be comparable (support == and != operators)
type Keyable[T comparable] interface {
	// Key returns a comparable value that uniquely identifies this object.
	// This value can be used as a key in maps or for equality comparison.
	Key() T
}

// KeyableWrapper is a utility struct that wraps any comparable value
// to make it implement the Keyable interface. This allows simple values
// to be used in contexts requiring the Keyable interface.
//
// Type parameters:
//   - T: The wrapped value type, which must be comparable
type KeyableWrapper[T comparable] struct {
	// Wrapped is the underlying value being wrapped
	Wrapped T
}

// Key returns the wrapped value as the key.
// This implements the Keyable interface.
func (keyableWrapper *KeyableWrapper[T]) Key() T {
	return keyableWrapper.Wrapped
}

// Value returns the wrapped value.
// This provides direct access to the underlying value.
func (keyableWrapper *KeyableWrapper[T]) Value() T {
	return keyableWrapper.Wrapped
}

func WrapPrimitive[T comparable](item T) *KeyableWrapper[T] {
	return &KeyableWrapper[T]{item}
}
