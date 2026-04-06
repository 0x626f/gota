package array

import "github.com/0x626f/gota/collections"

// IsSorted checks whether the array is sorted in either ascending or descending order
// according to the provided comparator.
// Returns true if the array is sorted, false otherwise.
func (array *Base[I, T]) IsSorted(comparator collections.Comparator[T]) bool {
	if array.Size() < 2 {
		return true
	}

	order := comparator(array.At(0), array.At(1))

	for index := 1; index < array.Size(); index++ {
		direction := comparator(array.At(index-1), array.At(index))
		if direction != order && direction != collections.EQUAL {
			return false
		}
	}
	return true
}

// InsertionSort sorts the array in place using insertion sort algorithm.
// This algorithm has O(n²) time complexity but works well for small or nearly sorted arrays.
func (array *Base[I, T]) InsertionSort(comparator collections.Comparator[T]) {
	for i := 1; i < array.Size(); i++ {
		for j := i; j != 0 && comparator(array.At(j), array.At(j-1)) == collections.LOWER; j-- {
			array.Swap(j, j-1)
		}
	}
}

// HeapSort sorts the array in place using heap sort algorithm.
// This algorithm has O(n log n) time complexity and O(1) space complexity.
func (array *Base[I, T]) HeapSort(comparator collections.Comparator[T]) {
	shiftDown := func(left, right int) {
		root := left
		for {
			child := 2*root + 1
			if child >= right {
				break
			}
			if child+1 < right && comparator(array.At(child), array.At(child+1)) == collections.LOWER {
				child++
			}
			if comparator(array.At(root), array.At(child)) != collections.LOWER {
				return
			}
			array.Swap(root, child)
			root = child
		}
	}

	size := array.Size()

	for i := (size - 1) / 2; i >= 0; i-- {
		shiftDown(i, size)
	}

	for i := size - 1; i >= 0; i-- {
		array.Swap(0, i)
		shiftDown(0, i)
	}
}

// BinarySearch finds an element equal to the target according to the provided comparator.
// This algorithm has O(log n) time complexity and requires the array to be sorted.
// Returns the found element and true if found, otherwise returns collections.Zero[T]() value and false.
func (array *Base[I, T]) BinarySearch(target T, comparator collections.Comparator[T]) (T, bool) {
	n := array.Size()
	i, j := 0, n
	for i < j {
		h := int(uint(i+j) >> 1)
		if comparator(array.At(h), target) == collections.LOWER {
			i = h + 1
		} else {
			j = h
		}
	}

	if i < n && comparator(array.At(i), target) == collections.EQUAL {
		return array.At(i), true
	}

	return collections.Zero[T](), false
}

// Min finds the minimum element in the array according to the provided comparator.
// Returns the minimum element and true if the array is not empty,
// otherwise returns collections.Zero[T]() value and false.
func (array *Base[I, T]) Min(comparator collections.Comparator[T]) (T, bool) {
	if array.Size() == 0 {
		return collections.Zero[T](), false
	}

	result := array.At(0)

	for index := range array.Size() {
		item := array.At(index)
		if comparator(item, result) == collections.LOWER {
			result = item
		}

	}

	return result, true
}

// Max finds the maximum element in the array according to the provided comparator.
// Returns the maximum element and true if the array is not empty,
// otherwise returns collections.Zero[T]() value and false.
func (array *Base[I, T]) Max(comparator collections.Comparator[T]) (T, bool) {
	if array.Size() == 0 {
		return collections.Zero[T](), false
	}

	result := array.At(0)

	for index := range array.Size() {
		item := array.At(index)
		if comparator(item, result) == collections.GREATER {
			result = item
		}

	}

	return result, true
}
