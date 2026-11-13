package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func naiveTestGraph(unequal bool) *gia.Gia_Man_t {
	g := gia.Gia_ManStart(24)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c)
	y := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, c), b)
	if unequal {
		y = gia.Gia_ManAppendAnd(g, a, b)
	}
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y)
	return g
}
func TestNaiveMappedCNFChecksBothDirectionsAndKeepsSourceState(t *testing.T) {
	for _, different := range []bool{false, true} {
		g := naiveTestGraph(different)
		p := Cec_ParCec_t{nBTLimit: 1000, iOutFail: 17}
		got := Cec_ManVerifyNaive(g, &p)
		want := int32(1)
		if different {
			want = 0
		}
		if got != want || p.Error != nil || p.iOutFail != 17 || g.CombinationalCounterexample() != nil {
			t.Fatalf("different %v status %d error %v output %d", different, got, p.Error, p.iOutFail)
		}
	}
}
func TestNaiveVerificationAgainstOriginalABC(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	binary := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../abc"))
	if _, err := os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	for _, different := range []bool{false, true} {
		g := naiveTestGraph(different)
		data, err := gia.MarshalAiger(g, gia.AigerWriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "input.aig")
		if err = os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		log, err := exec.Command(binary, "-c", fmt.Sprintf("&r -s %s; &cec -n -m -d", path)).CombinedOutput()
		if err != nil {
			t.Fatalf("ABC: %v %s", err, log)
		}
		text := string(log)
		want := int32(-1)
		for _, line := range strings.Split(text, "\n") {
			var unsat, sat, undec int
			if n, _ := fmt.Sscanf(line, "UNSAT = %d.  SAT = %d.   UNDEC = %d.", &unsat, &sat, &undec); n == 3 {
				want = 1
				if sat > 0 {
					want = 0
				} else if undec > 0 {
					want = -1
				}
			}
		}
		p := Cec_ParCec_t{nBTLimit: 1000}
		got := Cec_ManVerifyNaive(g, &p)
		if got != want {
			t.Fatalf("different %v: Go %d ABC %d: %s", different, got, want, text)
		}
	}
}

func TestNaiveVerificationKeepsSourceWitnessAndChecksAllPairs(t *testing.T) {
	g := gia.Gia_ManStart(20)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	ab := gia.Gia_ManAppendAnd(g, a, b)
	x := gia.Gia_ManAppendAnd(g, ab, c)
	y := gia.Gia_ManAppendAnd(g, a, gia.Gia_ManAppendAnd(g, b, c))
	gia.Gia_ManAppendCo(g, ab)
	gia.Gia_ManAppendCo(g, ab) // trivial pair
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y) // nontrivial UNSAT pair
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, ab) // SAT pair
	if err := g.SetCombinationalCounterexample([]bool{true, true, false}, 2); err != nil {
		t.Fatal(err)
	}
	before := g.CombinationalCounterexample()
	p := Cec_ParCec_t{nBTLimit: 1000, iOutFail: 19}
	if got := Cec_ManVerifyNaive(g, &p); got != 0 || p.Error != nil {
		t.Fatalf("status %d error %v", got, p.Error)
	}
	if g.CombinationalCounterexample() != before || p.iOutFail != 19 {
		t.Fatal("naive source routine unexpectedly replaced caller state")
	}
}

func TestNaiveVerificationConstantOutputs(t *testing.T) {
	for _, different := range []bool{false, true} {
		g := gia.Gia_ManStart(4)
		gia.Gia_ManAppendCo(g, 0)
		second := int32(0)
		if different {
			second = 1
		}
		gia.Gia_ManAppendCo(g, second)
		p := Cec_ParCec_t{nBTLimit: 100}
		want := int32(1)
		if different {
			want = 0
		}
		if got := Cec_ManVerifyNaive(g, &p); got != want || p.Error != nil {
			t.Fatalf("different %v status %d error %v", different, got, p.Error)
		}
	}
}

// MF stores variables by coarsened object ID; reordered COs must map back to
// their original object IDs just as Gia_ManGenerateCnf does in ABC.
func TestNaiveMappedCNFReorderedOutputs(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x := gia.Gia_ManAppendCi(g)
	for _, lit := range []int32{0, 0, x, x} {
		gia.Gia_ManAppendCo(g, lit)
	}
	ids := make([]int32, 4)
	for i := range ids {
		ids[i] = gia.Gia_ObjId(g, gia.Gia_ManCo(g, int32(i)))
	}
	ids[1], ids[2] = ids[2], ids[1]
	data, vars, err := cecNaiveMappedCNF(g, ids)
	if err != nil {
		t.Fatal(err)
	}
	solver := cnf.Cnf_DataWriteIntoSolver(&cnf.Cnf_Dat_t{Variables: data.Variables, Clauses: data.Clauses}, 1, 0)
	if solver == nil {
		t.Fatal("cannot load CNF")
	}
	// ABC maps physical-order coarsened outputs back in logical CO order.
	// Thus the reordered, functionally different pairs are incorrectly UNSAT.
	for _, pair := range [][2]int{{0, 1}, {2, 3}} {
		a, b := vars[ids[pair[0]]], vars[ids[pair[1]]]
		if solver.Solve(bsat.Positive(a), bsat.Negative(b)) != bsat.Unsatisfiable || solver.Solve(bsat.Negative(a), bsat.Positive(b)) != bsat.Unsatisfiable {
			t.Fatalf("reordered pair %v differs from ABC", pair)
		}
	}
}
