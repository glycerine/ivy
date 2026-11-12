package lucky

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestHexAndStoreRoundtrip(t *testing.T) {
	for _, vars := range []int32{2, 6, 7, 16} {
		table := make([]uint64, Kit_TruthWordNum_64bit(vars))
		for i := range table {
			table[i] = uint64(i)*0x0123456789abcdef + 0xdeadbeef
		}
		if vars < 6 {
			table[0] &= (1 << uint(1<<uint(vars))) - 1
		}
		var buffer bytes.Buffer
		if err := Abc_TruthWriteHex(&buffer, table, vars); err != nil {
			t.Fatal(err)
		}
		restored := make([]uint64, len(table))
		if err := Abc_TruthReadHex(restored, "0x"+buffer.String(), vars); err != nil || !slices.Equal(restored, table) {
			t.Fatal(vars, err, restored, table)
		}
	}
	store := Abc_TruthStoreAlloc(7, 2)
	store.Funcs[0][0] = 1
	store.Funcs[1][1] = 2
	name := filepath.Join(t.TempDir(), "truths.txt")
	if err := Abc_TruthStoreWrite(name, store); err != nil {
		t.Fatal(err)
	}
	restored, err := ReadTruthStore(name)
	if err != nil || restored.NFuncs != 2 || restored.NVars != 7 || !slices.Equal(restored.Funcs[1], store.Funcs[1]) {
		t.Fatal(restored, err)
	}
	empty := filepath.Join(t.TempDir(), "empty.txt")
	if err = os.WriteFile(empty, nil, 0600); err != nil {
		t.Fatal(err)
	}
	store, err = ReadTruthStore(empty)
	if err != nil || store.NFuncs != 0 {
		t.Fatal(store, err)
	}
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }
func TestHexShortWrite(t *testing.T) {
	if err := Abc_TruthWriteHex(shortWriter{}, []uint64{0}, 6); err != io.ErrShortWrite {
		t.Fatalf("short write silently accepted: %v", err)
	}
}
func TestSimpleSemiCanonicalizerNilStore(t *testing.T) {
	table := []uint64{0x0123456789abcdef}
	Kit_TruthSemiCanonicize_Yasha_simple(table, 6, nil)
}

func TestPrefixedStoreWrite(t *testing.T) {
	store := Abc_TruthStoreAlloc(7, 1)
	store.Funcs[0][0] = 1
	store.Funcs[0][1] = 2
	name := filepath.Join(t.TempDir(), "prefixed.txt")
	if err := WriteToFile1(name, store, store.Funcs); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	want := "0x00000000000000020000000000000001\n\n"
	if string(data) != want {
		t.Fatalf("got %q, want %q", data, want)
	}
	restored, err := ReadTruthStore(name)
	if err != nil || !slices.Equal(restored.Funcs[0], store.Funcs[0]) {
		t.Fatal(restored, err)
	}
}
