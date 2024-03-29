package tabulator

import (
	"time"
)

// Tabulator tabulates a set of entries in rows by timestamp and columns by label.  For performance reasons, Data must
// be added sequentially.
type Tabulator struct {
	Timestamps indexer[time.Time]
	Columns    indexer[string]
	Data       [][]float64

	lastTimestamp time.Time
	lastRow       int
}

// New creates a new Tabulator
func New(columns ...string) *Tabulator {
	t := &Tabulator{
		Timestamps: makeIndexer[time.Time](),
		Columns:    makeIndexer[string](),
	}
	t.RegisterColumn(columns...)
	return t
}

// Add adds a value for a specified timestamp and column to the table.  If there is already a value for that
// timestamp and column, the specified value is added to the existing value.
//
// Returns false if the column does not exist. Use RegisterColumn to add it first.
func (t *Tabulator) Add(timestamp time.Time, column string, value float64) bool {
	return t.addOrSet(timestamp, column, value, true)
}

// Set sets the value for a specified timestamp and column to the table.
//
// Returns false if the column does not exist. Use RegisterColumn to add it first.
func (t *Tabulator) Set(timestamp time.Time, column string, value float64) bool {
	return t.addOrSet(timestamp, column, value, false)
}

func (t *Tabulator) addOrSet(timestamp time.Time, column string, value float64, add bool) bool {
	col, found := t.Columns.GetIndex(column)
	if !found {
		return false
	}

	var row int
	// perf tweak: if Data is mostly added in order, with many records for the same timestamp, cache the lastRow
	// to avoid map lookups in indexer.Add
	if timestamp.Equal(t.lastTimestamp) {
		row = t.lastRow
	} else {
		var added bool
		if row, added = t.Timestamps.Add(timestamp); added {
			t.Data = append(t.Data, make([]float64, t.Columns.Count()))
		}
		t.lastTimestamp = timestamp
		t.lastRow = row
	}

	if add {
		value += t.Data[row][col]
	}
	t.Data[row][col] = value
	return true
}

// RegisterColumn adds the specified columns to the table.
func (t *Tabulator) RegisterColumn(column ...string) {
	for _, c := range column {
		t.ensureColumnExists(c)
	}
}

func (t *Tabulator) ensureColumnExists(column string) {
	if _, added := t.Columns.Add(column); added {
		// new column. add Data for the new column to each row
		for key, entry := range t.Data {
			entry = append(entry, 0)
			t.Data[key] = entry
		}
	}
}

// Size returns the number of rows in the table.
func (t *Tabulator) Size() int {
	return len(t.Data)
}

// GetTimestamps returns the (sorted) list of timestamps in the table.
func (t *Tabulator) GetTimestamps() []time.Time {
	return t.Timestamps.List()
}

// GetColumns returns the (sorted) list of column names.
func (t *Tabulator) GetColumns() []string {
	return t.Columns.List()
}

// GetValues returns the value for the specified column for each timestamp in the table. The values are sorted by timestamp.
func (t *Tabulator) GetValues(columnName string) ([]float64, bool) {
	var values []float64
	column, ok := t.Columns.GetIndex(columnName)
	if ok {
		values = make([]float64, len(t.Data))
		for index, timestamp := range t.Timestamps.List() {
			row, _ := t.Timestamps.GetIndex(timestamp)
			values[index] = t.Data[row][column]
		}
	}
	return values, ok
}

// Accumulate increments the values in each column.
func (t *Tabulator) Accumulate() {
	accumulated := make([]float64, t.Columns.Count())

	for _, timestamp := range t.GetTimestamps() {
		row, _ := t.Timestamps.GetIndex(timestamp)
		for idx, value := range t.Data[row] {
			accumulated[idx] += value
		}
		copy(t.Data[row], accumulated)
	}
}

// Filter removes all rows that do not fall inside the specified time range. If the specified time is zero, it will be ignored.
func (t *Tabulator) Filter(from, to time.Time) {
	timestamps := make([]time.Time, 0, len(t.Data))
	d := make([][]float64, 0, len(t.Data))

	for _, timestamp := range t.GetTimestamps() {
		if !from.IsZero() && timestamp.Before(from) {
			continue
		}
		if !to.IsZero() && timestamp.After(to) {
			continue
		}
		row, _ := t.Timestamps.GetIndex(timestamp)
		timestamps = append(timestamps, timestamp)
		d = append(d, t.Data[row])
	}

	t.Timestamps = makeIndexerFromData(timestamps)
	t.Data = d
}

func (t *Tabulator) Copy() *Tabulator {
	result := &Tabulator{
		Timestamps: t.Timestamps.Copy(),
		Columns:    t.Columns.Copy(),
		Data:       make([][]float64, len(t.Data)),
	}
	for idx, row := range t.Data {
		result.Data[idx] = make([]float64, len(row))
		copy(result.Data[idx], row)
	}
	return result
}
