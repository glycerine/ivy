package ssc

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/satoko"
)

// Exercise the public transformations on actual native XOR, MUX and buffer
// objects. Most random sweeping fixtures build those operations out of ANDs.
func TestRandomNativeGateTransformations(t *testing.T) {
	rng := rand.New(rand.NewSource(59201))
	for trial := 0; trial < 50; trial++ {
		g := gia.Gia_ManStart(100)
		g.EnableMuxes()
		var nodes []int32
		for i := 0; i < 4; i++ {
			nodes = append(nodes, gia.Gia_ManAppendCi(g))
		}
		for i := 0; i < 20; i++ {
			// Pick distinct objects so native gates satisfy their preconditions.
			p := rng.Perm(len(nodes))
			a, b, c := nodes[p[0]]^int32(rng.Intn(2)), nodes[p[1]]^int32(rng.Intn(2)), nodes[p[2]]^int32(rng.Intn(2))
			var n int32
			switch i % 4 {
			case 0:
				n = gia.Gia_ManAppendAnd(g, a, b)
			case 1:
				n = gia.Gia_ManAppendXorReal(g, a, b)
			case 2:
				n = gia.Gia_ManAppendMuxReal(g, a, b, c)
			case 3:
				n = gia.Gia_ManAppendBuf(g, a)
			}
			nodes = append(nodes, n)
		}
		for i := 0; i < 4; i++ {
			gia.Gia_ManAppendCo(g, nodes[4+rng.Intn(len(nodes)-4)]^int32(rng.Intn(2)))
		}
		care := gia.Gia_ManStart(10)
		care.EnableMuxes()
		var inputs []int32
		for i := 0; i < 4; i++ {
			inputs = append(inputs, gia.Gia_ManAppendCi(care))
		}
		violation := gia.Gia_ManAppendMuxReal(care, inputs[0], inputs[1], inputs[2]^1)
		gia.Gia_ManAppendCo(care, violation)
		params := DefaultParams()
		params.Verify = true
		params.SatVarMax = 1
		params.CallsRecycle = 0
		result := Ssc_PerformSweeping(g, care, &params)
		if hasNativeGates(result) {
			t.Fatal("sweeping must return an AND-inverter graph")
		}
		checkCare(t, g, result, care)

		for _, transformed := range []*gia.Gia_Man_t{Gia_ManDropContained(g), Gia_ManOptimizeRing(g)} {
			for mask := 0; mask < 16; mask++ {
				in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				before, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				after, err := transformed.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				union := func(values []bool) bool {
					for _, v := range values {
						if v {
							return true
						}
					}
					return false
				}
				if union(before) != union(after) {
					t.Fatalf("trial %d mask %d: output union changed", trial, mask)
				}
			}
		}
	}
}

func TestDropContainedSequentialCones(t *testing.T) {
	g := gia.Gia_ManStart(32)
	g.EnableMuxes()
	x := gia.Gia_ManAppendCi(g)
	q0 := gia.Gia_ManAppendCi(g)
	q1 := gia.Gia_ManAppendCi(g)
	_ = gia.Gia_ManAppendCi(g) // unused register
	contained := gia.Gia_ManAppendAnd(g, x, q0)
	next0 := gia.Gia_ManAppendXorReal(g, x, q1)
	next1 := gia.Gia_ManAppendMuxReal(g, x, q0, q1)
	gia.Gia_ManAppendCo(g, contained)
	gia.Gia_ManAppendCo(g, q0)
	gia.Gia_ManAppendCo(g, next0)
	gia.Gia_ManAppendCo(g, next1)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 3)

	result := Gia_ManDropContained(g)
	if result.RegisterCount() != 2 || len(result.InputLiterals()) != 3 || len(result.OutputLiterals()) != 3 {
		t.Fatalf("expected one PO and two transitively reachable registers, got %d inputs, %d outputs, %d registers", len(result.InputLiterals()), len(result.OutputLiterals()), result.RegisterCount())
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, err := result.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		want, err := g.Evaluate(append(append([]bool(nil), in...), false))
		if err != nil {
			t.Fatal(err)
		}
		for i := range got {
			if got[i] != want[i+1] {
				t.Fatalf("mask %d output %d: got %v want %v", mask, i, got[i], want[i+1])
			}
		}
	}
}

func TestDropContainedKeepsRegistersWhenNoPrimaryOutputDrops(t *testing.T) {
	g := gia.Gia_ManStart(10)
	x := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, x^1)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	result := Gia_ManDropContained(g)
	if result.RegisterCount() != 1 || len(result.OutputLiterals()) != 3 {
		t.Fatalf("all distinct POs and the register must survive: %d outputs, %d registers", len(result.OutputLiterals()), result.RegisterCount())
	}
}

// Ssc_GiaRandomPiPattern in sscSim.c iterates PIs rather than all CIs.
func TestRandomPatternLeavesRegisterOutputsZero(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 1)
	sim := NewSimulation(g, 2)
	Ssc_GiaRandomPiPattern(sim, 2, []bool{true, true})
	if sim.Inputs[0][0]&1 == 0 {
		t.Fatal("primary input did not receive pivot bit")
	}
	for _, word := range sim.Inputs[1] {
		if word != 0 {
			t.Fatal("register output was randomized")
		}
	}
}

func TestRandomPatternUsesGiaRandomWords(t *testing.T) {
	g := gia.Gia_ManStart(3)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManRandom(1)
	high, low := gia.Gia_ManRandom(0), gia.Gia_ManRandom(0)
	want := uint64(high)<<32 | uint64(low)<<2 | 2
	gia.Gia_ManRandom(1)
	sim := NewSimulation(g, 1)
	Ssc_GiaRandomPiPattern(sim, 1, nil)
	if sim.Inputs[0][0] != want {
		t.Fatalf("got %x want %x", sim.Inputs[0][0], want)
	}
}

func TestClassWordRefinementNormalizesCurrentFirstBit(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCi(g)
	s := NewSimulation(g, 1)
	Ssc_GiaResetSimInfo(s)
	s.Values[1][0] = ^uint64(0)
	c := &Classes{Groups: [][]int32{{0, 1}}, Repr: []int32{-1, 0}, Phase: []bool{false, false}}
	if n := Ssc_GiaClassesRefine(c, s); n != 0 || c.Repr[1] != 0 {
		t.Fatal("complementary simulation words should remain in the same class")
	}
}

// For two departed constant candidates, the C hash table has 101 buckets;
// 2*1291 and 204*1291 collide but the signatures must still be split.
func TestClassRefinementSplitsConstantHashCollision(t *testing.T) {
	g := gia.Gia_ManStart(5)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	s := NewSimulation(g, 1)
	Ssc_GiaResetSimInfo(s)
	s.Values[1][0], s.Values[2][0] = 2, 204
	c := &Classes{Groups: [][]int32{{0, 1, 2}}, Repr: []int32{-1, 0, 0}, Phase: make([]bool, 3)}
	if count := Ssc_GiaClassesRefine(c, s); count != 2 || c.Repr[1] != -1 || c.Repr[2] != -1 {
		t.Fatalf("collision must split, counting two departures: count=%d repr=%v", count, c.Repr)
	}
}

func TestVerificationChecksDifferenceUnderCare(t *testing.T) {
	a := newBuilder(1, 8)
	b := newBuilder(1, 8)
	care := newBuilder(1, 8)
	gia.Gia_ManAppendCo(a.graph, a.graph.InputLiterals()[0])
	gia.Gia_ManAppendCo(b.graph, 0)
	gia.Gia_ManAppendCo(care.graph, 0)
	if got := Ssc_PerformVerification(a.graph, b.graph, care.graph); got != 0 {
		t.Fatalf("unconstrained difference: got %d", got)
	}
	masked := newBuilder(1, 8)
	gia.Gia_ManAppendCo(masked.graph, masked.graph.InputLiterals()[0])
	if got := Ssc_PerformVerification(a.graph, b.graph, masked.graph); got != 1 {
		t.Fatalf("difference excluded by care: got %d", got)
	}
}

func TestStartSetsCarePhasesFromPivot(t *testing.T) {
	subject := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCi(subject)
	care := gia.Gia_ManStart(4)
	a := gia.Gia_ManAppendCi(care)
	gia.Gia_ManAppendCo(care, a^1)
	manager, status := Ssc_ManStart(subject, care, nil)
	if manager == nil || status != 1 {
		t.Fatalf("expected satisfiable pivot, got %d", status)
	}
	if got := gia.Gia_ObjPhase(gia.Gia_ManCi(care, 0)); got != 1 {
		t.Fatalf("care input phase = %d, want 1", got)
	}
	if got := gia.Gia_ManCheckCoPhase(care); got != 0 {
		t.Fatalf("care violation phase = %d, want 0", got)
	}
}

func TestStartCopiesAndInvertsSourceCareGraph(t *testing.T) {
	subject := newBuilder(2, 10)
	care := gia.Gia_ManStart(10)
	care.SetIdentity("care-name", "care-spec")
	a, b := gia.Gia_ManAppendCi(care), gia.Gia_ManAppendCi(care)
	lit := gia.Gia_ManAppendAnd(care, a, b)
	gia.Gia_ManAppendCo(care, lit)
	expected := gia.Gia_ManDupDfs(care)
	gia.Gia_ManInvertPos(expected)
	p, status := Ssc_ManStart(subject.graph, care, nil)
	if status != satoko.SATOKO_SAT || p == nil {
		t.Fatalf("start status %d", status)
	}
	if p.Aig.NetworkName() != expected.NetworkName() || p.Aig.NetworkSpec() != expected.NetworkSpec() || p.Aig.ObjectCount() != expected.ObjectCount() || !reflect.DeepEqual(p.Aig.OutputLiterals(), expected.OutputLiterals()) {
		t.Fatal("working FRAIG did not preserve the DFS-copied and inverted care graph")
	}
}

func TestSweepRejectsDuplicateCareNodesAtSourceHashStart(t *testing.T) {
	subject := newBuilder(2, 8)
	gia.Gia_ManAppendCo(subject.graph, subject.graph.InputLiterals()[0])
	care := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(care), gia.Gia_ManAppendCi(care)
	first := gia.Gia_ManAppendAnd(care, a, b)
	second := gia.Gia_ManAppendAnd(care, a, b)
	gia.Gia_ManAppendCo(care, first)
	gia.Gia_ManAppendCo(care, second)
	defer func() {
		if recover() == nil {
			t.Fatal("C Gia_ManHashStart asserts on duplicate care nodes")
		}
	}()
	Ssc_PerformSweeping(subject.graph, care, nil)
}

func TestFailedOptionalVerificationStillReturnsResult(t *testing.T) {
	subject := newBuilder(1, 4)
	gia.Gia_ManAppendCo(subject.graph, subject.graph.InputLiterals()[0])
	result := newBuilder(1, 4)
	gia.Gia_ManAppendCo(result.graph, 0)
	care := emptyCare(subject.graph)
	params := DefaultParams()
	params.Verify = true
	if got := finishSweep(subject.graph, result.graph, care, &params); got != result.graph {
		t.Fatal("failed verification must return the original result")
	}
}

func TestMappedCandidateSourceOrderAndPhase(t *testing.T) {
	p := &Ssc_Man_t{}
	if status, equal := checkMappedCandidate(p, 2, 3, 1); !equal || status != satoko.SATOKO_UNSAT || p.Stats.Calls != 0 {
		t.Fatal("matching mapped literals must use the no-SAT fast path")
	}
	if status, equal := checkMappedCandidate(p, 6, 4, 0); equal || status != satoko.SATOKO_UNDEC || p.Stats.Calls != 1 || p.Stats.Undecided != 1 {
		t.Fatalf("reversed FRAIG IDs must count as undecided: status %d, stats %+v", status, p.Stats)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("same FRAIG node with opposite phase must assert")
		}
		if p.Stats.Calls != 1 {
			t.Fatal("assertion must precede SAT call accounting")
		}
	}()
	checkMappedCandidate(p, 2, 3, 0)
}

func TestInitialSimulationDoesNotIncrementSourceCounters(t *testing.T) {
	subject, care := fixture()
	p, status := Ssc_ManStart(subject, care, nil)
	if status != satoko.SATOKO_SAT {
		t.Fatalf("pivot status %d", status)
	}
	carePatterns(p)
	if p.Stats != (Statistics{}) {
		t.Fatalf("initial simulation changed C's untouched counters: %+v", p.Stats)
	}
}

func TestStartAcceptsInactiveRecycleOptions(t *testing.T) {
	subject, care := fixture()
	params := DefaultParams()
	params.SatVarMax, params.CallsRecycle = -1, -3
	p, status := Ssc_ManStart(subject, care, &params)
	if status != satoko.SATOKO_SAT || p == nil {
		t.Fatalf("inactive C options changed start status: %d", status)
	}
	if p.Params.SatVarMax != -1 || p.Params.CallsRecycle != -3 {
		t.Fatal("inactive options were not retained")
	}
}

// The original sat_solver_solve accepts signed limits. A compiled C SSC
// oracle returns unknown for this satisfiable care when nBTLimit is -1.
func TestNegativeConflictLimitReturnsSourceUnknown(t *testing.T) {
	subject := newBuilder(1, 5)
	gia.Gia_ManAppendCo(subject.graph, subject.graph.InputLiterals()[0])
	care := newBuilder(1, 5)
	gia.Gia_ManAppendCo(care.graph, 0)
	params := DefaultParams()
	params.BTLimit = -1
	p, status := Ssc_ManStart(subject.graph, care.graph, &params)
	if p != nil || status != satoko.SATOKO_UNDEC {
		t.Fatalf("manager=%v status=%d; want source unknown", p, status)
	}
	result := Ssc_PerformSweeping(subject.graph, care.graph, &params)
	for _, value := range []bool{false, true} {
		out, err := result.Evaluate([]bool{value})
		if err != nil || len(out) != 1 || out[0] != value {
			t.Fatalf("input=%v output=%v err=%v", value, out, err)
		}
	}
}

func TestNegativeConflictLimitStillDetectsRootUnsatCare(t *testing.T) {
	subject := newBuilder(1, 5)
	care := newBuilder(1, 5)
	gia.Gia_ManAppendCo(care.graph, 1)
	params := DefaultParams()
	params.BTLimit = -1
	p, status := Ssc_ManStart(subject.graph, care.graph, &params)
	if p != nil || status != satoko.SATOKO_UNSAT {
		t.Fatalf("manager=%v status=%d; want UNSAT", p, status)
	}
}

func TestStartUnsatisfiableCareDoesNotReturnManager(t *testing.T) {
	subject := newBuilder(1, 5)
	care := newBuilder(1, 5)
	gia.Gia_ManAppendCo(care.graph, 1)
	p, status := Ssc_ManStart(subject.graph, care.graph, nil)
	if p != nil || status != satoko.SATOKO_UNSAT {
		t.Fatalf("got manager %v, status %d; want no manager and UNSAT", p, status)
	}
}

func TestPivotBudgetExhaustionLeavesSubjectUnchanged(t *testing.T) {
	// Four pigeons in three holes requires search beyond one conflict. Care
	// violation is the negation of the pigeonhole formula.
	care := newBuilder(12, 200)
	in := care.graph.InputLiterals()
	formula := int32(1)
	for pigeon := 0; pigeon < 4; pigeon++ {
		row := int32(0)
		for hole := 0; hole < 3; hole++ {
			row = care.or(row, in[pigeon*3+hole])
		}
		formula = care.and(formula, row)
	}
	for hole := 0; hole < 3; hole++ {
		for a := 0; a < 4; a++ {
			for b := a + 1; b < 4; b++ {
				formula = care.and(formula, care.or(in[a*3+hole]^1, in[b*3+hole]^1))
			}
		}
	}
	gia.Gia_ManAppendCo(care.graph, formula^1)
	subject := newBuilder(12, 30)
	subject.and(subject.graph.InputLiterals()[1], subject.graph.InputLiterals()[2]) // dangling node survives C Gia_ManDup
	gia.Gia_ManAppendCo(subject.graph, subject.graph.InputLiterals()[0])
	params := DefaultParams()
	params.BTLimit = 1
	p, status := Ssc_ManStart(subject.graph, care.graph, &params)
	if status != satoko.SATOKO_UNDEC || p != nil {
		t.Fatalf("one-conflict pivot: status %d manager %v", status, p)
	}
	result := Ssc_PerformSweeping(subject.graph, care.graph, &params)
	if result.ObjectCount() != subject.graph.ObjectCount() {
		t.Fatal("inconclusive pivot must duplicate even dangling objects")
	}
	if subject.graph.HasRepresentatives() {
		t.Fatal("inconclusive pivot must not initialize classes")
	}
	for _, value := range []bool{false, true} {
		inputs := make([]bool, 12)
		inputs[0] = value
		got, err := result.Evaluate(inputs)
		if err != nil || len(got) != 1 || got[0] != value {
			t.Fatalf("fallback output %v error %v", got, err)
		}
	}
}

func TestConstraintWrapperUsesLevelizedSubject(t *testing.T) {
	build := func() *gia.Gia_Man_t {
		g := gia.Gia_ManStart(20)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		deep := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, x, y), z)
		shallow := gia.Gia_ManAppendAnd(g, x, z)
		gia.Gia_ManAppendCo(g, deep)
		gia.Gia_ManAppendCo(g, shallow)
		return g
	}
	// These POs force the earlier-created depth-two node to precede the
	// later depth-one node before the C wrapper's levelization step.
	g := build()
	params := DefaultParams()
	got := Ssc_PerformSweepingConstr(g, &params)
	reference := gia.Gia_ManDupLevelized(build())
	want := Ssc_PerformSweeping(reference, emptyCare(reference), &params)
	if got.ObjectCount() != want.ObjectCount() || !reflect.DeepEqual(got.OutputLiterals(), want.OutputLiterals()) {
		t.Fatal("wrapper differs from source levelize-then-sweep sequence")
	}
	for id := int32(0); int(id) < got.ObjectCount(); id++ {
		a, b, c, kind := got.Gate(id)
		x, y, z, otherKind := want.Gate(id)
		if a != x || b != y || c != z || kind != otherKind {
			t.Fatalf("object %d differs from levelized source sequence", id)
		}
	}
}

func TestRandomSequentialConstraintSweeps(t *testing.T) {
	rng := rand.New(rand.NewSource(38122))
	for trial := 0; trial < 80; trial++ {
		g := gia.Gia_ManStart(50)
		g.EnableMuxes()
		var nodes []int32
		for i := 0; i < 4; i++ {
			nodes = append(nodes, gia.Gia_ManAppendCi(g))
		}
		for i := 0; i < 12; i++ {
			p := rng.Perm(len(nodes))
			a := nodes[p[0]] ^ int32(rng.Intn(2))
			b := nodes[p[1]] ^ int32(rng.Intn(2))
			c := nodes[p[2]] ^ int32(rng.Intn(2))
			switch i % 3 {
			case 0:
				nodes = append(nodes, gia.Gia_ManAppendAnd(g, a, b))
			case 1:
				nodes = append(nodes, gia.Gia_ManAppendXorReal(g, a, b))
			case 2:
				nodes = append(nodes, gia.Gia_ManAppendMuxReal(g, a, b, c))
			}
		}
		constraints := trial % 3
		for i := 0; i < 4+constraints; i++ {
			gia.Gia_ManAppendCo(g, nodes[rng.Intn(len(nodes))]^int32(rng.Intn(2)))
		}
		gia.Gia_ManSetRegNum(g, 2)
		gia.Gia_ManSetConstrNum(g, int32(constraints))
		for _, appendCare := range []bool{false, true} {
			params := DefaultParams()
			params.Append = appendCare
			params.Verify = true
			result := Ssc_PerformSweepingConstr(g, &params)
			nCare := 0
			if appendCare {
				nCare = constraints
				if nCare == 0 {
					nCare = 1
				}
			}
			if result.RegisterCount() != 2 || int(gia.Gia_ManConstrNum(result)) != nCare || len(result.OutputLiterals()) != 4+nCare {
				t.Fatalf("trial %d append %v: unexpected interface", trial, appendCare)
			}
			for mask := 0; mask < 16; mask++ {
				in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				before, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				after, err := result.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				allowed := true
				for i := 0; i < constraints; i++ {
					if before[2+i] {
						allowed = false
					}
					if appendCare && after[2+i] != before[2+i] {
						t.Fatalf("trial %d mask %d constraint changed", trial, mask)
					}
				}
				if !allowed {
					continue
				}
				if before[0] != after[0] || before[1] != after[1] || before[2+constraints] != after[2+nCare] || before[3+constraints] != after[3+nCare] {
					t.Fatalf("trial %d append %v mask %d: PO or next state changed", trial, appendCare, mask)
				}
			}
		}
	}
}
