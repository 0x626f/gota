package graph

// removeFromSlice removes the first occurrence of v from s using swap-with-last
// (O(n) search, O(1) removal, does not preserve order).
func removeFromSlice[T comparable](s []T, v T) []T {
	for i, k := range s {
		if k == v {
			last := len(s) - 1
			s[i] = s[last]
			var zero T
			s[last] = zero
			return s[:last]
		}
	}
	return s
}

// growTo ensures s has at least index+1 elements, appending zero values as
// needed. Returns the (possibly new) slice.
func growTo[T any](s []T, index int) []T {
	if index >= len(s) {
		return append(s, make([]T, index-len(s)+1)...)
	}
	return s
}

// insertSorted inserts v into a sorted []int slice, maintaining order.
// No-op if v is already present.
func insertSorted(s *[]int, v int) {
	slice := *s
	lo, hi := 0, len(slice)
	for lo < hi {
		mid := (lo + hi) >> 1
		if slice[mid] < v {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < len(slice) && slice[lo] == v {
		return
	}
	*s = append(slice, 0)
	copy((*s)[lo+1:], (*s)[lo:])
	(*s)[lo] = v
}

// removeSorted removes v from a sorted []int slice if present.
func removeSorted(slice []int, v int) []int {
	lo, hi := 0, len(slice)
	for lo < hi {
		mid := (lo + hi) >> 1
		if slice[mid] < v {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo >= len(slice) || slice[lo] != v {
		return slice
	}
	return append(slice[:lo], slice[lo+1:]...)
}

// searchSorted reports whether v is present in a sorted []int slice.
func searchSorted(slice []int, v int) bool {
	lo, hi := 0, len(slice)
	for lo < hi {
		mid := (lo + hi) >> 1
		if slice[mid] == v {
			return true
		} else if slice[mid] < v {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return false
}
