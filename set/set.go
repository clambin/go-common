package set

import (
	"cmp"
	"maps"
	"slices"
)

// Set holds a set of unique values
type Set[T cmp.Ordered] map[T]struct{}

// New creates a new set containing the optional values
func New[T cmp.Ordered](values ...T) Set[T] {
	s := make(Set[T], len(values))
	for _, value := range values {
		s[value] = struct{}{}
	}
	return s
}

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
	values := make([]T, 0, len(s))
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

// Clone returns a copy of the set
func (s Set[T]) Clone() Set[T] {
	if s == nil {
		return New[T]()
	}
	return maps.Clone(s)
}

// Union returns a new set containing all values from setA and setB
func Union[T cmp.Ordered](setA, setB Set[T]) Set[T] {
	union := setA.Clone()
	for key := range setB {
		union[key] = struct{}{}
	}
	return union
}

// Intersection returns a new set containing the common values between setA and setB
func Intersection[T cmp.Ordered](setA, setB Set[T]) Set[T] {
	intersection := make(Set[T])
	for key := range setA {
		if _, ok := setB[key]; ok {
			intersection[key] = struct{}{}
		}
	}
	return intersection
}

// Difference returns a new set containing the values from setA that don't exist in setB
func Difference[T cmp.Ordered](setA, setB Set[T]) Set[T] {
	difference := make(Set[T])
	for key := range setA {
		if _, ok := setB[key]; !ok {
			difference[key] = struct{}{}
		}
	}
	return difference
}
