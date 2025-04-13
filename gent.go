package gent

import (
	"bufio"
	"fmt"
	"os"
)

// Pair is a pair of values.
type Pair[T any, U any] struct {
	First  T
	Second U
}

// NewPair create a Pair.
func NewPair[T any, U any](first T, second U) Pair[T, U] {
	return Pair[T, U]{First: first, Second: second}
}

// Set is a naive map backed set.
type Set[T comparable] struct {
	m map[T]bool
}

// NewSet creates a new Set.
func NewSet[T comparable](items ...T) *Set[T] {
	set := &Set[T]{m: map[T]bool{}}
	for _, each := range items {
		set.Add(each)
	}
	return set
}

// Add item to the set, return true if it was added. Otherwise it already existed.
func (v *Set[T]) Add(item T) (added bool) {
	_, existed := v.m[item]
	if existed {
		return
	}
	added = true
	v.m[item] = false
	return
}

// Clear the set, remove all items.
func (v *Set[T]) Clear() {
	v.m = map[T]bool{}
}

// Equal returns true when sets are equivalent in content.
func (v *Set[T]) Equal(s *Set[T]) bool {
	if len(v.m) != s.Len() {
		return false
	}
	for each := range v.m {
		if !s.Has(each) {
			return false
		}
	}
	return true
}

// Contains checks if item exists in the set.
func (v *Set[T]) Contains(item T) bool {
	return v.Has(item)
}

// Has checks if item exists in the set.
func (v *Set[T]) Has(item T) bool {
	_, ok := v.m[item]
	return ok
}

// ForEach iterates all items in the set, calls f for each item, stops if stop is called.
func (v *Set[T]) ForEach(f func(each T, stop func())) {
	breaker := false
	for each := range v.m {
		f(each, func() {
			breaker = true
		})
		if breaker {
			break
		}
	}
}

// Len returns the number of items in the set.
func (v *Set[T]) Len() int {
	return len(v.m)
}

// Count returns the number of items in the set.
func (v *Set[T]) Count() int {
	return len(v.m)
}

// Remove removes an item in the set, returns true if it was. I.e. if it existed.
func (v *Set[T]) Remove(item T) (existed bool) {
	_, existed = v.m[item]
	delete(v.m, item)
	return
}

// ToSlice returns a slice with all set items. Set itself doesn't change.
func (v *Set[T]) ToSlice() []T {
	keys := []T{}
	for each := range v.m {
		keys = append(keys, each)
	}
	return keys
}

// ReadLines read all lines in file filep.
func ReadLines(filep string) (lines []string, err error) {
	var f *os.File
	if f, err = os.Open(filep); err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return
}

// Tri returns one of the two values based on the condition.
func Tri[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// Map a slice into another slice of the same size.
func Map[T any, U any](s []T, f func(T) U) []U {
	mapped := make([]U, len(s))
	for i, v := range s {
		mapped[i] = f(v)
	}
	return mapped
}

// Filter values in s with f. When f returns true, item is included in the response slice.
func Filter[T any](s []T, f func(T) bool) []T {
	var filtered []T
	for _, v := range s {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// OrPanic2 return function that returns value if err is nil, else panics with message.
func OrPanic2[T any](value T, err error) func(message string) T {
	if err == nil {
		return func(_ string) T {
			return value
		}
	}
	return func(message string) T {
		panic(fmt.Sprintf("Message: %s. Error: %s.", message, err))
	}
}
