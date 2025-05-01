// Package gent contains miscellaneous tools that the author keeps rewriting in various projects.
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

// NewPair create an initialized [gent.Pair].
func NewPair[T any, U any](first T, second U) Pair[T, U] {
	return Pair[T, U]{First: first, Second: second}
}

// Set is a naive map backed set.
type Set[T comparable] struct {
	m map[T]bool
}

// NewSet creates a new [gent.Set].
func NewSet[T comparable](items ...T) *Set[T] {
	set := &Set[T]{m: map[T]bool{}}
	for _, each := range items {
		set.Add(each)
	}
	return set
}

// Add item to the set, return true if it was added.
// Otherwise it already existed and wasn't added.
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

// Equal returns true when the sets contain the exact same items.
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
// Alias for [gent.Set.Has].
func (v *Set[T]) Contains(item T) bool {
	return v.Has(item)
}

// Has checks if item exists in the set.
func (v *Set[T]) Has(item T) bool {
	_, ok := v.m[item]
	return ok
}

// ForEach iterates all items in the set, calls f for each item, stops if stop is called.
// Use [gent.ForEachAll] if there's no need to stop iteration.
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

// ForEachAll iterates all items in the set and calls f for each item.
// Use [gent.ForEach] if you need to stop iteration.
func (v *Set[T]) ForEachAll(f func(each T)) {
	for key := range v.m {
		f(key)
	}
}

// Len returns the number of items in the set.
func (v *Set[T]) Len() int {
	return len(v.m)
}

// Count returns the number of items in the set.
// Alias for [gent.Set.Len].
func (v *Set[T]) Count() int {
	return v.Len()
}

// Remove removes an item in the set, returns true if it was.
// I.e. if it existed.
func (v *Set[T]) Remove(item T) (existed bool) {
	_, existed = v.m[item]
	delete(v.m, item)
	return
}

// ToSlice returns a slice with all set items.
// Set itself doesn't change.
func (v *Set[T]) ToSlice() []T {
	keys := []T{}
	for each := range v.m {
		keys = append(keys, each)
	}
	return keys
}

// ReadLines read all lines in file filep.
// Empty lines are included.
// Returned lines do not contain newlines at the end.
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
// I.e. this is a ternary "operator".
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

// Filter values in s with f.
// When f returns true, item is included in the response slice.
func Filter[T any](s []T, f func(T) bool) []T {
	var filtered []T
	for _, v := range s {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// OrPanic2 returns function that returns value if err is nil, else panics with message.
// Useful for cases where failure should result in panic
// and you don't want to deal with the returned error.
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

// NewOption is a general function to implement option pattern.
func NewOption[T any](t T, options ...func(t *T)) T {
	for _, each := range options {
		each(&t)
	}
	return t
}
