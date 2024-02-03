package set

import (
	"cmp"
	"slices"
)

// Set holds a set of unique values
type Set[T cmp.Ordered] map[T]struct{}

// Add adds value to the set
func (s Set[T]) Add(value ...T) {
	for i := range value {
		s[value[i]] = struct{}{}
	}
}

// Remove deletes value from the set, if present
func (s Set[T]) Remove(value ...T) {
	for i := range value {
		delete(s, value[i])
	}
}

// Contains returns true if the set contains value
func (s Set[T]) Contains(value T) bool {
	_, found := s[value]
	return found
}

// List returns all values present in the set. Order is not guaranteed.  If ListOrdered() if required.
func (s Set[T]) List() []T {
	var values []T
	for k := range s {
		values = append(values, k)
	}
	return values
}

// ListOrdered returns all values present in the set, in order
func (s Set[T]) ListOrdered() []T {
	values := s.List()
	slices.Sort(values)
	return values
}

// Equals returns true if both sets contain the same values
func (s Set[T]) Equals(other Set[T]) bool {
	for key := range s {
		if _, ok := other[key]; !ok {
			return false
		}
	}
	for key := range other {
		if _, ok := s[key]; !ok {
			return false
		}
	}
	return true
}

// Copy returns a copy of the set
func (s Set[T]) Copy() Set[T] {
	output := make(map[T]struct{})
	for key := range s {
		output[key] = struct{}{}
	}
	return output
}

// New creates a new set containing the optional values
func New[T cmp.Ordered](values ...T) Set[T] {
	return Create(values...)
}

// Create creates a new set containing the optional values
// Deprecated: use New() instead.
func Create[T cmp.Ordered](values ...T) Set[T] {
	s := make(Set[T])
	for _, value := range values {
		s[value] = struct{}{}
	}
	return s
}

// Union returns a new set containing the values from setA and setB
func Union[T cmp.Ordered](setA, setB Set[T]) Set[T] {
	union := setA.Copy()
	for key := range setB {
		union[key] = struct{}{}
	}
	return union
}

// Intersection returns a new set containing the values that exist in both setA and setB
func Intersection[T cmp.Ordered](setA, setB Set[T]) Set[T] {
	intersection := make(Set[T])
	for key := range setA {
		if setB.Contains(key) {
			intersection[key] = struct{}{}
		}
	}
	return intersection
}

// Difference returns a new set holding the values from setA that don't exist in setB
func Difference[T cmp.Ordered](setA, setB Set[T]) Set[T] {
	difference := make(Set[T])
	for key := range setA {
		if _, ok := setB[key]; !ok {
			difference[key] = struct{}{}
		}
	}
	return difference
}
