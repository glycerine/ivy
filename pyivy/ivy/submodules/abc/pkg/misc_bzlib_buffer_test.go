package abc

// Tests for BZ2_bzlibVersion and BZ2_bzBuffToBuffDecompress in src/misc/bzlib/bzlib.c.

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

type bzip2BufferOracle struct {
	Version string `json:"version"`
	Cases   []struct {
		Level  int    `json:"level"`
		Status int    `json:"status"`
		Source string `json:"source"`
		Stream string `json:"stream"`
	} `json:"cases"`
}

func readBZip2BufferOracle(t testing.TB) bzip2BufferOracle {
	t.Helper()
	data, err := os.ReadFile("test_vectors/oracle/bzlib_buffer.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle bzip2BufferOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 6 {
		t.Fatal(len(oracle.Cases), err)
	}
	return oracle
}

func TestBZip2BufferNative(t *testing.T) {
	oracle := readBZip2BufferOracle(t)
	if oracle.Version != BZip2Version {
		t.Fatal(oracle.Version, BZip2Version)
	}
	for index, record := range oracle.Cases {
		if record.Status != 0 || (record.Level != 1 && record.Level != 9) {
			t.Fatal(index, record.Status, record.Level)
		}
		source, err := hex.DecodeString(record.Source)
		if err != nil {
			t.Fatal(index, err)
		}
		stream, err := hex.DecodeString(record.Stream)
		if err != nil {
			t.Fatal(index, err)
		}
		plain, err := DecompressBZip2(stream, int64(len(source)+1), nil)
		if err != nil || !slices.Equal(plain, source) {
			t.Fatal(index, len(plain), len(source), err)
		}
		if len(source) > 0 {
			exact, err := DecompressBZip2(stream, int64(len(source)), nil)
			if err != nil || !slices.Equal(exact, source) {
				t.Fatal("exact cap", index, len(exact), len(source), err)
			}
		}
		if len(source) > 1 {
			limited, err := DecompressBZip2(stream, int64(len(source)-1), nil)
			if limited != nil || !errors.Is(err, ErrLimit) {
				t.Fatal(index, len(limited), err)
			}
		}
	}
}

func TestBZip2BufferErrorsCancellationAndOwnership(t *testing.T) {
	oracle := readBZip2BufferOracle(t)
	stream, err := hex.DecodeString(oracle.Cases[4].Stream)
	if err != nil {
		t.Fatal(err)
	}
	if output, err := DecompressBZip2(stream, -1, nil); output != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(output, err)
	}
	if output, err := DecompressBZip2(stream, 0, func() bool { return true }); output != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(output, err)
	}
	for _, malformed := range [][]byte{nil, []byte("not bzip2"), stream[:len(stream)-1]} {
		if output, err := DecompressBZip2(malformed, 1024, nil); output != nil || !errors.Is(err, ErrFormat) {
			t.Fatal(len(malformed), output, err)
		}
	}
	first, err := DecompressBZip2(stream, 1024, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantFirst := first[0]
	first[0] ^= 0xff
	second, err := DecompressBZip2(stream, 1024, nil)
	if err != nil || second[0] != wantFirst {
		t.Fatal(second, err)
	}
}

func FuzzDecompressBZip2(f *testing.F) {
	oracle := readBZip2BufferOracle(f)
	for _, record := range oracle.Cases {
		stream, err := hex.DecodeString(record.Stream)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(stream)
	}
	f.Add([]byte{})
	f.Add([]byte("not bzip2"))
	f.Fuzz(func(t *testing.T, stream []byte) {
		if len(stream) > 4096 {
			stream = stream[:4096]
		}
		output, err := DecompressBZip2(stream, 8192, nil)
		if err == nil {
			if len(output) > 8192 {
				t.Fatal(len(output))
			}
			return
		}
		if output != nil || (!errors.Is(err, ErrFormat) && !errors.Is(err, ErrLimit)) {
			t.Fatal(len(output), err)
		}
	})
}
