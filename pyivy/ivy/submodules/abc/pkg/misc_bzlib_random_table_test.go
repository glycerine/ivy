package abc

// Tests for src/misc/bzlib/randtable.c.

import (
	"encoding/json"
	"os"
	"testing"
)

func TestBZip2RandomNumbersNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/bzlib_random_numbers.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []int32
	if err = json.Unmarshal(data, &native); err != nil || len(native) != 512 {
		t.Fatal(err, len(native))
	}
	table := BZip2RandomNumbers()
	for index, value := range native {
		if table[index] != value {
			t.Fatal(index, table[index], value)
		}
		if value < 50 || value > 999 {
			t.Fatal("native value outside documented table range", index, value)
		}
	}
	table[0] ^= 1
	if fresh := BZip2RandomNumbers(); fresh[0] != native[0] {
		t.Fatal("returned table aliases package storage")
	}
}
