package tabulator_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/clambin/go-common/tabulator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTabulator_MarshalJSON(t *testing.T) {
	table := tabulator.New("A", "B", "C", "D")

	table.Add(time.Date(2023, time.July, 30, 0, 0, 0, 0, time.UTC), "D", 4.0)
	table.Add(time.Date(2023, time.July, 29, 0, 0, 0, 0, time.UTC), "C", 3.0)
	table.Add(time.Date(2023, time.July, 28, 0, 0, 0, 0, time.UTC), "B", 2.0)
	table.Add(time.Date(2023, time.July, 27, 0, 0, 0, 0, time.UTC), "A", 1.0)

	body, err := json.Marshal(table)
	require.NoError(t, err)
	assert.Equal(t, `{"Timestamps":["2023-07-27T00:00:00Z","2023-07-28T00:00:00Z","2023-07-29T00:00:00Z","2023-07-30T00:00:00Z"],"Columns":["A","B","C","D"],"Data":[[1,0,0,0],[0,2,0,0],[0,0,3,0],[0,0,0,4]]}`, string(body))
}

func BenchmarkTabulator_MarshalJSON(b *testing.B) {
	table := makeBigTable()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(table); err != nil {
			b.Fatal(err)
		}
	}
}

func TestTabulator_UnmarshalJSON(t *testing.T) {
	body := `{"Timestamps":["2023-07-27T00:00:00Z","2023-07-28T00:00:00Z","2023-07-29T00:00:00Z","2023-07-30T00:00:00Z"],"Columns":["A","B","C","D"],"Data":[[1,0,0,0],[0,2,0,0],[0,0,3,0],[0,0,0,4]]}`

	var table tabulator.Tabulator
	err := json.Unmarshal([]byte(body), &table)
	require.NoError(t, err)

	assert.Len(t, table.GetTimestamps(), 4)
	assert.Equal(t, []string{"A", "B", "C", "D"}, table.GetColumns())

	for _, column := range []struct {
		name   string
		values []float64
	}{
		{name: "A", values: []float64{1, 0, 0, 0}},
		{name: "B", values: []float64{0, 2, 0, 0}},
		{name: "C", values: []float64{0, 0, 3, 0}},
		{name: "D", values: []float64{0, 0, 0, 4}},
	} {
		values, ok := table.GetValues(column.name)
		require.True(t, ok)
		assert.Equal(t, column.values, values, column.name)
	}
}

func BenchmarkTabulator_UnmarshalJSON(b *testing.B) {
	body, err := json.Marshal(makeBigTable())
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var table *tabulator.Tabulator
		if err = json.Unmarshal(body, &table); err != nil {
			b.Fatal(err)
		}
	}
}

func TestTabulator_MarshalBinary(t *testing.T) {
	table := tabulator.New("A", "B", "C", "D")

	table.Add(time.Date(2023, time.July, 30, 0, 0, 0, 0, time.UTC), "D", 4.0)
	table.Add(time.Date(2023, time.July, 29, 0, 0, 0, 0, time.UTC), "C", 3.0)
	table.Add(time.Date(2023, time.July, 28, 0, 0, 0, 0, time.UTC), "B", 2.0)
	table.Add(time.Date(2023, time.July, 27, 0, 0, 0, 0, time.UTC), "A", 1.0)

	var binary bytes.Buffer
	err := gob.NewEncoder(&binary).Encode(table)
	require.NoError(t, err)

	var table2 tabulator.Tabulator
	err = gob.NewDecoder(&binary).Decode(&table2)
	require.NoError(t, err)

	assert.Len(t, table2.GetTimestamps(), 4)
	require.Equal(t, table.GetColumns(), table2.GetColumns())

	for _, col := range table2.GetColumns() {
		values, ok := table.GetValues(col)
		require.True(t, ok)
		values2, ok := table2.GetValues(col)
		require.True(t, ok)
		assert.Equal(t, values, values2)
	}
}

func BenchmarkTabulator_MarshalBinary(b *testing.B) {
	table := makeBigTable()
	var binary bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Reset()
		if err := gob.NewEncoder(&binary).Encode(table); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTabulator_UnmarshalBinary(b *testing.B) {
	var binary bytes.Buffer
	require.NoError(b, gob.NewEncoder(&binary).Encode(makeBigTable()))
	body := binary.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var output tabulator.Tabulator
		if err := gob.NewDecoder(bytes.NewBuffer(body)).Decode(&output); err != nil {
			b.Fatal(err)
		}
	}
}
