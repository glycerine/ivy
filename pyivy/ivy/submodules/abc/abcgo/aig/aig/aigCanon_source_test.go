package aig

import (
	"fmt"
	"math/bits"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/ioa"
)

func TestCanonicalTruthAgainstCompiledSource(t *testing.T) {
	var input, want strings.Builder
	check := func(truth []uint32, n int32) {
		for _, ret := range []int32{0, 1} {
			fmt.Fprintf(&input, "%d %d", n, ret)
			for _, word := range truth {
				fmt.Fprintf(&input, " %d", word)
			}
			input.WriteByte('\n')
			in, out := append([]uint32(nil), truth...), make([]uint32, len(truth))
			perm := make([]byte, n)
			for i := range perm {
				perm[i] = byte(i)
			}
			sigs := make([]Aig_VSig_t, 2*n)
			phase := Aig_RManSemiCanonicize(out, in, n, perm, sigs, ret)
			result := out
			if ret != 0 {
				result = in
			}
			fmt.Fprintf(&want, "%d %d", phase, Aig_RManVarsAreUnique(sigs, n))
			for _, v := range perm {
				fmt.Fprintf(&want, " %d", v)
			}
			for _, word := range result {
				fmt.Fprintf(&want, " %d", word)
			}
			for _, sig := range sigs {
				fmt.Fprintf(&want, " %d", sig.nOnes)
				for _, v := range sig.nCofOnes[:n] {
					fmt.Fprintf(&want, " %d", v)
				}
			}
			want.WriteByte('\n')
		}
	}
	check([]uint32{0}, 0)
	check([]uint32{^uint32(0)}, 0)
	for n := int32(1); n <= 3; n++ {
		for bits := uint32(0); bits < 1<<uint(1<<uint(n)); bits++ {
			var truth uint32
			for bit := uint32(0); bit < 32; bit++ {
				truth |= (bits >> (bit & ((1 << uint(n)) - 1)) & 1) << bit
			}
			check([]uint32{truth}, n)
		}
	}
	r := NewAigRandomSource()
	for n := int32(4); n <= RMAN_MAXVARS; n++ {
		for trial := 0; trial < 8; trial++ {
			truth := make([]uint32, canonWords(n))
			for i := range truth {
				truth[i] = r.Uint32()
			}
			if n == 4 {
				truth[0] = truth[0]&0xffff | truth[0]<<16
			}
			check(truth, n)
		}
	}
	out := runAigSourceFixtureInput(t, "canonical_truth_compare.c", input.String(), "../../bool/kit/kitTruth.c")
	if out != want.String() {
		t.Fatal("canonical truth/signatures/permutations differ from compiled original C")
	}
}

func TestCanonicalResizeUsesSourceInclusivePrime(t *testing.T) {
	p := &Aig_RMan_t{nBins: 1, pBins: make([]*Aig_Tru_t, 1)}
	Aig_RManTableResize(p)
	if p.nBins != 3 {
		t.Fatalf("resized to %d bins, want source's smallest odd prime >= 3", p.nBins)
	}
}

func TestCanonicalEmptyVariableDiagnosticMatchesSource(t *testing.T) {
	text := captureAigOutput(t, func() { Aig_RManPrintUniqueVars(nil, 0) })
	if text != "\n0\n" {
		t.Fatalf("empty signature diagnostic = %q, want source's blank phase line and 0", text)
	}
}

func TestCanonicalRecordedLibraryTruthRoundTrip(t *testing.T) {
	p := Aig_RManStart()
	defer Aig_RManStop(p)
	parity := make([]uint32, canonWords(RMAN_MAXVARS))
	for mask := 0; mask < 1<<RMAN_MAXVARS; mask++ {
		if bits.OnesCount(uint(mask))&1 != 0 {
			canonSet(parity, int32(mask))
		}
	}
	cases := []struct {
		n     int32
		truth []uint32
	}{
		{0, []uint32{0}}, {0, []uint32{^uint32(0)}},
		{1, []uint32{0xaaaaaaaa}}, {1, []uint32{0x55555555}},
		{4, []uint32{0x96e896e8}}, {6, []uint32{0x175ae987, 0xb846137e}},
		{RMAN_MAXVARS, parity},
	}
	for _, tc := range cases {
		Aig_RManSaveOne(p, tc.truth, tc.n)
	}
	if p.pAig.OutputCount() != len(cases) {
		t.Fatalf("saved %d outputs, want %d", p.pAig.OutputCount(), len(cases))
	}
	name := filepath.Join(t.TempDir(), "canonical.aig")
	if err := ioa.Ioa_WriteAiger(p.pAig, name, 0, 1); err != nil {
		t.Fatal(err)
	}
	m, err := ioa.Ioa_ReadAiger(name, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Check(); err != nil {
		t.Fatal(err)
	}
	if len(m.Outputs) != len(cases) || len(m.Inputs) != RMAN_MAXVARS {
		t.Fatal("library interface changed")
	}
	// Evaluate parsed AIGER literals independently of Graph.Evaluate, including
	// AIGER's false-at-zero constant convention and complemented root outputs.
	values := make([]bool, m.MaxVar+1)
	value := func(lit uint32) bool { return values[lit>>1] != (lit&1 != 0) }
	for mask := 0; mask < 1<<RMAN_MAXVARS; mask++ {
		for i, lit := range m.Inputs {
			values[lit>>1] = mask&(1<<i) != 0
		}
		for _, gate := range m.ANDs {
			values[gate.Literal>>1] = value(gate.Left) && value(gate.Right)
		}
		for i, tc := range cases {
			assignment := int32(mask & ((1 << uint(tc.n)) - 1))
			if got, want := value(m.Outputs[i]), canonGet(tc.truth, assignment) != 0; got != want {
				t.Fatalf("output %d assignment %d: got %v want %v", i, mask, got, want)
			}
		}
	}
}
