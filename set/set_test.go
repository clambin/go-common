package set_test

import (
	"github.com/clambin/go-common/set"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
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
			assert.Equal(t, tt.expected, set.New(tt.input...).ListOrdered())
		})
	}
}

func TestSet_Add(t *testing.T) {
	tests := []struct {
		name     string
		start    set.Set[string]
		add      string
		expected set.Set[string]
	}{
		{name: "empty", start: set.New[string](), add: "A", expected: set.New("A")},
		{name: "add", start: set.New("A"), add: "B", expected: set.New("A", "B")},
		{name: "duplicate", start: set.New("A"), add: "A", expected: set.New("A")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.start.Add(tt.add)
			assert.Equal(t, tt.expected, tt.start)
		})
	}
}

func TestSet_Add2(t *testing.T) {
	tests := []struct {
		name     string
		start    set.Set[string]
		add      []string
		expected set.Set[string]
	}{
		{name: "empty", start: set.New[string](), add: []string{"A"}, expected: set.New("A")},
		{name: "add", start: set.New("A"), add: []string{"B"}, expected: set.New("A", "B")},
		{name: "duplicate", start: set.New("A"), add: []string{"A"}, expected: set.New("A")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.start.Add(tt.add...)
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
		{name: "empty", start: set.New[string](), remove: "A", expected: set.New[string]()},
		{name: "remove", start: set.New("A", "B"), remove: "B", expected: set.New("A")},
		{name: "remove last", start: set.New("A"), remove: "A", expected: set.New[string]()},
		{name: "non-existent", start: set.New("A"), remove: "B", expected: set.New("A")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.start.Remove(tt.remove)
			assert.Equal(t, tt.expected, tt.start)
		})
	}
}

func TestSet_Contains(t *testing.T) {
	tests := []struct {
		name     string
		input    set.Set[string]
		has      string
		expected bool
	}{
		{name: "empty", has: "A", expected: false},
		{name: "match", input: set.New("A", "B", "C"), has: "A", expected: true},
		{name: "mismatch", input: set.New("A", "B", "C"), has: "D", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.Contains(tt.has))
		})
	}
}

func TestSet_List(t *testing.T) {
	tests := []struct {
		name     string
		input    set.Set[string]
		expected []string
	}{
		{name: "empty", input: set.New[string](), expected: nil},
		{name: "single", input: set.New("A"), expected: []string{"A"}},
		{name: "multiple", input: set.New("B", "A"), expected: []string{"A", "B"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.ListOrdered())
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
		{name: "same", setA: set.New("A", "B", "C"), setB: set.New("C", "B", "A"), expected: true},
		{name: "subset", setA: set.New("A", "B"), setB: set.New("C", "B", "A"), expected: false},
		{name: "superset", setA: set.New("A", "B", "C", "D"), setB: set.New("C", "B", "A"), expected: false},
		{name: "different", setA: set.New("A", "B", "C", "D"), setB: set.New("X", "Y", "Z"), expected: false},
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
		{name: "empty", expected: set.New[string]()},
		{name: "not empty", input: set.New("A", "B", "C"), expected: set.New("C", "B", "A")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.Copy())
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
		{name: "empty", expected: set.New[string]()},
		{name: "first empty", setB: set.New("A", "B"), expected: set.New("A", "B")},
		{name: "second empty", setA: set.New("A", "B"), expected: set.New("A", "B")},
		{name: "union", setA: set.New("A", "B"), setB: set.New("B", "C"), expected: set.New("A", "B", "C")},
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
		{name: "empty", expected: set.New[string]()},
		{name: "first empty", setB: set.New("A", "B"), expected: set.New[string]()},
		{name: "second empty", setA: set.New("A", "B"), expected: set.New[string]()},
		{name: "intersection", setA: set.New("A", "B"), setB: set.New("B", "C"), expected: set.New("B")},
		{name: "no match", setA: set.New("A", "B"), setB: set.New("C", "D"), expected: set.New[string]()},
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
		{name: "empty", expected: set.New[string]()},
		{name: "first empty", setB: set.New("A", "B"), expected: set.New[string]()},
		{name: "second empty", setA: set.New("A", "B"), expected: set.New("A", "B")},
		{name: "overlap", setA: set.New("A", "B"), setB: set.New("B", "C"), expected: set.New("A")},
		{name: "full overlap", setA: set.New("A", "B"), setB: set.New("A", "B", "C"), expected: set.New[string]()},
		{name: "no overlap", setA: set.New("A", "B"), setB: set.New("C", "D"), expected: set.New("A", "B")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, set.Difference(tt.setA, tt.setB))
		})
	}
}
