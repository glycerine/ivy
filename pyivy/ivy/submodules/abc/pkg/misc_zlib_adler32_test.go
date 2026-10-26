package abc

// Tests for Adler-32 routines from misc/zlib/adler32.c.

import (
	"encoding/json"
	"hash/adler32"
	"os"
	"testing"
)

type zlibAdlerOracle struct {
	Nil     uint32 `json:"nil"`
	Updates []struct {
		Initial  uint32 `json:"initial"`
		Pattern  int    `json:"pattern"`
		Length   int    `json:"length"`
		Checksum uint32 `json:"checksum"`
	} `json:"updates"`
	Combines []struct {
		First        uint32 `json:"first"`
		Second       uint32 `json:"second"`
		SecondLength uint64 `json:"secondLength"`
		Result       uint32 `json:"result"`
	} `json:"combines"`
}

func zlibAdlerPattern(pattern, length int) []byte {
	data := make([]byte, length)
	for i := range data {
		data[i] = byte(i*73 + pattern*29 + (i >> 3))
	}
	return data
}

func TestZlibAdler32Native(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/zlib_adler32.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle zlibAdlerOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Updates) != 32 || len(oracle.Combines) != 16 {
		t.Fatal(len(oracle.Updates), len(oracle.Combines), err)
	}
	if got := ZlibAdler32(0xdeadbeef, nil); got != oracle.Nil {
		t.Fatal(got, oracle.Nil)
	}
	for _, row := range oracle.Updates {
		if got := ZlibAdler32(row.Initial, zlibAdlerPattern(row.Pattern, row.Length)); got != row.Checksum {
			t.Fatal(row, got)
		}
	}
	for _, row := range oracle.Combines {
		if got := ZlibAdler32Combine(row.First, row.Second, row.SecondLength); got != row.Result {
			t.Fatal(row, got)
		}
	}
}

func TestZlibAdler32Composition(t *testing.T) {
	for firstLength := 0; firstLength <= 97; firstLength += 7 {
		for secondLength := 0; secondLength <= 89; secondLength += 9 {
			first := zlibAdlerPattern(3, firstLength)
			second := zlibAdlerPattern(5, secondLength)
			whole := append(append([]byte(nil), first...), second...)
			want := adler32.Checksum(whole)
			firstSum := ZlibAdler32(1, first)
			secondSum := ZlibAdler32(1, second)
			if got := ZlibAdler32(firstSum, second); got != want {
				t.Fatal("update", firstLength, secondLength, got, want)
			}
			if got := ZlibAdler32Combine(firstSum, secondSum, uint64(secondLength)); got != want {
				t.Fatal("combine", firstLength, secondLength, got, want)
			}
		}
	}
}

func FuzzZlibAdler32(f *testing.F) {
	f.Add([]byte("first"), []byte("second"))
	f.Add([]byte{}, []byte{0, 1, 2, 255})
	f.Fuzz(func(t *testing.T, first, second []byte) {
		if len(first) > 8192 {
			first = first[:8192]
		}
		if len(second) > 8192 {
			second = second[:8192]
		}
		// Fuzz byte slices can be nil; use non-nil zero-length slices for the
		// streaming calls because nil intentionally means the native query.
		first = append([]byte{}, first...)
		second = append([]byte{}, second...)
		whole := append(append([]byte{}, first...), second...)
		want := adler32.Checksum(whole)
		firstSum := ZlibAdler32(1, first)
		if got := ZlibAdler32(firstSum, second); got != want {
			t.Fatal("update", got, want)
		}
		secondSum := ZlibAdler32(1, second)
		if got := ZlibAdler32Combine(firstSum, secondSum, uint64(len(second))); got != want {
			t.Fatal("combine", got, want)
		}
	})
}
