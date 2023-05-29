package tabulator_test

import (
	"github.com/clambin/go-common/tabulator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

func TestTabulator(t *testing.T) {
	d := tabulator.New("B")
	assert.NotNil(t, d)
	d.RegisterColumn("A")

	for day := 1; day < 5; day++ {
		assert.True(t, d.Add(time.Date(2022, time.January, day, 0, 0, 0, 0, time.UTC), "A", float64(day)))
		assert.True(t, d.Add(time.Date(2022, time.January, day, 0, 0, 0, 0, time.UTC), "B", -float64(day)))
	}

	assert.Equal(t, 4, d.Size())
	assert.Equal(t, []string{"A", "B"}, d.GetColumns())
	assert.Equal(t, []time.Time{
		time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.January, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.January, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.January, 4, 0, 0, 0, 0, time.UTC),
	}, d.GetTimestamps())

	values, ok := d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{1, 2, 3, 4}, values)

	values, ok = d.GetValues("B")
	require.True(t, ok)
	assert.Equal(t, []float64{-1, -2, -3, -4}, values)

	d.RegisterColumn("C")
	values, ok = d.GetValues("C")
	require.True(t, ok)
	assert.Equal(t, []float64{0, 0, 0, 0}, values)
}

func TestTabulator_Set(t *testing.T) {
	d := tabulator.New("A")

	for day := 1; day < 5; day++ {
		assert.True(t, d.Add(time.Date(2022, time.January, day, 0, 0, 0, 0, time.UTC), "A", float64(day)))
	}

	values, ok := d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{1, 2, 3, 4}, values)

	for day := 1; day < 5; day++ {
		assert.True(t, d.Set(time.Date(2022, time.January, day, 0, 0, 0, 0, time.UTC), "A", -1.0))
	}

	values, ok = d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{-1, -1, -1, -1}, values)
}

func TestTabulator_Add_OutOfOrder(t *testing.T) {
	d := tabulator.New("A")

	timestamp := time.Date(2022, time.November, 20, 0, 0, 0, 0, time.UTC)
	for day := 1; day < 5; day++ {
		assert.True(t, d.Add(timestamp, "A", float64(day)))
		timestamp = timestamp.Add(-24 * time.Hour)
	}

	assert.Equal(t, []time.Time{
		time.Date(2022, time.November, 17, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.November, 18, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.November, 19, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.November, 20, 0, 0, 0, 0, time.UTC),
	}, d.GetTimestamps())
	values, ok := d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{4, 3, 2, 1}, values)
}

func TestTabulator_Accumulate(t *testing.T) {
	d := tabulator.New("A")

	timestamp := time.Date(2022, time.November, 20, 0, 0, 0, 0, time.UTC)
	for day := 1; day < 5; day++ {
		assert.True(t, d.Add(timestamp, "A", float64(day)))
		timestamp = timestamp.Add(-24 * time.Hour)
	}
	d.Accumulate()

	assert.Equal(t, []time.Time{
		time.Date(2022, time.November, 17, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.November, 18, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.November, 19, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.November, 20, 0, 0, 0, 0, time.UTC),
	}, d.GetTimestamps())
	values, ok := d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{4, 7, 9, 10}, values)

}

func TestTabulator_Filter(t *testing.T) {
	d := tabulator.New("A")

	timestamp := time.Date(2022, time.November, 20, 0, 0, 0, 0, time.UTC)
	for day := 1; day < 5; day++ {
		assert.True(t, d.Add(timestamp, "A", float64(day)))
		timestamp = timestamp.Add(24 * time.Hour)
	}

	d.Filter(time.Time{}, time.Time{})
	assert.Len(t, d.GetTimestamps(), 4)

	d.Filter(time.Date(2022, time.November, 21, 0, 0, 0, 0, time.UTC), time.Time{})
	assert.Len(t, d.GetTimestamps(), 3)

	d.Filter(time.Time{}, time.Date(2022, time.November, 22, 0, 0, 0, 0, time.UTC))
	assert.Len(t, d.GetTimestamps(), 2)
	values, ok := d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{2, 3}, values)

	d.Filter(time.Date(2022, time.November, 21, 0, 0, 0, 0, time.UTC), time.Date(2022, time.November, 21, 0, 0, 0, 0, time.UTC))
	assert.Len(t, d.GetTimestamps(), 1)
	values, ok = d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{2}, values)

	d.Filter(time.Date(2022, time.November, 22, 0, 0, 0, 0, time.UTC), time.Date(2022, time.November, 22, 0, 0, 0, 0, time.UTC))
	assert.Len(t, d.GetTimestamps(), 0)
	values, ok = d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{}, values)
}

func TestTabulator_Copy(t *testing.T) {
	d := tabulator.New("A", "B")

	for day := 1; day < 5; day++ {
		assert.True(t, d.Add(time.Date(2022, time.January, day, 0, 0, 0, 0, time.UTC), "A", float64(day)))
		assert.True(t, d.Add(time.Date(2022, time.January, day, 0, 0, 0, 0, time.UTC), "B", -float64(day)))
	}

	d2 := d.Copy()
	assert.Equal(t, d.GetTimestamps(), d2.GetTimestamps())
	for _, col := range d.GetColumns() {
		values, _ := d.GetValues(col)
		values2, ok := d2.GetValues(col)
		require.True(t, ok)
		assert.Equal(t, values, values2)
	}
}

func BenchmarkTabulator_Add(b *testing.B) {
	timestamps := make([]time.Time, 365)
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 365; day++ {
		timestamps[day] = timestamp
		timestamp = timestamp.Add(24 * time.Hour)
	}

	d := tabulator.New("A")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for day := range timestamps {
			for iter := 0; iter < 100; iter++ {
				d.Add(timestamps[day], "A", float64(iter))
			}
		}
	}
}

func BenchmarkTabulator_Add_OutOfOrder(b *testing.B) {
	timestamps := make([]time.Time, 365)
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 365; day++ {
		timestamps[day] = timestamp
		timestamp = timestamp.Add(24 * time.Hour)
	}
	indexes := make([]int, 365*100)
	for i := range indexes {
		indexes[i] = int(rand.Int31n(365))
	}

	d := tabulator.New("A")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for iter := 0; iter < 100*365; iter++ {
			timestamp = timestamps[indexes[iter]]
			d.Add(timestamp, "A", float64(i))
		}
	}
}

func BenchmarkTabulator_Add_Same_Timestamp(b *testing.B) {
	tab := tabulator.New("A")
	for i := 0; i < b.N; i++ {
		timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
		for j := 0; j < 1000; j++ {
			tab.Add(timestamp, "A", float64(i))
		}
	}
}

func BenchmarkTabulator_Accumulate(b *testing.B) {
	d := tabulator.New("A")
	b.ResetTimer()
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 365; day++ {
		d.Add(timestamp, "A", float64(day))
		timestamp = timestamp.Add(24 * time.Hour)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Accumulate()
	}
}

func BenchmarkTabulator_Filter(b *testing.B) {
	d := tabulator.New("A")
	b.ResetTimer()
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 365; day++ {
		d.Add(timestamp, "A", float64(day))
		timestamp = timestamp.Add(24 * time.Hour)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Filter(
			time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
		)
	}
}
