package set_test

import (
	"github.com/clambin/go-common/set"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestSet_Add(t *testing.T) {
	tests := []struct {
		name     string
		start    set.Set[string]
		add      string
		expected set.Set[string]
	}{
		{name: "empty", start: set.Create[string](), add: "A", expected: set.Create[string]("A")},
		{name: "add", start: set.Create[string]("A"), add: "B", expected: set.Create[string]("A", "B")},
		{name: "duplicate", start: set.Create[string]("A"), add: "A", expected: set.Create[string]("A")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.start.Add(tt.add)
			assert.Equal(t, tt.expected, tt.start)
		})
	}
}

func TestSet_Remove(t *testing.T) {
	tests := []struct {
		name     string
		start    set.Set[string]
		remove   string
		expected set.Set[string]
	}{
		{name: "empty", start: set.Create[string](), remove: "A", expected: set.Create[string]()},
		{name: "remove", start: set.Create[string]("A", "B"), remove: "B", expected: set.Create[string]("A")},
		{name: "remove last", start: set.Create[string]("A"), remove: "A", expected: set.Create[string]()},
		{name: "non-existent", start: set.Create[string]("A"), remove: "B", expected: set.Create[string]("A")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.start.Remove(tt.remove)
			assert.Equal(t, tt.expected, tt.start)
		})
	}
}

func TestSet_Has(t *testing.T) {
	tests := []struct {
		name     string
		input    set.Set[string]
		has      string
		expected bool
	}{
		{name: "empty", has: "A", expected: false},
		{name: "match", input: set.Create("A", "B", "C"), has: "A", expected: true},
		{name: "mismatch", input: set.Create("A", "B", "C"), has: "D", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.Has(tt.has))
		})
	}
}

func TestSet_List(t *testing.T) {
	tests := []struct {
		name     string
		input    set.Set[string]
		expected []string
	}{
		{name: "empty", input: set.Create[string](), expected: nil},
		{name: "single", input: set.Create[string]("A"), expected: []string{"A"}},
		{name: "multiple", input: set.Create[string]("B", "A"), expected: []string{"A", "B"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.input.List()
			sort.Strings(l)
			assert.Equal(t, tt.expected, l)
		})
	}
}

func TestSet_Equals(t *testing.T) {
	tests := []struct {
		name     string
		setA     set.Set[string]
		setB     set.Set[string]
		expected bool
	}{
		{name: "empty", expected: true},
		{name: "same", setA: set.Create("A", "B", "C"), setB: set.Create("C", "B", "A"), expected: true},
		{name: "subset", setA: set.Create("A", "B"), setB: set.Create("C", "B", "A"), expected: false},
		{name: "superset", setA: set.Create("A", "B", "C", "D"), setB: set.Create("C", "B", "A"), expected: false},
		{name: "different", setA: set.Create("A", "B", "C", "D"), setB: set.Create("X", "Y", "Z"), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.setA.Equals(tt.setB))
		})
	}
}

func TestSet_Copy(t *testing.T) {
	tests := []struct {
		name     string
		input    set.Set[string]
		expected set.Set[string]
	}{
		{name: "empty", expected: set.Create[string]()},
		{name: "not empty", input: set.Create("A", "B", "C"), expected: set.Create("C", "B", "A")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.Copy())
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{name: "empty"},
		{name: "not empty", input: []string{"A", "B"}, expected: []string{"A", "B"}},
		{name: "duplicates", input: []string{"A", "B", "A"}, expected: []string{"A", "B"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := set.Create(tt.input...).List()
			sort.Strings(output)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestUnion(t *testing.T) {
	tests := []struct {
		name     string
		setA     set.Set[string]
		setB     set.Set[string]
		expected set.Set[string]
	}{
		{name: "empty", expected: set.Create[string]()},
		{name: "first empty", setB: set.Create[string]("A", "B"), expected: set.Create[string]("A", "B")},
		{name: "second empty", setA: set.Create[string]("A", "B"), expected: set.Create[string]("A", "B")},
		{name: "union", setA: set.Create[string]("A", "B"), setB: set.Create[string]("B", "C"), expected: set.Create[string]("A", "B", "C")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, set.Union(tt.setA, tt.setB))
		})
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		name     string
		setA     set.Set[string]
		setB     set.Set[string]
		expected set.Set[string]
	}{
		{name: "empty", expected: set.Create[string]()},
		{name: "first empty", setB: set.Create[string]("A", "B"), expected: set.Create[string]()},
		{name: "second empty", setA: set.Create[string]("A", "B"), expected: set.Create[string]()},
		{name: "intersection", setA: set.Create[string]("A", "B"), setB: set.Create[string]("B", "C"), expected: set.Create[string]("B")},
		{name: "no match", setA: set.Create[string]("A", "B"), setB: set.Create[string]("C", "D"), expected: set.Create[string]()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, set.Intersection(tt.setA, tt.setB))
		})
	}
}

func TestDifference(t *testing.T) {
	tests := []struct {
		name     string
		setA     set.Set[string]
		setB     set.Set[string]
		expected set.Set[string]
	}{
		{name: "empty", expected: set.Create[string]()},
		{name: "first empty", setB: set.Create[string]("A", "B"), expected: set.Create[string]()},
		{name: "second empty", setA: set.Create[string]("A", "B"), expected: set.Create[string]("A", "B")},
		{name: "overlap", setA: set.Create[string]("A", "B"), setB: set.Create[string]("B", "C"), expected: set.Create[string]("A")},
		{name: "full overlap", setA: set.Create[string]("A", "B"), setB: set.Create[string]("A", "B", "C"), expected: set.Create[string]()},
		{name: "no overlap", setA: set.Create[string]("A", "B"), setB: set.Create[string]("C", "D"), expected: set.Create[string]("A", "B")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, set.Difference(tt.setA, tt.setB))
		})
	}
}
