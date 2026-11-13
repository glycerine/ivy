package ssc

import (
	"math/bits"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
)

// A compiled ABC oracle, with Gia_ManRandom(1), returns (0,1) as the first
// satisfying bit for the two-input XOR care graph.
func TestSimulationPivotMatchesSourceRandomSequence(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		b := newBuilder(2, 12)
		inputs := b.graph.InputLiterals()
		output := int32(0)
		if enabled {
			output = b.xor(inputs[0], inputs[1])
		}
		gia.Gia_ManAppendCo(b.graph, output)
		gia.Gia_ManRandom(1)
		pivot := Ssc_GiaFindPivotSim(NewSimulation(b.graph, 1))
		if !enabled {
			if pivot != nil {
				t.Fatalf("empty care produced pivot %v", pivot)
			}
			continue
		}
		if len(pivot) != 2 || pivot[0] || !pivot[1] {
			t.Fatalf("pivot=%v, want [false true]", pivot)
		}
		out, err := b.graph.Evaluate(pivot)
		if err != nil || !out[0] {
			t.Fatalf("pivot violates care: output=%v err=%v", out, err)
		}
	}
}

func TestPackedPrimitiveOperations(t *testing.T) {
	a := []uint64{0x123456789abcdef0, 0xffff0000ffff0000}
	b := []uint64{0x0f0ff0f00f0ff0f0, 0xaaaaaaaaaaaaaaaa}
	for _, ca := range []bool{false, true} {
		for _, cb := range []bool{false, true} {
			out := make([]uint64, 2)
			Ssc_SimAnd(out, a, b, 2, ca, cb)
			for i := range out {
				x, y := a[i], b[i]
				if ca {
					x = ^x
				}
				if cb {
					y = ^y
				}
				if out[i] != x&y {
					t.Fatalf("complement %v/%v word %d", ca, cb, i)
				}
			}
		}
	}
	out := make([]uint64, 2)
	Ssc_SimDup(out, a, 2, true)
	for i := range out {
		if out[i] != ^a[i] {
			t.Fatal("complemented copy")
		}
	}
	Ssc_SimConst(out, 2, false)
	Ssc_SimOr(out, b, 2, true)
	for i := range out {
		if out[i] != ^b[i] {
			t.Fatal("complemented union")
		}
	}
	if got := Ssc_SimCountBits(a, 2); got != int32(bits.OnesCount64(a[0])+bits.OnesCount64(a[1])) {
		t.Fatalf("count %d", got)
	}
	if got := Ssc_SimFindBit([]uint64{0, 0x80}, 2); got != 71 {
		t.Fatalf("first bit %d", got)
	}
	if Ssc_SimFindBit([]uint64{0, 0}, 2) != -1 {
		t.Fatal("empty bit set")
	}
}

func TestCareMaskAndPatternTransfer(t *testing.T) {
	b := newBuilder(2, 12)
	in := b.graph.InputLiterals()
	gia.Gia_ManAppendCo(b.graph, b.and(in[0], in[1]))
	care := NewSimulation(b.graph, 1)
	for _, p := range [][]bool{{false, false}, {true, false}, {false, true}, {true, true}} {
		Ssc_GiaSavePiPattern(care, p)
	}
	Ssc_GiaSimRound(care)
	if mask := Ssc_GiaGetCareMask(care); mask[0] != 8 {
		t.Fatalf("care mask %#x", mask[0])
	}
	if got := Ssc_GiaGetOneSim(care); !reflect.DeepEqual(got, []bool{true, true}) {
		t.Fatalf("care witness %v", got)
	}
	dst := NewSimulation(b.graph, 1)
	if count := Ssc_GiaTransferPiPattern(dst, care, []bool{false, true}); count != 1 {
		t.Fatalf("count %d", count)
	}
	if dst.Inputs[0][0] != 8 || dst.Inputs[1][0] != ^uint64(0) {
		t.Fatalf("transferred %x", dst.Inputs)
	}
}

// C reads the CI simulation objects, which are the snapshot from the last
// simulation round, rather than the separately stored next input patterns.
func TestCareHelpersUseLastSimulationSnapshot(t *testing.T) {
	b := newBuilder(1, 4)
	gia.Gia_ManAppendCo(b.graph, 1)
	care := NewSimulation(b.graph, 1)
	Ssc_GiaSavePiPattern(care, []bool{true})
	Ssc_GiaSimRound(care)
	care.Inputs[0][0] = 0 // change pending PI patterns without simulating again
	if got := Ssc_GiaGetOneSim(care); !reflect.DeepEqual(got, []bool{true}) {
		t.Fatalf("witness must use simulated input, got %v", got)
	}
	dst := NewSimulation(b.graph, 1)
	Ssc_GiaTransferPiPattern(dst, care, []bool{false})
	if dst.Inputs[0][0] != 1 {
		t.Fatalf("transfer must use simulated input, got %#x", dst.Inputs[0][0])
	}
}

func TestCarePatternsMatchesSourceTransferAndRandomOrder(t *testing.T) {
	makeManager := func() *Ssc_Man_t {
		a, c := fixture()
		p, status := Ssc_ManStart(a, c, nil)
		if status != satoko.SATOKO_SAT {
			t.Fatalf("start status %d", status)
		}
		return p
	}
	got, want := makeManager(), makeManager()
	gia.Gia_ManRandom(1)
	carePatterns(got)
	gotNext := gia.Gia_ManRandom(0)
	gia.Gia_ManRandom(1)
	care := NewSimulation(want.Aig, 5)
	candidates := gia.Gia_ManCiNum(want.Original) + gia.Gia_ManAndNum(want.Original)
	for {
		Ssc_GiaRandomPiPattern(care, 5, nil)
		Ssc_GiaSimRound(care)
		if Ssc_GiaTransferPiPattern(want.Sim, care, want.Pivot) == 0 {
			break
		}
		Ssc_GiaSimRound(want.Sim)
		if Ssc_GiaClassesRefine(want.Classes, want.Sim) <= candidates/100 {
			break
		}
	}
	if gotNext != gia.Gia_ManRandom(0) || !reflect.DeepEqual(got.Sim.Inputs, want.Sim.Inputs) || !reflect.DeepEqual(got.Classes, want.Classes) {
		t.Fatal("care refinement changed source transfer or random ordering")
	}
}

func TestCareMaskIgnoresRegisterInputs(t *testing.T) {
	g := gia.Gia_ManStart(6)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 1) // primary output is always true
	gia.Gia_ManAppendCo(g, 0) // register input is always false
	gia.Gia_ManSetRegNum(g, 1)
	s := NewSimulation(g, 1)
	Ssc_GiaSimRound(s)
	if got := Ssc_GiaGetCareMask(s)[0]; got != ^uint64(0) {
		t.Fatalf("care mask included the register input: got %#x", got)
	}
}

func TestClassRefinementAndComplementedSimulation(t *testing.T) {
	c := &Classes{
		Groups: [][]int32{{0, 1, 2, 3}},
		Repr:   make([]int32, 4),
		Phase:  []bool{false, false, true, false},
	}
	c.reindex()
	if Ssc_GiaClassesCheckPairs(c, [][2]int32{{0, 2}}) {
		t.Fatal("unrefined pair reported separated")
	}
	if got := Ssc_GiaSimClassRefineOneBit(c, []bool{false, false, false, true}, 0); got != 1 {
		t.Fatalf("splits %d", got)
	}
	if !reflect.DeepEqual(c.Groups, [][]int32{{0, 1}, {2, 3}}) || !Ssc_GiaClassesCheckPairs(c, [][2]int32{{0, 2}}) {
		t.Fatalf("groups %v", c.Groups)
	}
	sim := &Simulation{Words: 2, Values: [][]uint64{{0, 0}, {0, 0}, {^uint64(0), ^uint64(0)}, {^uint64(0), 0}}}
	if !Ssc_GiaSimAreEqual(sim, 0, 2) || !Ssc_GiaSimIsConst0(sim, 2) {
		t.Fatal("complement equivalence not recognized")
	}
	if Ssc_GiaSimAreEqual(sim, 0, 3) || Ssc_GiaSimIsConst0(sim, 3) {
		t.Fatal("unequal patterns merged")
	}
}

func TestSATEquivalenceAllNodesAndPolarities(t *testing.T) {
	subject, care := fixture()
	manager, status := Ssc_ManStart(subject, care, nil)
	if status != satoko.SATOKO_SAT {
		t.Fatalf("pivot status %d", status)
	}
	mapping := copySubjectIntoFraig(manager)
	sim := NewSimulation(subject, 1)
	for mask := 0; mask < 8; mask++ {
		Ssc_GiaSavePiPattern(sim, []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
	}
	Ssc_GiaSimRound(sim)
	for a := int32(0); int(a) < subject.ObjectCount(); a++ {
		for b := int32(0); int(b) < subject.ObjectCount(); b++ {
			for complement := int32(0); complement < 2; complement++ {
				distinguishable := false
				for mask := 0; mask < 8; mask++ {
					values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
					violations, err := care.Evaluate(values)
					if err != nil {
						t.Fatal(err)
					}
					if violations[0] {
						continue
					}
					x := sim.Values[a][0]&(1<<mask) != 0
					y := sim.Values[b][0]&(1<<mask) != 0
					if x != (y != (complement != 0)) {
						distinguishable = true
						break
					}
				}
				la, lb := mapping[a], mapping[b]
				phase := complement ^ (la & 1) ^ (lb & 1)
				if la>>1 > lb>>1 {
					la, lb = lb, la
				}
				got := satoko.SATOKO_UNSAT
				if la>>1 == lb>>1 {
					if phase != 0 {
						got = satoko.SATOKO_SAT
					}
				} else {
					got = Ssc_ManCheckEquivalence(manager, la>>1, lb>>1, phase)
				}
				want := satoko.SATOKO_UNSAT
				if distinguishable {
					want = satoko.SATOKO_SAT
				}
				if got != want {
					t.Fatalf("nodes %d,%d complement %d: status %d want %d", a, b, complement, got, want)
				}
				if got == satoko.SATOKO_SAT && la>>1 != lb>>1 {
					violations, err := care.Evaluate(manager.Pattern)
					if err != nil || violations[0] {
						t.Fatalf("invalid care witness %v: %v", manager.Pattern, err)
					}
					witness := NewSimulation(subject, 1)
					Ssc_GiaSavePiPattern(witness, manager.Pattern)
					Ssc_GiaSimRound(witness)
					x, y := witness.Values[a][0]&1 != 0, witness.Values[b][0]&1 != 0
					if x == (y != (complement != 0)) {
						t.Fatalf("witness does not distinguish nodes %d,%d", a, b)
					}
				}
			}
		}
	}
}
