package abc

// Tests for CRC-32 routines from misc/zlib/crc32.c.

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"os"
	"testing"
)

type zlibCRCOracle struct {
	Nil       uint32 `json:"nil"`
	TableHash string `json:"tableHash"`
	Updates   []struct {
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

func zlibCRCPattern(pattern, length int) []byte {
	data := make([]byte, length)
	for i := range data {
		data[i] = byte(i*67 + pattern*31 + (i >> 2))
	}
	return data
}

func zlibCRCTableHash(table [256]uint32) uint64 {
	hash := uint64(1469598103934665603)
	for _, value := range table {
		hash = (hash ^ uint64(value)) * 1099511628211
	}
	return hash
}

func TestZlibCRC32Native(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/zlib_crc32.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle zlibCRCOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Updates) != 32 || len(oracle.Combines) != 16 {
		t.Fatal(len(oracle.Updates), len(oracle.Combines), err)
	}
	if got := ZlibCRC32(0xdeadbeef, nil); got != oracle.Nil {
		t.Fatal(got, oracle.Nil)
	}
	if got := zlibCRCTableHash(ZlibCRC32Table()); got != parseHex64(t, oracle.TableHash) {
		t.Fatal(got, oracle.TableHash)
	}
	for _, row := range oracle.Updates {
		if got := ZlibCRC32(row.Initial, zlibCRCPattern(row.Pattern, row.Length)); got != row.Checksum {
			t.Fatal(row, got)
		}
	}
	for _, row := range oracle.Combines {
		if got := ZlibCRC32Combine(row.First, row.Second, row.SecondLength); got != row.Result {
			t.Fatal(row, got)
		}
	}
}

func parseHex64(t *testing.T, value string) uint64 {
	t.Helper()
	var result uint64
	if _, err := fmt.Sscanf(value, "%x", &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestZlibCRC32CompositionAndTableOwnership(t *testing.T) {
	firstTable := ZlibCRC32Table()
	firstTable[0] ^= 1
	if secondTable := ZlibCRC32Table(); firstTable[0] == secondTable[0] {
		t.Fatal("table aliases")
	}
	for firstLength := 0; firstLength <= 97; firstLength += 7 {
		for secondLength := 0; secondLength <= 89; secondLength += 9 {
			first := zlibCRCPattern(3, firstLength)
			second := zlibCRCPattern(5, secondLength)
			whole := append(append([]byte(nil), first...), second...)
			want := crc32.ChecksumIEEE(whole)
			firstSum := ZlibCRC32(0, first)
			secondSum := ZlibCRC32(0, second)
			if got := ZlibCRC32(firstSum, second); got != want {
				t.Fatal("update", firstLength, secondLength, got, want)
			}
			if got := ZlibCRC32Combine(firstSum, secondSum, uint64(secondLength)); got != want {
				t.Fatal("combine", firstLength, secondLength, got, want)
			}
		}
	}
}

func FuzzZlibCRC32(f *testing.F) {
	f.Add([]byte("first"), []byte("second"))
	f.Add([]byte{}, []byte{0, 1, 2, 255})
	f.Fuzz(func(t *testing.T, first, second []byte) {
		if len(first) > 8192 {
			first = first[:8192]
		}
		if len(second) > 8192 {
			second = second[:8192]
		}
		first = append([]byte{}, first...)
		second = append([]byte{}, second...)
		whole := append(append([]byte{}, first...), second...)
		want := crc32.ChecksumIEEE(whole)
		firstSum := ZlibCRC32(0, first)
		if got := ZlibCRC32(firstSum, second); got != want {
			t.Fatal("update", got, want)
		}
		secondSum := ZlibCRC32(0, second)
		if got := ZlibCRC32Combine(firstSum, secondSum, uint64(len(second))); got != want {
			t.Fatal("combine", got, want)
		}
	})
}
