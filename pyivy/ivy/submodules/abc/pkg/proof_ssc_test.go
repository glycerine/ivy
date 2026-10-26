package abc

// Go-only validation for src/proof/ssc.

import (
	"errors"
	"testing"
)

func sscTestCare(t *testing.T, inputs int, violation func([]GiaLit) GiaLit) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	cis := make([]GiaLit, inputs)
	var err error
	for i := range cis {
		cis[i], err = g.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = g.AppendCO(violation(cis)); err != nil {
		t.Fatal(err)
	}
	return g
}

func sscSequentialFixture(t *testing.T) (*GiaManager, *GiaManager, GiaLit, GiaLit) {
	t.Helper()
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	state, _ := g.AppendCI()
	first, _ := g.AppendAnd(a, b)
	duplicate, _ := g.AppendAnd(a, b)
	out, _ := g.AppendMuxReal(state, duplicate, a)
	g.AppendCO(out)
	g.AppendCO(first)
	g.SetRegisterCount(1)
	care := NewGiaManager()
	ca, _ := care.AppendCI()
	cb, _ := care.AppendCI()
	care.AppendCI()
	bad, _ := care.AppendAnd(ca, cb.Not())
	care.AppendCO(bad)
	return g, care, first, duplicate
}

func TestSSCClassesSweepSequentialAndWitness(t *testing.T) {
	source, care, first, duplicate := sscSequentialFixture(t)
	classes, pivot, err := SSCClasses(source, care, SSCParameters{})
	if err != nil || pivot == nil || len(classes) == 0 {
		t.Fatal(classes, pivot, err)
	}
	proved, err := SSCCheckEquivalence(source, care, first, duplicate, SSCParameters{})
	if err != nil || proved.Status != SatUnsatisfiable || proved.Witness != nil {
		t.Fatal(proved, err)
	}
	disproved, err := SSCCheckEquivalence(source, care, first, GiaLit(source.cis[1])<<1, SSCParameters{})
	if err != nil || disproved.Status != SatSatisfiable || disproved.Witness == nil {
		t.Fatal(disproved, err)
	}
	inputs := disproved.Witness.Inputs()
	holds, _ := sscCareHolds(care, inputs)
	values, _ := source.evaluateValues(inputs)
	if !holds || values[first.ID()] == values[source.cis[1]] {
		t.Fatal("witness does not replay", inputs)
	}
	result, err := SSCSweep(source, care, SSCParameters{Verify: true})
	if err != nil || result.Manager == nil || result.Unsatisfiable {
		t.Fatal(result, err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		in := sscBits(assignment, 3)
		holds, _ := sscCareHolds(care, in)
		if !holds {
			continue
		}
		before, beforeNext, _ := source.Step(in[:2], in[2:])
		after, afterNext, _ := result.Manager.Step(in[:2], in[2:])
		if !equalBools(before, after) || !equalBools(beforeNext, afterNext) {
			t.Fatalf("%v: (%v,%v) != (%v,%v)", in, before, beforeNext, after, afterNext)
		}
	}
	broken := result.Manager.Clone()
	if err = broken.FlipCODriver(0); err != nil {
		t.Fatal(err)
	}
	verification, err := SSCVerify(source, broken, care, SSCParameters{})
	if err != nil || verification.Status != SatSatisfiable || verification.Witness == nil {
		t.Fatal(verification, err)
	}
	witnessInputs := verification.Witness.Inputs()
	want, _ := source.Evaluate(witnessInputs)
	got, _ := broken.Evaluate(witnessInputs)
	if equalBools(want, got) {
		t.Fatal("verification witness does not replay")
	}
}

func TestSSCCareConventionPivotClassesAndComplements(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	zero, _ := g.AppendAnd(a, b)
	one, _ := g.AppendBuffer(zero.Not())
	g.AppendCO(zero)
	care := NewGiaManager()
	ca, _ := care.AppendCI()
	cb, _ := care.AppendCI()
	care.AppendCO(ca)
	care.AppendCO(cb.Not())
	for assignment := 0; assignment < 4; assignment++ {
		holds, err := sscCareHolds(care, sscBits(assignment, 2))
		if err != nil || holds != (assignment == 2) {
			t.Fatalf("assignment=%d holds=%v err=%v", assignment, holds, err)
		}
	}
	classes, pivot, err := SSCClasses(g, care, SSCParameters{})
	if err != nil || pivot == nil || !equalBools(pivot.Inputs(), []bool{false, true}) {
		t.Fatal(classes, pivot, err)
	}
	foundZero, foundOne := false, false
	for _, class := range classes {
		if class.Representative != 0 {
			continue
		}
		for _, member := range class.Members {
			if member.ID == zero.ID() && !member.Complemented {
				foundZero = true
			}
			if member.ID == one.ID() && member.Complemented {
				foundOne = true
			}
		}
	}
	if !foundZero || !foundOne {
		t.Fatalf("classes=%+v", classes)
	}
}

func equalBools(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sscPhase(value bool, complemented bool) bool {
	if complemented {
		return !value
	}
	return value
}

func TestSSCUnsatisfiableCareConstraintsUtilitiesAndLimits(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	g.AppendCO(a)
	care := sscTestCare(t, 1, func([]GiaLit) GiaLit { return GiaConst1 })
	result, err := SSCSweep(g, care, SSCParameters{})
	if err != nil || !result.Unsatisfiable {
		t.Fatal(result, err)
	}
	for _, in := range [][]bool{{false}, {true}} {
		out, _ := result.Manager.Evaluate(in)
		if out[0] {
			t.Fatal("nonzero UNSAT-care result")
		}
	}
	if _, err = SSCSweep(g, care, SSCParameters{MaxAssignments: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = SSCSweep(g, care, SSCParameters{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	tooLarge := NewGiaManager()
	tx, _ := tooLarge.AppendCI()
	ty, _ := tooLarge.AppendBuffer(tx)
	tooLarge.AppendCO(ty)
	if _, err = SSCVerify(g, tooLarge, sscTestCare(t, 1, func([]GiaLit) GiaLit { return GiaConst0 }), SSCParameters{MaxObjects: g.ObjectCount()}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = SSCDropContained(tooLarge, SSCParameters{MaxObjects: g.ObjectCount()}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	onehot, err := SSCGenerateOneHot(3, SSCParameters{})
	if err != nil || onehot.InputCount() != 3 {
		t.Fatal(onehot, err)
	}

	constrained := NewGiaManager()
	x, _ := constrained.AppendCI()
	constrained.AppendCO(x)
	constrained.AppendCO(x.Not())
	constrained.SetConstraintCount(1)
	cr, err := SSCSweepConstraints(constrained, SSCParameters{AppendConstraints: true})
	if err != nil || cr.Manager.ConstraintCount() != 1 || cr.Manager.OutputCount() != 2 {
		t.Fatal(cr, err)
	}
}

func TestSSCDropContained(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	ab, _ := g.AppendAnd(a, b)
	g.AppendCO(a)
	g.AppendCO(ab)
	g.AppendCO(b)
	dropped, err := SSCDropContained(g, SSCParameters{})
	if err != nil {
		t.Fatal(err)
	}
	// a&b can never be the sole true output because both a and b are true.
	if dropped.OutputCount() != 2 {
		t.Fatal(dropped.OutputCount())
	}
	optimized, err := SSCOptimizeRing(dropped, SSCParameters{Verify: true})
	if err != nil || optimized.OutputCount() != dropped.OutputCount() {
		t.Fatal(optimized, err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := sscBits(assignment, 2)
		want, _ := dropped.Evaluate(inputs)
		got, _ := optimized.Evaluate(inputs)
		if !equalBools(want, got) {
			t.Fatal(inputs, want, got)
		}
	}
}

func TestSSCIndependentBoundedSequentialModelAndWitnessReplay(t *testing.T) {
	state := uint64(0x243f6a8885a308d3)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		g := NewGiaManager()
		a, _ := g.AppendCI()
		b, _ := g.AppendCI()
		q, _ := g.AppendCI()
		x, _ := g.AppendAndSimplified(a.NotIf(state&1 != 0), b.NotIf(state&2 != 0))
		y, _ := g.AppendAndSimplified(a.NotIf(state&4 != 0), b.NotIf(state&8 != 0))
		z, _ := g.AppendMuxSimplified(q, x, y)
		g.AppendCO(z)
		g.AppendCO(x)
		g.SetRegisterCount(1)
		care := NewGiaManager()
		ca, _ := care.AppendCI()
		cb, _ := care.AppendCI()
		care.AppendCI()
		bad, _ := care.AppendAndSimplified(ca.NotIf(state&16 != 0), cb.NotIf(state&32 != 0))
		care.AppendCO(bad)
		result, err := SSCSweep(g, care, SSCParameters{MaxAssignments: 8, MaxObjects: 64})
		if err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputs := sscBits(assignment, 3)
			aValue, bValue, qValue := inputs[0], inputs[1], inputs[2]
			bad := sscPhase(aValue, state&16 != 0) && sscPhase(bValue, state&32 != 0)
			holds := !bad
			if !holds {
				continue
			}
			xValue := sscPhase(aValue, state&1 != 0) && sscPhase(bValue, state&2 != 0)
			yValue := sscPhase(aValue, state&4 != 0) && sscPhase(bValue, state&8 != 0)
			zValue := yValue
			if qValue {
				zValue = xValue
			}
			got, gotNext, _ := result.Manager.Step(inputs[:2], inputs[2:])
			if !equalBools([]bool{zValue}, got) || !equalBools([]bool{xValue}, gotNext) {
				t.Fatalf("execution %d assignment %d", execution, assignment)
			}
		}
		query, err := SSCCheckEquivalence(g, care, x, y, SSCParameters{MaxAssignments: 8, MaxObjects: 64})
		if err != nil {
			t.Fatal(err)
		}
		if query.Status == SatSatisfiable {
			inputs := query.Witness.Inputs()
			aValue, bValue := inputs[0], inputs[1]
			holds := !(sscPhase(aValue, state&16 != 0) && sscPhase(bValue, state&32 != 0))
			xValue := sscPhase(aValue, state&1 != 0) && sscPhase(bValue, state&2 != 0)
			yValue := sscPhase(aValue, state&4 != 0) && sscPhase(bValue, state&8 != 0)
			if !holds || xValue == yValue {
				t.Fatal("invalid replay")
			}
		}
	}
}

func FuzzSSCSweepSequential(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4})
	f.Add([]byte{15, 7, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			return
		}
		var mask byte
		for _, b := range data {
			mask ^= b
		}
		g := NewGiaManager()
		a, _ := g.AppendCI()
		b, _ := g.AppendCI()
		q, _ := g.AppendCI()
		x, _ := g.AppendAndSimplified(a.NotIf(mask&1 != 0), b.NotIf(mask&2 != 0))
		y, _ := g.AppendMuxSimplified(q, x, a.NotIf(mask&4 != 0))
		g.AppendCO(y)
		g.AppendCO(x)
		g.SetRegisterCount(1)
		care := NewGiaManager()
		ca, _ := care.AppendCI()
		cb, _ := care.AppendCI()
		care.AppendCI()
		bad, _ := care.AppendAndSimplified(ca, cb.NotIf(mask&8 != 0))
		care.AppendCO(bad)
		result, err := SSCSweep(g, care, SSCParameters{MaxAssignments: 8, MaxObjects: 64})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			in := sscBits(assignment, 3)
			aValue, bValue, qValue := in[0], in[1], in[2]
			holds := !(aValue && sscPhase(bValue, mask&8 != 0))
			if !holds {
				continue
			}
			xValue := sscPhase(aValue, mask&1 != 0) && sscPhase(bValue, mask&2 != 0)
			fallback := sscPhase(aValue, mask&4 != 0)
			yValue := fallback
			if qValue {
				yValue = xValue
			}
			got, gotNext, _ := result.Manager.Step(in[:2], in[2:])
			if !equalBools([]bool{yValue}, got) || !equalBools([]bool{xValue}, gotNext) {
				t.Fatal(in, yValue, got)
			}
		}
	})
}
