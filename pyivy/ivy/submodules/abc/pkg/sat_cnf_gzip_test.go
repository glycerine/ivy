package abc

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
)

func cnfGunzip(t testing.TB, data []byte) []byte {
	t.Helper()
	r, e := gzip.NewReader(bytes.NewReader(data))
	if e != nil {
		t.Fatal(e)
	}
	plain, e := io.ReadAll(r)
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Close(); e != nil {
		t.Fatal(e)
	}
	return plain
}
func TestCNFGZIPNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_gzip.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Readable, Prefix, Empty int
		GZIPHex                 string
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 10 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		variables := 5
		clauses := [][]SatLit{{0, 3, 4}, {}, {7, 8, 1, 2}, {5, 6, 9}}
		if row.Empty != 0 {
			variables = 0
			clauses = nil
		}
		f, e := NewCNFFormula(variables, clauses, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		options := CNFGZIPOptions{DIMACS: CNFDIMACSOptions{Readable: row.Readable != 0}}
		if row.Prefix&1 != 0 {
			options.DIMACS.ForAll = []SatVar{0, 2, 4}
		}
		if row.Prefix&2 != 0 {
			options.DIMACS.Exists = []SatVar{1, 3}
		}
		got, e := f.DIMACSGZIP(options)
		if e != nil {
			t.Fatal(e)
		}
		native, e := hex.DecodeString(row.GZIPHex)
		if e != nil {
			t.Fatal(e)
		}
		plain, e := f.DIMACS(options.DIMACS)
		if e != nil || !bytes.Equal(cnfGunzip(t, got), cnfGunzip(t, native)) || !bytes.Equal(cnfGunzip(t, got), plain) {
			t.Fatal(row, e)
		}
		again, e := f.DIMACSGZIP(options)
		if e != nil || !bytes.Equal(got, again) {
			t.Fatal("nondeterministic gzip", e)
		}
	}
}
func TestCNFGZIPLimitsCancellationAndWrappers(t *testing.T) {
	f, e := NewCNFFormula(3, [][]SatLit{{0, 3}, {4}}, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	data, e := f.DIMACSGZIP(CNFGZIPOptions{})
	if e != nil {
		t.Fatal(e)
	}
	exact, e := f.DIMACSGZIP(CNFGZIPOptions{MaxBytes: len(data)})
	if e != nil || !bytes.Equal(exact, data) {
		t.Fatal(e)
	}
	for _, o := range []CNFGZIPOptions{{MaxBytes: len(data) - 1}, {MaxBytes: 1}, {DIMACS: CNFDIMACSOptions{MaxBytes: 1}}} {
		got, e := f.DIMACSGZIP(o)
		if got != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(got, e)
		}
	}
	calls := 0
	_, e = f.DIMACS(CNFDIMACSOptions{Limits: CNFTransformOptions{Stop: func() bool { calls++; return false }}})
	if e != nil {
		t.Fatal(e)
	}
	threshold := calls + 1
	calls = 0
	out, e := f.DIMACSGZIP(CNFGZIPOptions{DIMACS: CNFDIMACSOptions{Limits: CNFTransformOptions{Stop: func() bool { calls++; return calls > threshold }}}})
	if out != nil || e != ErrCancelled {
		t.Fatal(out, e, calls, threshold)
	}
	g := mappedCNFFixture(t, 0, 0)
	c, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
	if e != nil {
		t.Fatal(e)
	}
	compressed, e := c.DIMACSGZIP(CNFGZIPOptions{})
	plain, e2 := c.DIMACS(CNFDIMACSOptions{})
	if e != nil || e2 != nil || !bytes.Equal(cnfGunzip(t, compressed), plain) {
		t.Fatal(e, e2)
	}
	frames, e := c.UnrollFrames(g, 2, false, CNFFrameOptions{})
	if e != nil {
		t.Fatal(e)
	}
	compressed, e = frames.DIMACSGZIP(CNFGZIPOptions{})
	plain, e2 = frames.DIMACS(CNFDIMACSOptions{})
	if e != nil || e2 != nil || !bytes.Equal(cnfGunzip(t, compressed), plain) {
		t.Fatal(e, e2)
	}
}
func FuzzCNFGZIP(f *testing.F) {
	f.Add([]byte{0, 3, 4, 7, 8, 1, 2, 5, 6, 9}, false, true)
	f.Fuzz(func(t *testing.T, data []byte, readable, prefix bool) {
		if len(data) > 512 {
			data = data[:512]
		}
		flat := make([]SatLit, len(data))
		for i, b := range data {
			flat[i] = SatLit(b % 16)
		}
		middle := len(flat) / 2
		formula, e := NewCNFFormula(8, [][]SatLit{flat[:middle], {}, flat[middle:]}, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		options := CNFGZIPOptions{DIMACS: CNFDIMACSOptions{Readable: readable}}
		if prefix {
			options.DIMACS.ForAll = []SatVar{0, 7}
			options.DIMACS.Exists = []SatVar{}
		}
		compressed, e := formula.DIMACSGZIP(options)
		if e != nil {
			t.Fatal(e)
		}
		plain, e := formula.DIMACS(options.DIMACS)
		if e != nil || !bytes.Equal(cnfGunzip(t, compressed), plain) {
			t.Fatal(e)
		}
	})
}
