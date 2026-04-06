package collections

type Collection[I comparable, T any] interface {
	Size() int

	IsEmpty() bool

	At(I) T

	Get(I) T

	Push(T)

	PushAll(...T)

	Join(Collection[I, T])

	Merge(Collection[I, T]) Collection[I, T]

	Delete(I)

	DeleteBy(Predicate[T])

	DeleteAll()

	Some(Predicate[T]) bool

	Find(Predicate[T]) (T, bool)

	Filter(Predicate[T]) Collection[I, T]

	ForEach(IndexedReceiver[I, T])
}

type Functor func()
type Receiver[T any] func(T) bool
type IndexedReceiver[I any, T any] func(I, T) bool
type Predicate[T any] func(T) bool
type Transformer[T, R any] func(*T) R

// Comparison result constants represent the standard return values for comparison operations.
const (
	// GREATER indicates that the left operand is greater than the right operand.
	GREATER = 1
	// EQUAL indicates that the left operand is equal to the right operand.
	EQUAL = 0
	// LOWER indicates that the left operand is less than the right operand.
	LOWER = -1
)

// Comparable is an interface for types that can be compared to other instances of the same type.
// Types implementing this interface provide a way to determine their ordering relative to other instances.
//
// Type parameters:
//   - T: The type being compared
type Comparable[T any] interface {
	// Compare compares this instance with another instance of the same type.
	// Returns:
	//   - GREATER (1) if this instance is greater than the argument
	//   - EQUAL (0) if this instance is equal to the argument
	//   - LOWER (-1) if this instance is less than the argument
	Compare(arg *T) int
}

// Comparator is a function type that defines a comparison operation between two values of the same type.
// It can be used to provide custom comparison logic for sorting, searching, and other operations
// that require establishing order between elements.
//
// Type parameters:
//   - T: The type being compared
//
// Returns:
//   - Positive value (typically 1) if arg0 > arg1
//   - Zero if arg0 == arg1
//   - Negative value (typically -1) if arg0 < arg1
type Comparator[T any] func(arg0, arg1 T) int
