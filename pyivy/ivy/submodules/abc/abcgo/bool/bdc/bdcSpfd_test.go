package bdc

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSPFDCountsAndCofactors(t *testing.T) {
	a, b := spfdTruths[0], spfdTruths[1]
	if Bdc_CountSpfd(a, a) != 1024 || Bdc_CountSpfd(a, b) != 512 || Bdc_SpfdAdjCost(a) != 64 {
		t.Fatal("counts")
	}
	for v := int32(0); v < 6; v++ {
		if Bdc_Cof6(spfdTruths[v], v, 0) != 0 || Bdc_Cof6(spfdTruths[v], v, 1) != ^uint64(0) {
			t.Fatal(v)
		}
	}
}
func TestSPFDDecomposeSmallTruths(t *testing.T) {
	for _, truth := range []uint64{0, ^uint64(0), spfdTruths[0], ^spfdTruths[1], spfdTruths[0] & spfdTruths[1], spfdTruths[0] ^ spfdTruths[1], (spfdTruths[0] & spfdTruths[1]) ^ spfdTruths[2]} {
		r, err := DecomposeSPFD(truth, 3, 30, 3)
		if err != nil {
			t.Fatal(truth, err)
		}
		got := r.Root.Truth
		if r.Constant != nil {
			if *r.Constant {
				got = ^uint64(0)
			} else {
				got = 0
			}
		} else if r.Complement {
			got = ^got
		}
		if got != truth || r.Expression() == "" {
			t.Fatal(truth, got, r.Expression())
		}
	}
	if _, err := DecomposeSPFD(0x123456789abcdef0, 6, 10, 0); err == nil {
		t.Fatal("unbounded fallback")
	}
}
func TestSPFDMarksAndHash(t *testing.T) {
	e := make([]Bdc_Ent_t, 8)
	e[1].iFan0 = BDC_TERM
	e[2].iFan0 = BDC_TERM
	e[3] = Bdc_Ent_t{iFan0: 1, iFan1: 2}
	e[4] = Bdc_Ent_t{iFan0: 3, iFan1: 2}
	e[5] = Bdc_Ent_t{iFan0: 3, iFan1: 1}
	if n := Bdc_SpfdCheckOverlap(e, &e[4], &e[5]); n != 1 {
		t.Fatal(n)
	}
	for _, x := range e {
		if x.fMark0 || x.fMark1 {
			t.Fatal("mark leak")
		}
	}
	table := make([]Bdc_Ent_t, 100)
	f := uint64(0x123456789abcdef0)
	slot := Bdc_SpfdHashLookup(table, 53, f)
	*slot = 60
	table[60].Truth = f
	if Bdc_SpfdHashLookup(table, 53, f) != nil {
		t.Fatal("duplicate hash")
	}
	slot = Bdc_SpfdHashLookup(table, 53, 0)
	if slot == nil {
		t.Fatal("insert slot")
	}
}
func TestSPFDDictionaryAndCover(t *testing.T) {
	truths, weights, err := GenerateSPFDDictionary(3, 2, 1000)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[uint64]bool{}
	for i, f := range truths {
		if seen[f] || f&1 != 0 {
			t.Fatal("duplicate/phase", f)
		}
		seen[f] = true
		if i < 3 && weights[i] != 0 {
			t.Fatal(weights)
		}
	}
	for _, truth := range []uint64{spfdTruths[0], spfdTruths[0] ^ spfdTruths[1], spfdTruths[0] & spfdTruths[1]} {
		if _, err := DecomposeSPFDWithDivisors(truth, truths, weights); err != nil {
			t.Fatal(truth, err)
		}
	}
	if _, _, err := FindBestSPFD(nil, nil, 1, 2); err == nil {
		t.Fatal("missing divisor")
	}
	a, b, err := GenerateSPFDDictionary(3, 6, 3)
	if err != nil || len(a) != 3 || len(b) != 3 {
		t.Fatal(len(a), len(b), err)
	}
}
func TestSPFDDictionaryIOAndBenchmark(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "truths"), filepath.Join(dir, "weights")
	truths, weights := []uint64{1, 0xabcdef1234567890}, []int32{3, 4}
	if err := WriteSPFDDictionary(a, b, truths, weights); err != nil {
		t.Fatal(err)
	}
	got, w, err := ReadSPFDDictionary(a, b)
	if err != nil || got[1] != truths[1] || w[1] != 4 {
		t.Fatal(got, w, err)
	}
	if err := os.WriteFile(b, []byte{1}, 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadSPFDDictionary(a, b); err == nil {
		t.Fatal("corrupt dictionary")
	}
	var out bytes.Buffer
	if err := SPFDSubsetBenchmark(&out, 100, 5, 1); err != nil || strings.Count(out.String(), "Total =") != 2 {
		t.Fatal(err, out.String())
	}
}

func TestSPFDConstantCoverHasZeroCost(t *testing.T) {
	for _, f := range []uint64{0, ^uint64(0)} {
		cost, err := DecomposeSPFDWithDivisors(f, nil, nil)
		if err != nil || cost != 0 {
			t.Fatal(cost, err)
		}
	}
}

type spfdFailWriter struct{}

func (spfdFailWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
func TestSPFDBenchmarkWriteError(t *testing.T) {
	if err := SPFDSubsetBenchmark(spfdFailWriter{}, 4, 2, 1); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
}
