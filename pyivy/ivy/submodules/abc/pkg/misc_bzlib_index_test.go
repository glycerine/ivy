package abc

// Tests for BZ2_indexIntoF in src/misc/bzlib/bzlib.c.

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func bzip2IndexTable(test int) []int32 {
	result := make([]int32, 257)
	for index := 0; index < 256; index++ {
		result[index+1] = result[index] + int32(((index+3)*(test+5))%11)
	}
	return result
}

func TestBZip2IndexIntoFNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/bzlib_index.json")
	if err != nil {
		t.Fatal(err)
	}
	var records [][3]int32
	if err = json.Unmarshal(data, &records); err != nil || len(records) != 144 {
		t.Fatal(err, len(records))
	}
	for _, record := range records {
		result, err := BZip2IndexIntoF(record[1], bzip2IndexTable(int(record[0])))
		if err != nil || result != int(record[2]) {
			t.Fatal(record, result, err)
		}
	}
	if result, err := BZip2IndexIntoF(0, make([]int32, 256)); result != 0 || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	invalid := make([]int32, 257)
	invalid[8], invalid[9] = 2, 1
	if result, err := BZip2IndexIntoF(0, invalid); result != 0 || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
}

func FuzzBZip2IndexIntoF(f *testing.F) {
	f.Add([]byte{1, 0, 2, 3}, int32(2))
	f.Add([]byte{}, int32(-1))
	f.Fuzz(func(t *testing.T, increments []byte, query int32) {
		if len(increments) > 256 {
			increments = increments[:256]
		}
		cumulative := make([]int32, 257)
		for index := 0; index < 256; index++ {
			if index < len(increments) {
				cumulative[index+1] = cumulative[index] + int32(increments[index])
			} else {
				cumulative[index+1] = cumulative[index]
			}
		}
		got, err := BZip2IndexIntoF(query, cumulative)
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		for index := 1; index < 256; index++ {
			if query >= cumulative[index] {
				want = index
			}
		}
		if got != want {
			t.Fatal(query, got, want)
		}
	})
}
