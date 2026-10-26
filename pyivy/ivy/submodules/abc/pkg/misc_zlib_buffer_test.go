package abc

// Tests for memory-buffer zlib operations from misc/zlib/compress_.c and uncompr.c.

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"slices"
	"testing"
)

type zlibBufferOracleRow struct {
	Pattern    int    `json:"pattern"`
	Level      int    `json:"level"`
	Bound      uint64 `json:"bound"`
	Compressed string `json:"compressed"`
}

func zlibOraclePattern(pattern int) []byte {
	switch pattern {
	case 0:
		return nil
	case 1:
		return []byte{0}
	case 2:
		return []byte("abc")
	case 3:
		return slices.Repeat([]byte{'A'}, 128)
	case 4:
		result := make([]byte, 256)
		for i := range result {
			result[i] = byte(i)
		}
		return result
	case 5:
		result := make([]byte, 257)
		for i := range result {
			result[i] = byte(i*73 + 19)
		}
		return result
	case 6:
		result := make([]byte, 4096)
		for i := range result {
			result[i] = byte("ABC-logic-"[i%10])
		}
		return result
	}
	return nil
}

func TestZlibBufferNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/zlib_buffer.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []zlibBufferOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 7 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		source := zlibOraclePattern(row.Pattern)
		bound, err := ZlibCompressBound(uint64(len(source)))
		if err != nil || bound != row.Bound {
			t.Fatal(row, bound, err)
		}
		native, err := hex.DecodeString(row.Compressed)
		if err != nil {
			t.Fatal(err)
		}
		plain, err := ZlibUncompress(native, ZlibBufferOptions{MaxOutput: len(source) + 1})
		if err != nil || !slices.Equal(plain, source) {
			t.Fatal(row.Pattern, err, plain)
		}
		compressed, err := ZlibCompress(source, row.Level, ZlibBufferOptions{MaxOutput: int(bound)})
		if err != nil || len(compressed) > int(bound) {
			t.Fatal(row.Pattern, len(compressed), bound, err)
		}
		roundTrip, err := ZlibUncompress(compressed, ZlibBufferOptions{MaxOutput: len(source) + 1})
		if err != nil || !slices.Equal(roundTrip, source) {
			t.Fatal(row.Pattern, err)
		}
	}
}

func TestZlibBufferErrorsOwnershipAndLimits(t *testing.T) {
	for _, level := range []int{-2, 10} {
		if output, err := ZlibCompress(nil, level, ZlibBufferOptions{}); output != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(level, output, err)
		}
	}
	if _, err := ZlibCompressBound(math.MaxUint64); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if output, err := ZlibCompress([]byte("x"), 6, ZlibBufferOptions{MaxOutput: -1}); output != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(output, err)
	}
	if output, err := ZlibUncompress([]byte("not zlib"), ZlibBufferOptions{}); output != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(output, err)
	}
	compressed, err := ZlibCompress([]byte("longer output"), 6, ZlibBufferOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if output, err := ZlibUncompress(compressed, ZlibBufferOptions{MaxOutput: len("longer output")}); err != nil || string(output) != "longer output" {
		t.Fatal(string(output), err)
	}
	if output, err := ZlibUncompress(compressed, ZlibBufferOptions{MaxOutput: 4}); output != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(output, err)
	}
	if output, err := ZlibCompress([]byte("x"), 6, ZlibBufferOptions{MaxOutput: 1}); output != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(output, err)
	}
	for _, operation := range []func(func() bool) ([]byte, error){
		func(stop func() bool) ([]byte, error) {
			return ZlibCompress([]byte("cancel"), 6, ZlibBufferOptions{Stop: stop})
		},
		func(stop func() bool) ([]byte, error) {
			return ZlibUncompress(compressed, ZlibBufferOptions{Stop: stop})
		},
	} {
		if output, err := operation(func() bool { return true }); output != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(output, err)
		}
	}
	source := []byte("owned")
	output, err := ZlibCompress(source, 6, ZlibBufferOptions{})
	if err != nil {
		t.Fatal(err)
	}
	source[0] = 'X'
	plain, err := ZlibUncompress(output, ZlibBufferOptions{})
	if err != nil || string(plain) != "owned" {
		t.Fatal(string(plain), err)
	}
}

func FuzzZlibBuffer(f *testing.F) {
	f.Add([]byte("ABC"), int8(-1))
	f.Add([]byte{0, 1, 2, 3, 255}, int8(9))
	f.Fuzz(func(t *testing.T, source []byte, rawLevel int8) {
		if len(source) > 4096 {
			source = source[:4096]
		}
		level := int(rawLevel)%11 - 1
		if level < -1 || level > 9 {
			return
		}
		bound, err := ZlibCompressBound(uint64(len(source)))
		if err != nil {
			t.Fatal(err)
		}
		compressed, err := ZlibCompress(source, level, ZlibBufferOptions{MaxOutput: int(bound)})
		if err != nil {
			t.Fatal(level, len(source), bound, err)
		}
		plain, err := ZlibUncompress(compressed, ZlibBufferOptions{MaxOutput: len(source) + 1})
		if err != nil || !slices.Equal(plain, source) {
			t.Fatal(level, len(source), err)
		}
	})
}

func FuzzZlibUncompressRobust(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("not zlib"))
	f.Add([]byte{0x78, 0x9c, 0x03, 0x00, 0x00, 0x00, 0x00, 0x01})
	f.Fuzz(func(t *testing.T, compressed []byte) {
		if len(compressed) > 4096 {
			compressed = compressed[:4096]
		}
		output, err := ZlibUncompress(compressed, ZlibBufferOptions{MaxOutput: 8192})
		if err == nil {
			if len(output) > 8192 {
				t.Fatal(len(output))
			}
			return
		}
		if output != nil || (!errors.Is(err, ErrInvariant) && !errors.Is(err, ErrLimit)) {
			t.Fatal(len(output), err)
		}
	})
}
