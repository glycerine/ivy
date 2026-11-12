package mf

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

func coarseTestAnd(n *testNetwork, a, b int32) int32 {
	if a > b {
		a, b = b, a
	}
	lit := int32(len(n.gates)) * 2
	n.gates = append(n.gates, testGate{a: a, b: b, kind: 2})
	return lit
}
func coarseTestMux(n *testNetwork, c, t, e int32) int32 {
	p, q := coarseTestAnd(n, c, t), coarseTestAnd(n, c^1, e)
	return coarseTestAnd(n, p^1, q^1) ^ 1
}
func coarseTestEval(n Network, in []bool) []bool {
	values := make([]bool, n.ObjectCount())
	ci := 0
	value := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
	for id := 0; id < n.ObjectCount(); id++ {
		a, b, c, k := n.Gate(int32(id))
		switch k {
		case 1:
			values[id] = in[ci]
			ci++
		case 2:
			values[id] = value(a) && value(b)
		case 3:
			values[id] = value(a) != value(b)
		case 4:
			if value(a) {
				values[id] = value(b)
			} else {
				values[id] = value(c)
			}
		case 5, 6:
			values[id] = value(a)
		}
	}
	out := make([]bool, len(n.OutputLiterals()))
	for i, l := range n.OutputLiterals() {
		out[i] = value(l)
	}
	return out
}
func coarseTestEquivalent(t *testing.T, a, b Network) {
	t.Helper()
	for mask := 0; mask < 1<<uint(len(a.InputLiterals())); mask++ {
		in := make([]bool, len(a.InputLiterals()))
		for i := range in {
			in[i] = mask&(1<<uint(i)) != 0
		}
		x, y := coarseTestEval(a, in), coarseTestEval(b, in)
		if fmt.Sprint(x) != fmt.Sprint(y) {
			t.Fatalf("coarsening changed truth at %v: %v / %v", in, x, y)
		}
	}
}
func coarseTestCounts(n Network) (ands, xors, muxes int) {
	for id := 0; id < n.ObjectCount(); id++ {
		_, _, _, kind := n.Gate(int32(id))
		switch kind {
		case 2:
			ands++
		case 3:
			xors++
		case 4:
			muxes++
		}
	}
	return
}
func TestCoarsenAllMuxAndXorPhases(t *testing.T) {
	for phase := int32(0); phase < 8; phase++ {
		n := &testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {kind: 1}, {kind: 1}}, inputs: []int32{2, 4, 6}}
		root := coarseTestMux(n, 2^(phase&1), 4^((phase>>1)&1), 6^((phase>>2)&1))
		n.outputs = []int32{root}
		for _, limit := range []int{0, 1, 2, 3} {
			out, err := Coarsen(n, limit)
			if err != nil {
				t.Fatal(err)
			}
			coarseTestEquivalent(t, n, out)
			_, _, muxes := coarseTestCounts(out)
			if (limit >= 2) != (muxes == 1) {
				t.Fatal("mux reference threshold", phase, limit, muxes)
			}
		}
	}
	for phase := int32(0); phase < 4; phase++ {
		n := &testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {kind: 1}}, inputs: []int32{2, 4}}
		a, b := 2^(phase&1), 4^((phase>>1)&1)
		p, q := coarseTestAnd(n, a, b^1), coarseTestAnd(n, a^1, b)
		n.outputs = []int32{coarseTestAnd(n, p^1, q^1) ^ 1}
		out, err := Coarsen(n, 0)
		if err != nil {
			t.Fatal(err)
		}
		coarseTestEquivalent(t, n, out)
		ands, xors, muxes := coarseTestCounts(out)
		if ands != 0 || xors != 1 || muxes != 0 {
			t.Fatal("XOR not recognized at limit zero", phase, ands, xors, muxes)
		}
	}
}
func TestCoarsenReferencesCleanupAndOwnership(t *testing.T) {
	n := &testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {kind: 1}, {kind: 1}}, inputs: []int32{2, 4, 6}}
	root := coarseTestMux(n, 2, 4, 6)
	n.outputs = []int32{root, 8, 8}
	coarseTestAnd(n, 2, 6)
	out, err := Coarsen(n, 3)
	if err != nil {
		t.Fatal(err)
	}
	coarseTestEquivalent(t, n, out)
	_, _, muxes := coarseTestCounts(out)
	if muxes != 0 {
		t.Fatal("shared child ignored reference limit")
	}
	before := append([]int32(nil), n.outputs...)
	n.outputs = nil
	n.gates = append(n.gates, testGate{a: root, kind: 6})
	n.outputs = []int32{int32(len(n.gates)-1) * 2}
	if _, err = Coarsen(n, 3); err != nil {
		t.Fatal(err)
	}
	if n.outputs[0] == before[0] {
		t.Fatal("source output slice mutated")
	}
}

type siblingTestNetwork struct {
	testNetwork
	sibs map[int32]int32
}

func (n *siblingTestNetwork) Sibling(id int32) int32 { return n.sibs[id] }
func TestCoarsenChoiceBarrierAndRetainedBuffers(t *testing.T) {
	n := &siblingTestNetwork{testNetwork: testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {kind: 1}, {kind: 1}}, inputs: []int32{2, 4, 6}}, sibs: map[int32]int32{}}
	alt := coarseTestAnd(&n.testNetwork, 2, 4)
	root := coarseTestMux(&n.testNetwork, 2, 4, 6)
	n.sibs[5] = alt >> 1
	n.outputs = []int32{root}
	out, err := Coarsen(n, 3)
	if err != nil {
		t.Fatal(err)
	}
	_, _, muxes := coarseTestCounts(out)
	if muxes != 0 {
		t.Fatal("choice-bearing child absorbed")
	}
	coarseTestEquivalent(t, n, out)
	n2 := &testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {kind: 1}, {a: 2, b: 4, kind: 2}, {a: 6, kind: 5}}, inputs: []int32{2, 4}, outputs: []int32{2}}
	out, err = Coarsen(n2, 3)
	if err != nil {
		t.Fatal(err)
	}
	ands, _, _ := coarseTestCounts(out)
	if ands != 1 || out.ObjectCount() != 6 {
		t.Fatal("cleanup dropped dangling buffer cone", out.ObjectCount(), ands)
	}
}
func TestCoarsenMalformed(t *testing.T) {
	for _, n := range []*testNetwork{{gates: []testGate{{kind: 1}}}, {gates: []testGate{{kind: 0}, {a: 2, b: 0, kind: 2}}}, {gates: []testGate{{kind: 0}, {kind: 1}}}, {gates: []testGate{{kind: 0}, {kind: 3}}}} {
		if _, err := Coarsen(n, 3); err == nil {
			t.Fatal("malformed graph accepted")
		}
	}
	out, _ := Coarsen(&testNetwork{gates: []testGate{{kind: 0}}, outputs: []int32{0}}, 3)
	if _, err := Coarsen(out, 3); err == nil {
		t.Fatal("real mux storage accepted")
	}
}
func TestCoarsenOriginalABCStatistics(t *testing.T) {
	abc := filepath.Clean("../../../../abc")
	if _, err := os.Stat(abc); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	r := rand.New(rand.NewSource(991))
	stats := regexp.MustCompile(`xor\s*=\s*(\d+).*mux\s*=\s*(\d+).*and\s*=\s*(\d+)`)
	for trial := 0; trial < 12; trial++ {
		n := &testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {kind: 1}, {kind: 1}, {kind: 1}}, inputs: []int32{2, 4, 6, 8}}
		for k := 0; k < 8; k++ {
			available := len(n.gates)
			pick := func() int32 { return int32(1+r.Intn(available-1))*2 + int32(r.Intn(2)) }
			c, a, b := pick(), pick(), pick()
			for a>>1 == c>>1 {
				a = pick()
			}
			for b>>1 == a>>1 || b>>1 == c>>1 {
				b = pick()
			}
			if k%2 == 0 {
				coarseTestMux(n, c, a, b)
			} else {
				coarseTestAnd(n, a, b)
			}
		}
		last := int32(len(n.gates)-1) * 2
		n.outputs = []int32{last, last ^ 1}
		model := &ioa.Model{MaxVar: uint32(len(n.gates) - 1), Inputs: []uint32{2, 4, 6, 8}, Outputs: []uint32{uint32(last), uint32(last ^ 1)}}
		for id, g := range n.gates {
			if g.kind == 2 {
				model.ANDs = append(model.ANDs, ioa.And{Literal: uint32(id * 2), Left: uint32(g.b), Right: uint32(g.a)})
			}
		}
		data, err := ioa.Marshal(model, ioa.WriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "input.aig")
		os.WriteFile(path, data, 0644)
		out, err := Coarsen(n, 3)
		if err != nil {
			t.Fatal(err)
		}
		coarseTestEquivalent(t, n, out)
		cmd := exec.Command(abc, "-c", fmt.Sprintf("&r -s %s; &st -m -L 3; &ps -m", path))
		raw, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("ABC oracle failed %v\n%s", err, raw)
		}
		match := stats.FindStringSubmatch(string(raw))
		if match == nil {
			t.Fatalf("no stats\n%s", raw)
		}
		a, x, m := coarseTestCounts(out)
		actual := []int{x, m, a}
		for i := range actual {
			want, _ := strconv.Atoi(match[i+1])
			if actual[i] != want {
				t.Fatalf("trial %d gate counts Go=%v C=%v\n%s", trial, actual, match[1:], raw)
			}
		}
	}
}
func TestCoarsenRejectsMisorderedInterfaces(t *testing.T) {
	for _, n := range []*testNetwork{
		{gates: []testGate{{kind: 0}, {kind: 1}, {kind: 1}}, inputs: []int32{4, 2}, outputs: []int32{2}},
		{gates: []testGate{{kind: 0}, {kind: 1}, {a: 2, kind: 6}}, inputs: []int32{2}, outputs: []int32{3}},
	} {
		if _, err := Coarsen(n, 3); err == nil {
			t.Error("noncanonical interface silently changed")
		}
	}
}
func TestCoarsenHashRealExhaustive(t *testing.T) {
	n := newCoarsened()
	for i := 0; i < 3; i++ {
		n.append(coarseGate{kind: 1})
	}
	h := coarseHash{n: n, table: make(map[coarseGate]int32)}
	for a := int32(0); a < 8; a++ {
		for b := int32(0); b < 8; b++ {
			x := h.xor(a, b)
			if x != h.xor(b, a) {
				t.Fatal("noncanonical XOR")
			}
			for c := int32(0); c < 8; c++ {
				result := h.mux(c, a, b)
				if result != h.mux(c^1, b, a) {
					t.Fatal("noncanonical MUX control")
				}
				n.outputs = []int32{result}
				for mask := 0; mask < 8; mask++ {
					input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
					v := func(lit int32) bool {
						if lit < 2 {
							return lit == 1
						}
						return input[(lit>>1)-1] != (lit&1 != 0)
					}
					want := v(b)
					if v(c) {
						want = v(a)
					}
					if got := coarseTestEval(n, input)[0]; got != want {
						t.Fatal("hash MUX changed function", a, b, c, mask)
					}
				}
			}
		}
	}
}
func TestCoarsenDegenerateSimpleAndPreservesFunction(t *testing.T) {
	for _, b := range []int32{2, 3} {
		n := &testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {a: 2, b: b, kind: 2}}, inputs: []int32{2}, outputs: []int32{4}}
		out, err := Coarsen(n, 3)
		if err != nil {
			t.Fatal(err)
		}
		coarseTestEquivalent(t, n, out)
		if b == 2 {
			_, _, _, kind := out.Gate(2)
			if kind != 5 {
				t.Fatal("equal literals did not retain source buffer")
			}
		} else {
			if out.OutputLiterals()[0] != 0 {
				t.Fatal("complementary literals did not become false")
			}
		}
	}
}
