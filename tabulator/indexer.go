package tabulator

import (
	"sort"
	"time"
)

type ordered interface {
	time.Time | string
}

// indexer holds a unique set of values, and records the order in which they were added.
// Currently, it supports string and time.Time data.
type indexer[T ordered] struct {
	values  []T
	indices map[T]int
	inOrder bool
}

// makeIndexer returns a new indexer
func makeIndexer[T ordered]() indexer[T] {
	return indexer[T]{
		values:  make([]T, 0),
		indices: make(map[T]int),
		inOrder: true,
	}
}

// GetIndex returns the index of a value (i.e. when that value was added)
func (idx *indexer[T]) GetIndex(value T) (index int, found bool) {
	index, found = idx.indices[value]
	return
}

// Count returns the number of values in the indexer
func (idx *indexer[T]) Count() int {
	return len(idx.values)
}

// List returns the (sorted) values in the indexer
func (idx *indexer[T]) List() []T {
	if !idx.inOrder {
		sort.Slice(idx.values, func(i, j int) bool { return isLessThan(idx.values[i], idx.values[j]) })
		idx.inOrder = true
	}
	return idx.values
}

// Add adds a new value to the indexer. It returns the index of that value and whether the value was actually added.
func (idx *indexer[T]) Add(value T) (int, bool) {
	index, found := idx.indices[value]

	if found {
		return index, false
	}

	index = len(idx.values)
	idx.indices[value] = index

	if idx.inOrder && index > 0 {
		idx.inOrder = !isLessThan(value, idx.values[index-1])
	}
	idx.values = append(idx.values, value)
	return index, true
}

func isLessThan[T ordered](a, b T) bool {
	// this works around the fact that we can't type switch on T
	var x interface{} = a
	var y interface{} = b
	switch (x).(type) {
	case string:
		return x.(string) < y.(string)
	case time.Time:
		return x.(time.Time).Before(y.(time.Time))
	}
	panic("unsupported type")
}

func (idx *indexer[T]) Copy() indexer[T] {
	values := make([]T, len(idx.values))
	copy(values, idx.values)

	indices := make(map[T]int)
	for key, value := range idx.indices {
		indices[key] = value
	}

	return indexer[T]{
		values:  values,
		indices: indices,
		inOrder: idx.inOrder,
	}
}
