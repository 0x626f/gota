package cache

// Pair stores two related values.
type Pair[F, S any] struct {
	First  F
	Second S
}
