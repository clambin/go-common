package tabulator

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"time"
)

var _ json.Marshaler = &Tabulator{}
var _ json.Unmarshaler = &Tabulator{}

func (t *Tabulator) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.saveToTabulatorAsJSON())
}

func (t *Tabulator) UnmarshalJSON(bytes []byte) error {
	var j tabulatorAsJSON
	err := json.Unmarshal(bytes, &j)
	if err == nil {
		t.loadFromTabulatorAsJSON(j)
	}
	return err
}

type tabulatorAsJSON struct {
	Timestamps []time.Time
	Columns    []string
	Data       [][]float64
}

func (t *Tabulator) saveToTabulatorAsJSON() tabulatorAsJSON {
	j := tabulatorAsJSON{
		Timestamps: t.GetTimestamps(),
		Columns:    t.GetColumns(),
		Data:       make([][]float64, t.Timestamps.Count()),
	}
	for index := range j.Timestamps {
		j.Data[index] = make([]float64, len(j.Columns))
	}

	for column, columnName := range j.Columns {
		vals, _ := t.GetValues(columnName)

		for row := range vals {
			j.Data[row][column] = vals[row]
		}
	}

	return j
}

func (t *Tabulator) loadFromTabulatorAsJSON(j tabulatorAsJSON) {
	t2 := New(j.Columns...)
	for index, timestamp := range j.Timestamps {
		for column, value := range j.Data[index] {
			t2.Set(timestamp, j.Columns[column], value)
		}
	}
	*t = *t2
}

func (t *Tabulator) SaveBinary() ([]byte, error) {
	var output bytes.Buffer
	err := gob.NewEncoder(&output).Encode(t)
	return output.Bytes(), err
}
