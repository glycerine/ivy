package ssc

import (
	"math/bits"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
)

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
