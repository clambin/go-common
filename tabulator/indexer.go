package tabulator

import (
	"cmp"
	"slices"
	"time"
)

type ordered interface {
	time.Time | string
}

// indexer holds a unique set of values, and records the order in which they were added.
// Currently, it supports string and time.Time Data.
type indexer[T ordered] struct {
	Values  []T
	Indices map[T]int
	InOrder bool
}

// makeIndexer returns a new indexer
func makeIndexer[T ordered]() indexer[T] {
	return indexer[T]{
		//values:  make([]T, 0),
		Indices: make(map[T]int),
		InOrder: true,
	}
}

// makeIndexerFromData returns a new indexer, initialized with the provided Data.
func makeIndexerFromData[T ordered](data []T) indexer[T] {
	index := indexer[T]{
		Values:  data,
		Indices: make(map[T]int),
		InOrder: true,
	}

	var previous T
	for row, entry := range data {
		if compare(previous, entry) <= 0 {
			index.InOrder = false
		}
		index.Indices[entry] = row
	}
	return index
}

// GetIndex returns the index of a value (i.e. when that value was added)
func (idx *indexer[T]) GetIndex(value T) (index int, found bool) {
	index, found = idx.Indices[value]
	return
}

// Count returns the number of values in the indexer
func (idx *indexer[T]) Count() int {
	return len(idx.Values)
}

// List returns the (sorted) values in the indexer
func (idx *indexer[T]) List() []T {
	if !idx.InOrder {
		slices.SortFunc(idx.Values, compare[T])
		idx.InOrder = true
	}
	return idx.Values
}

// Add adds a new value to the indexer. It returns the index of that value and whether the value was actually added.
func (idx *indexer[T]) Add(value T) (int, bool) {
	index, found := idx.Indices[value]

	if found {
		return index, false
	}

	index = len(idx.Values)
	idx.Indices[value] = index

	if idx.InOrder && index > 0 {
		idx.InOrder = compare(value, idx.Values[index-1]) >= 0
	}
	idx.Values = append(idx.Values, value)
	return index, true
}

func compare[T ordered](a, b T) int {
	// this works around the fact that we can't type switch on T
	var x interface{} = a
	var y interface{} = b
	switch (x).(type) {
	case string:
		return cmp.Compare(x.(string), y.(string)) //x.(string) < y.(string)
	case time.Time:
		if x.(time.Time).Equal(y.(time.Time)) {
			return 0
		}
		if x.(time.Time).Before(y.(time.Time)) {
			return -1
		}
		return 1
	}
	panic("unsupported type")
}

func (idx *indexer[T]) Copy() indexer[T] {
	values := make([]T, len(idx.Values))
	copy(values, idx.Values)

	indices := make(map[T]int)
	for key, value := range idx.Indices {
		indices[key] = value
	}

	return indexer[T]{
		Values:  values,
		Indices: indices,
		InOrder: idx.InOrder,
	}
}
